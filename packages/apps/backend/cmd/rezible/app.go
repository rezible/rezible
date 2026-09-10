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

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/samber/do/v2"

	rez "github.com/rezible/rezible"
)

type (
	Package         = func(do.Injector)
	Provider[T any] = do.Provider[T]
	InitProvider    = func(context.Context) Package

	Application struct {
		i            do.Injector
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

func (a *Application) Init(ctx context.Context, provs ...InitProvider) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	for _, p := range provs {
		p(ctx)(a.i)
	}

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
		return fmt.Errorf("failed to invoke %T: %w", t, invErr)
	}
	return fn(t)
}

func (a *Application) RunLifecycle[S rez.LifecycleService](ctx context.Context) error {
	if bootstrapErr := a.setup(ctx); bootstrapErr != nil {
		return fmt.Errorf("bootstrap: %w", bootstrapErr)
	}
	return a.runLifecycleServices(ctx, a.invokeLifecycleServices(a.mustInvoke[S]()))
}

func (a *Application) RunAiEvalScenario(ctx context.Context, evalName string, writer io.Writer) error {
	return a.With(func(svc rezai.EvalScenarioRunner) error {
		result, runErr := svc.RunNamedScenario(ctx, evalName)
		if runErr != nil || result == nil {
			return fmt.Errorf("failed to run scenario: %w", runErr)
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

func (a *Application) invokeLifecycleServices(runSvcs ...rez.LifecycleService) []rez.LifecycleService {
	svcs := []rez.LifecycleService{
		a.mustInvoke[*river.JobService](),
		a.mustInvoke[*watermill.MessageService](),
	}
	for _, intg := range a.getAvailableIntegrationsWith[rez.LifecycleServiceProvider]() {
		if ls := intg.LifecycleService(); ls != nil {
			svcs = append(svcs, ls)
		}
	}
	svcs = append(svcs, runSvcs...)
	return svcs
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

	if seedErr := a.seedDevelopmentIdentity(ctx); seedErr != nil {
		return fmt.Errorf("seed dev identity: %w", seedErr)
	}

	return nil
}

func (a *Application) getAvailableIntegrationsWith[T any]() []T {
	ir := a.mustInvoke[rez.IntegrationRegistry]()
	var pt []T
	for _, intg := range ir.GetAvailable() {
		if ls, hasLifecycle := intg.(T); hasLifecycle {
			pt = append(pt, ls)
		}
	}
	return pt
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
	workerProviders := makeDefaultJobWorkerProviders()
	workerDefs := make([]jobs.WorkerDefinition, len(workerProviders))
	for idx, provider := range workerProviders {
		def, provideErr := provider(a.i)
		if provideErr != nil {
			return fmt.Errorf("provide job worker: %w", provideErr)
		}
		workerDefs[idx] = def
	}

	periodicJobs := []*jobs.PeriodicJob{
		jobs.CloseInactiveAlertEpisodesPeriodicJob,
	}

	r := a.mustInvoke[JobsRegistrar]()
	return r.RegisterWorkers(jobs.Definition{
		Workers:      workerDefs,
		PeriodicJobs: periodicJobs,
	})
}

func (a *Application) registerMessageHandlers() error {
	handlers := a.mustInvoke[*db.SituationService]().GetMessageHandlers()

	type ProvidesMessageHandlers interface {
		GetMessageHandlers() []rez.MessageEventHandler
	}
	for _, mhIntg := range a.getAvailableIntegrationsWith[ProvidesMessageHandlers]() {
		handlers = append(handlers, mhIntg.GetMessageHandlers()...)
	}

	r := a.mustInvoke[MessageHandlerRegistrar]()
	return r.AddHandlers(handlers...)
}

func (a *Application) seedDevelopmentIdentity(ctx context.Context) error {
	cfg := a.mustInvoke[rez.Config]()
	if !cfg.HttpServer.Auth.EnableDevSkipMode {
		return nil
	}

	sessions := a.mustInvoke[rez.AuthSessionService]()
	orgs := a.mustInvoke[rez.OrganizationService]()

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

	client := a.mustInvoke[rez.Database]().Client(ctx)

	queryOrgRole := client.OrganizationRole.Query().
		Where(organizationrole.OrganizationID(org.ID), organizationrole.UserID(sess.UserID))
	role, queryRoleErr := queryOrgRole.Only(ctx)
	if queryRoleErr != nil && !ent.IsNotFound(queryRoleErr) {
		return fmt.Errorf("load development admin role: %w", queryRoleErr)
	}

	var roleErr error
	if role == nil {
		roleErr = client.OrganizationRole.Create().
			SetOrganizationID(org.ID).
			SetUserID(sess.UserID).
			SetRole(organizationrole.RoleAdmin).
			Exec(ctx)
	} else if role.Role != organizationrole.RoleAdmin {
		roleErr = role.Update().
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
