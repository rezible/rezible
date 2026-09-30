package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	demoprovider "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/samber/do/v2"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"
)

type (
	Package         = func(do.Injector)
	Provider[T any] = do.Provider[T]

	Application struct {
		i            do.Injector
		ready        chan struct{}
		cancelRunCtx context.CancelFunc
		services     []appService
	}

	appService struct {
		service          rez.LifecycleService
		cancelServiceCtx context.CancelFunc
		chStarted        <-chan error
	}
)

func NewApplication() *Application {
	i := do.NewWithOpts(&do.InjectorOpts{}, basePackages)
	return &Application{i: i}
}

func (a *Application) Init(ctx context.Context) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	applicationPackages(ctx)(a.i)

	return ctx, nil
}

func (a *Application) Override[T any](prov Provider[T]) {
	do.Override(a.i, prov)
}

func (a *Application) invoke[T any]() (T, error) {
	t, invErr := do.Invoke[T](a.i)
	if invErr != nil {
		return t, fmt.Errorf("failed to invoke %T: %w", t, invErr)
	}
	return t, nil
}

func (a *Application) mustInvoke[T any]() T {
	return do.MustInvoke[T](a.i)
}

func (a *Application) With[T any](fn func(T) error) error {
	t, invErr := a.invoke[T]()
	if invErr != nil {
		return invErr
	}
	return fn(t)
}

func (a *Application) RunLifecycle(ctx context.Context, svcs ...rez.LifecycleService) error {
	if bootstrapErr := a.setup(ctx); bootstrapErr != nil {
		return fmt.Errorf("bootstrap: %w", bootstrapErr)
	}
	return a.runLifecycleServices(ctx, append(a.invokeBaseLifecycleServices(), svcs...))
}

func (a *Application) RunAiEvalScenario(ctx context.Context, evalName string, writer io.Writer) error {
	return a.With(func(svc rezai.EvalScenarioRunner) error {
		result, runErr := svc.RunScenario(ctx, evalName)
		if runErr != nil {
			return fmt.Errorf("failed to run evaluation %q: %w", evalName, runErr)
		}
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if jsonErr := encoder.Encode(result); jsonErr != nil {
			return fmt.Errorf("encode evaluation result: %w", jsonErr)
		}
		if result.Status != rezai.EvalRunStatusPassed {
			return fmt.Errorf("evaluation %s", result.Status)
		}
		return nil
	})
}

func (a *Application) Shutdown(ctx context.Context) error {
	cancelCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	shutdown := a.i.ShutdownWithContext(cancelCtx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

func (a *Application) invokeBaseLifecycleServices() []rez.LifecycleService {
	svcs := []rez.LifecycleService{
		a.mustInvoke[*watermill.MessageQueue](),
		a.mustInvoke[*river.JobService](),
	}
	for _, intg := range getAvailableIntegrationsWith[rez.LifecycleServiceProvider](a.i) {
		if ls := intg.LifecycleService(); ls != nil {
			svcs = append(svcs, ls)
		}
	}
	return svcs
}

// startLifecycle runs the lifecycle in the background. ready closes once all
// services have started; done receives the lifecycle result exactly once.
func (a *Application) startLifecycle(ctx context.Context) (ready <-chan struct{}, done <-chan error) {
	a.ready = make(chan struct{})
	lifecycleResult := make(chan error, 1)
	go func() {
		lifecycleResult <- a.RunLifecycle(ctx)
	}()
	return a.ready, lifecycleResult
}

func (a *Application) runLifecycleFunc(ctx context.Context, fn func(context.Context) error) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ready, lifecycleResult := a.startLifecycle(runCtx)
	select {
	case <-ready:
	case lifecycleErr := <-lifecycleResult:
		return lifecycleErr
	}

	fnErr := fn(runCtx)
	cancel()
	return errors.Join(fnErr, <-lifecycleResult)
}

func (a *Application) getServiceName(s rez.LifecycleService) string {
	return fmt.Sprintf("%T", s)
}

func (a *Application) runLifecycleServices(ctx context.Context, services []rez.LifecycleService) (runErr error) {
	slog.Info("=== Starting Services ===")
	runCtx, cancelRunCtx := context.WithCancel(context.WithoutCancel(ctx))
	a.cancelRunCtx = cancelRunCtx
	defer cancelRunCtx()

	a.services = make([]appService, 0, len(services))
	defer func() {
		if shutdownErr := a.shutdownServices(runCtx); shutdownErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("shutdown: %w", shutdownErr))
		}
		slog.Info("=== Services Stopped ===")
	}()

	if startErr := a.startServices(runCtx, services); startErr != nil {
		return fmt.Errorf("start services: %w", startErr)
	}
	if a.ready != nil {
		close(a.ready)
	}

	return a.waitForServices(ctx)
}

func (a *Application) startServices(ctx context.Context, services []rez.LifecycleService) error {
	startupCtx, cancelStartupCtx := context.WithTimeout(ctx, 30*time.Second)
	defer cancelStartupCtx()

	runService := func(svcCtx context.Context, svc rez.LifecycleService, chStarted chan error, chReady chan struct{}) {
		chStarted <- svc.Run(svcCtx, chReady)
	}

	waitForServiceReady := func(rs *appService, chReady chan struct{}) (bool, error) {
		select {
		case serviceErr := <-rs.chStarted:
			if serviceErr != nil {
				return false, fmt.Errorf("stopped before ready: %w", serviceErr)
			}
			return false, fmt.Errorf("stopped before ready")
		case <-chReady:
			// service started & entered ready state
			return true, nil
		case <-startupCtx.Done():
			if ctx.Err() == nil {
				a.cancelRunCtx()
				return false, startupCtx.Err()
			}
			slog.Debug("startup context finished, no error?")
			return false, nil
		}
	}

	for _, svc := range services {
		svcName := a.getServiceName(svc)
		serviceCtx, cancelServiceFn := context.WithCancel(ctx)
		chStarted := make(chan error, 1)
		chReady := make(chan struct{})

		rs := &appService{
			service:          svc,
			cancelServiceCtx: cancelServiceFn,
			chStarted:        chStarted,
		}
		a.services = append(a.services, *rs)

		go runService(serviceCtx, svc, chStarted, chReady)

		slog.Info("starting service", "service", svcName)

		ready, startupErr := waitForServiceReady(rs, chReady)
		if !ready {
			if startupErr != nil {
				startupErr = fmt.Errorf("start %s: %w", svcName, startupErr)
			}
			return startupErr
		}
	}
	return nil
}

func (a *Application) waitForServices(ctx context.Context) error {
	serviceExit := make(chan error, len(a.services))
	waitForService := func(s appService) {
		if serviceErr := <-s.chStarted; serviceErr != nil {
			serviceExit <- fmt.Errorf("%T: %w", s.service, serviceErr)
			return
		}
		serviceExit <- fmt.Errorf("%T stopped unexpectedly", s.service)
	}

	for _, rs := range a.services {
		go waitForService(rs)
	}

	select {
	case <-ctx.Done():
		return nil
	case serviceErr := <-serviceExit:
		return fmt.Errorf("run services: %w", serviceErr)
	}
}

func (a *Application) shutdownServices(ctx context.Context) error {
	var shutdownErr error
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	for _, running := range slices.Backward(a.services) {
		svcName := a.getServiceName(running.service)
		if stopErr := running.service.Shutdown(stopCtx); stopErr != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown %s: %w", svcName, stopErr))
			slog.Error("service shutdown failed",
				"service", svcName,
				"error", stopErr,
			)
		}
		running.cancelServiceCtx()
	}
	if stopCtx.Err() != nil {
		a.cancelRunCtx()
	}
	return shutdownErr
}

func (a *Application) setup(ctx context.Context) error {
	if intgsErr := a.registerIntegrations(); intgsErr != nil {
		return fmt.Errorf("register integrations: %w", intgsErr)
	}

	if jobsErr := a.registerJobWorkers(); jobsErr != nil {
		return fmt.Errorf("register job workers: %w", jobsErr)
	}

	if msgsErr := a.registerMessageHandlers(); msgsErr != nil {
		return fmt.Errorf("register message handlers: %w", msgsErr)
	}

	if devErr := a.registerDevelopmentData(ctx); devErr != nil {
		return fmt.Errorf("register dev: %w", devErr)
	}

	return nil
}

func (a *Application) registerIntegrations() error {
	intgReg := a.mustInvoke[rez.IntegrationRegistry]()
	eventProcessors := a.mustInvoke[rez.ProviderEventProcessorRegistry]()
	for _, def := range a.mustInvoke[[]rez.IntegrationDefinition]() {
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

func (a *Application) registerJobWorkers() error {
	def, defErr := do.InvokeNamed[jobs.Definition](a.i, "jobs-default")
	if defErr != nil {
		return fmt.Errorf("get jobs definition: %w", defErr)
	}
	return a.With(func(r jobs.Registrar) error {
		return r.Register(def)
	})
}

func (a *Application) registerMessageHandlers() error {
	def, defErr := do.InvokeNamed[messages.Definition](a.i, "messages-default")
	if defErr != nil {
		return fmt.Errorf("get messages definition: %w", defErr)
	}
	return a.With(func(r messages.Registrar) error {
		return r.Register(def)
	})
}

func (a *Application) registerDevelopmentData(ctx context.Context) error {
	// TODO: check current environment == dev

	// used to ensure evals are registered when running the `serve` command
	if _, workflowErr := do.Invoke[rezai.EvalScenarioRunner](a.i); workflowErr != nil {
		return fmt.Errorf("initialize ai eval workflows: %w", workflowErr)
	}
	if seedErr := a.seedDevelopmentIdentity(ctx); seedErr != nil {
		return fmt.Errorf("seed dev identity: %w", seedErr)
	}
	return nil
}

func (a *Application) makeDevelopmentAuthSession(ctx context.Context) (*ent.UserAuthSession, error) {
	svc := a.mustInvoke[rez.AuthSessionService]()
	sess, sessionErr := svc.CreateFromUserAuthResponse(execution.NewSystemContext(ctx), makeDevelopmentAuthSession())
	if sessionErr != nil {
		return nil, fmt.Errorf("http development session: %w", sessionErr)
	}
	return sess, nil
}

func makeDevelopmentAuthSession() *rez.UserAuthProviderSession {
	return &rez.UserAuthProviderSession{
		User: ent.User{
			Email:          "test@dev.rezible.com",
			Name:           "Dev User",
			AuthProviderID: "dev-user",
		},
		Org: ent.Organization{
			Name:           "Dev Org",
			AuthProviderID: "dev-org",
		},
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func (a *Application) seedDevelopmentIdentity(ctx context.Context) error {
	if !a.mustInvoke[rez.Config]().HttpServer.Auth.EnableDevSkipMode {
		return nil
	}

	sess, sessErr := a.makeDevelopmentAuthSession(ctx)
	if sessErr != nil {
		return fmt.Errorf("seed development identity: %w", sessErr)
	}
	ctx = execution.NewTenantContext(ctx, sess.TenantID)

	orgs := a.mustInvoke[rez.OrganizationService]()
	org, orgErr := orgs.Get(ctx, organization.ID(sess.OrganizationID))
	if orgErr != nil {
		return fmt.Errorf("load development organization: %w", orgErr)
	}

	client := a.mustInvoke[rez.Database]().Client(ctx)

	queryOrgRole := client.OrganizationRole.Query().
		Where(organizationrole.OrganizationID(org.ID), organizationrole.UserID(sess.UserID))
	orgRole, queryOrgRoleErr := queryOrgRole.Only(ctx)
	if queryOrgRoleErr != nil && !ent.IsNotFound(queryOrgRoleErr) {
		return fmt.Errorf("load development admin role: %w", queryOrgRoleErr)
	}

	var roleErr error
	if orgRole == nil {
		roleErr = client.OrganizationRole.Create().
			SetOrganizationID(org.ID).
			SetUserID(sess.UserID).
			SetRole(organizationrole.RoleAdmin).
			Exec(ctx)
	} else if orgRole.Role != organizationrole.RoleAdmin {
		roleErr = orgRole.Update().
			SetRole(organizationrole.RoleAdmin).
			Exec(ctx)
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

func (a *Application) setupDemo(ctx context.Context) error {
	return a.runLifecycleFunc(ctx, func(ctx context.Context) error {
		sess, sessErr := a.makeDevelopmentAuthSession(ctx)
		if sessErr != nil {
			return fmt.Errorf("seed development identity: %w", sessErr)
		}
		ctx = execution.NewUserContext(ctx, sess)

		installErr := a.With(func(i rez.IntegrationService) error {
			_, installErr := i.InstallNew(ctx, "demo", []byte("{}"))
			return installErr
		})
		if installErr != nil {
			return fmt.Errorf("install demo integration: %w", installErr)
		}

		fmt.Printf("wait for integration data ingestion...\n")
		time.Sleep(time.Second * 5)

		return demoprovider.SeedDemoData(
			ctx,
			a.mustInvoke[rez.Database](),
			a.mustInvoke[rez.KnowledgeGraphQueryService](),
			a.mustInvoke[rez.IncidentService](),
			a.mustInvoke[rez.RetrospectiveService](),
			a.mustInvoke[rez.SystemAnalysisService](),
			a.mustInvoke[rez.EventsService](),
			a.mustInvoke[rez.SituationService](),
		)
	})
}
