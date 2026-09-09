package main

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/samber/do/v2"

	rez "github.com/rezible/rezible"
)

type appService struct {
	service          rez.LifecycleService
	cancelServiceCtx context.CancelFunc
	chStarted        <-chan error
}

type application struct {
	i            do.Injector
	cancelRunCtx context.CancelFunc
	services     []appService
}

func newApplication() *application {
	i := do.NewWithOpts(&do.InjectorOpts{}, basePackages)
	return &application{i: i}
}

func (a *application) init(ctx context.Context, provs ...WithProvider) {
	for _, p := range provs {
		p(ctx)(a.i)
	}
}

func (a *application) use(p func(do.Injector)) {
	p(a.i)
}

func (a *application) override[T any](prov do.Provider[T]) {
	do.Override(a.i, prov)
}

func (a *application) with[T any](fn func(T) error) error {
	t, invErr := do.Invoke[T](a.i)
	if invErr != nil {
		return fmt.Errorf("failed to invoke %T: %w", t, invErr)
	}
	return fn(t)
}

func (a *application) Run[S rez.LifecycleService](ctx context.Context) error {
	if bootstrapErr := a.setup(ctx); bootstrapErr != nil {
		return fmt.Errorf("bootstrap: %w", bootstrapErr)
	}
	return a.runLifecycleServices(ctx, a.invokeLifecycleServices(do.MustInvoke[S](a.i)))
}

func (a *application) invokeLifecycleServices(runSvcs ...rez.LifecycleService) []rez.LifecycleService {
	svcs := []rez.LifecycleService{
		do.MustInvoke[*river.JobService](a.i),
		do.MustInvoke[*watermill.MessageService](a.i),
	}
	type lifecycleServiceProvider interface {
		LifecycleService() rez.LifecycleService
	}
	for _, intg := range getAvailableIntegrationsWith[lifecycleServiceProvider](a.i) {
		if ls := intg.LifecycleService(); ls != nil {
			svcs = append(svcs, ls)
		}
	}
	svcs = append(svcs, runSvcs...)
	return svcs
}

func getServiceName(s rez.LifecycleService) string {
	return fmt.Sprintf("%T", s)
}

func (a *application) runLifecycleServices(ctx context.Context, services []rez.LifecycleService) (runErr error) {
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

func (a *application) startServices(ctx context.Context, services []rez.LifecycleService) error {
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
		svcName := getServiceName(svc)
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

func (a *application) waitForServices(ctx context.Context) error {
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

func (a *application) shutdownServices(ctx context.Context) error {
	var shutdownErr error
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	for _, running := range slices.Backward(a.services) {
		svcName := getServiceName(running.service)
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

func (a *application) shutdown(ctx context.Context) error {
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

func (a *application) setup(ctx context.Context) error {
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

func getAvailableIntegrationsWith[T any](i do.Injector) []T {
	ir := do.MustInvoke[rez.IntegrationRegistry](i)
	var pt []T
	for _, intg := range ir.GetAvailable() {
		if ls, hasLifecycle := intg.(T); hasLifecycle {
			pt = append(pt, ls)
		}
	}
	return pt
}

func (a *application) registerIntegrations() error {
	intgReg := do.MustInvoke[rez.IntegrationRegistry](a.i)
	eventProcessors := do.MustInvoke[rez.ProviderEventProcessorRegistry](a.i)
	for _, def := range do.MustInvoke[[]rez.IntegrationDefinition](a.i) {
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

func (a *application) registerJobWorkers() error {
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

	r := do.MustInvoke[JobsRegistrar](a.i)
	return r.RegisterWorkers(jobs.Definition{
		Workers:      workerDefs,
		PeriodicJobs: periodicJobs,
	})
}

func (a *application) registerMessageHandlers() error {
	handlers := do.MustInvoke[*db.SituationService](a.i).GetMessageHandlers()

	type ProvidesMessageHandlers interface {
		GetMessageHandlers() []rez.MessageEventHandler
	}
	for _, mhIntg := range getAvailableIntegrationsWith[ProvidesMessageHandlers](a.i) {
		handlers = append(handlers, mhIntg.GetMessageHandlers()...)
	}

	r := do.MustInvoke[MessageHandlerRegistrar](a.i)
	return r.AddHandlers(handlers...)
}

func (a *application) seedDevelopmentIdentity(ctx context.Context) error {
	cfg := do.MustInvoke[rez.Config](a.i)
	if !cfg.HttpServer.Auth.EnableDevSkipMode {
		return nil
	}

	sessions := do.MustInvoke[rez.AuthSessionService](a.i)
	orgs := do.MustInvoke[rez.OrganizationService](a.i)

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

	client := do.MustInvoke[rez.Database](a.i).Client(ctx)

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
