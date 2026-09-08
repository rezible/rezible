package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"

	rez "github.com/rezible/rezible"
)

func runLifecycleServices(ctx context.Context, i do.Injector) error {
	return (&runner{i: i}).run(ctx)
}

type runner struct {
	i do.Injector
}

func (r *runner) run(ctx context.Context) error {
	if bootstrapErr := r.setup(); bootstrapErr != nil {
		return fmt.Errorf("bootstrap: %w", bootstrapErr)
	}

	if seedErr := r.seedDevelopmentIdentity(ctx); seedErr != nil {
		return fmt.Errorf("seed dev identity: %w", seedErr)
	}

	ls, lifecyclesErr := do.Invoke[[]rez.LifecycleService](r.i)
	if lifecyclesErr != nil {
		return fmt.Errorf("get service lifecycles: %w", lifecyclesErr)
	}

	slog.Info("=== Starting Services ===")

	const lifecycleStopTimeout = 10 * time.Second

	group, runCtx := errgroup.WithContext(ctx)
	for _, lp := range ls {
		lifecycle := lp.Lifecycle()
		if lifecycle == nil {
			continue
		}

		name := fmt.Sprintf("%T", lp)
		slog.Info("Starting " + name)

		for _, runFn := range lifecycle.StartFns {
			group.Go(func() error {
				runErr := runFn(runCtx)
				if runErr != nil {
					if ctxErr := runCtx.Err(); ctxErr != nil && errors.Is(runErr, ctxErr) {
						return nil
					}
					return fmt.Errorf("%s: %w", name, runErr)
				}
				if runCtx.Err() == nil {
					return fmt.Errorf("%s stopped unexpectedly", name)
				}
				return nil
			})
		}

		if lifecycle.StopFn != nil {
			group.Go(func() error {
				<-runCtx.Done()
				stopCtx, cancel := context.WithTimeout(context.WithoutCancel(runCtx), lifecycleStopTimeout)
				defer cancel()

				if stopErr := lifecycle.StopFn(stopCtx); stopErr != nil {
					return fmt.Errorf("stop %s: %w", name, stopErr)
				}
				return nil
			})
		}
	}

	servicesErr := group.Wait()
	slog.Info("=== Services Stopped ===")

	if servicesErr != nil {
		return fmt.Errorf("run services: %w", servicesErr)
	}
	return nil
}

func shutdownInjector(ctx context.Context, i do.Injector) error {
	cancelCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	shutdown := i.ShutdownWithContext(cancelCtx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

func (r *runner) setup() error {
	if intgsErr := r.registerIntegrations(); intgsErr != nil {
		return fmt.Errorf("register integrations: %w", intgsErr)
	}

	if jobsErr := r.registerJobWorkers(); jobsErr != nil {
		return fmt.Errorf("register job workers: %w", jobsErr)
	}

	if msgsErr := r.registerMessageHandlers(); msgsErr != nil {
		return fmt.Errorf("register message handlers: %w", msgsErr)
	}

	return nil
}

func (r *runner) registerIntegrations() error {
	intgReg := do.MustInvoke[rez.IntegrationRegistry](r.i)
	eventProcessors := do.MustInvoke[rez.ProviderEventProcessorRegistry](r.i)
	for _, def := range do.MustInvoke[[]rez.IntegrationDefinition](r.i) {
		if regErr := intgReg.Register(def); regErr != nil {
			return fmt.Errorf("failed to register integration package: %w", regErr)
		}
		if procPkg, isEventProcessor := def.(rez.ProviderEventProcessor); isEventProcessor {
			provider := def.Provider()
			if _, exists := eventProcessors[provider]; exists {
				return fmt.Errorf("failed to register event processor for provider %q: provider already registered", provider)
			}
			eventProcessors[provider] = procPkg
		}
	}
	return nil
}

func (r *runner) registerJobWorkers() error {
	return do.MustInvoke[jobsRegistrar](r.i).Init(do.MustInvoke[jobs.Definition](r.i))
}

func (r *runner) registerMessageHandlers() error {
	return do.MustInvoke[messageHandlerRegistrar](r.i).AddHandlers(do.MustInvoke[[]rez.MessageEventHandler](r.i)...)
}

func (r *runner) seedDevelopmentIdentity(ctx context.Context) error {
	cfg := do.MustInvoke[rez.Config](r.i)
	if !cfg.HttpServer.Auth.EnableDevSkipMode {
		return nil
	}

	sessions := do.MustInvoke[rez.AuthSessionService](r.i)
	orgs := do.MustInvoke[rez.OrganizationService](r.i)

	devIdentity := http.NewDevelopmentSessionIdentity()

	sess, sessionErr := sessions.CreateFromUserAuthResponse(ctx, devIdentity)
	if sessionErr != nil {
		return fmt.Errorf("seed development identity: %w", sessionErr)
	}
	ctx = execution.NewTenantContext(ctx, sess.TenantID)

	org, orgErr := orgs.Get(ctx, organization.ID(sess.OrganizationID))
	if orgErr != nil {
		return fmt.Errorf("load development organization: %w", orgErr)
	}

	client := do.MustInvoke[rez.Database](r.i).Client(ctx)

	queryOrgRole := client.OrganizationRole.Query().
		Where(organizationrole.OrganizationID(org.ID), organizationrole.UserID(sess.UserID))
	role, queryRoleErr := queryOrgRole.Only(ctx)
	if queryRoleErr != nil && !ent.IsNotFound(queryRoleErr) {
		return fmt.Errorf("load development admin role: %w", queryRoleErr)
	}

	var roleErr error
	if role == nil {
		createRole := client.OrganizationRole.Create().
			SetOrganizationID(org.ID).
			SetUserID(sess.UserID).
			SetRole(organizationrole.RoleAdmin)
		roleErr = createRole.Exec(ctx)
	} else if role.Role != organizationrole.RoleAdmin {
		updateRole := role.Update().
			SetRole(organizationrole.RoleAdmin)
		roleErr = updateRole.Exec(ctx)
	}
	if roleErr != nil {
		return fmt.Errorf("set development admin role: %w", roleErr)
	}

	var prefsErr error
	if org.Edges.Preferences == nil {
		prefsErr = client.OrganizationPreferences.Create().
			SetOrganizationID(org.ID).
			SetInitialSetupAt(time.Now().UTC()).
			Exec(ctx)
	} else if org.Edges.Preferences.InitialSetupAt.IsZero() {
		prefsErr = org.Edges.Preferences.Update().
			SetInitialSetupAt(time.Now().UTC()).
			Exec(ctx)
	}
	if prefsErr != nil {
		return fmt.Errorf("complete development organization setup: %w", prefsErr)
	}

	return nil
}
