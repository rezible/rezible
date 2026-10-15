package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/samber/do/v2"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"
)

type (
	Package         = func(do.Injector)
	Provider[T any] = do.Provider[T]

	Application struct {
		i            do.Injector
		profiles     []Package
		prepared     bool
		telemetry    *opentelemetry.Service
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

// NewApplication builds the application's container. The profiles adjust it for one way of running it; they
// are applied by Init, after the packages that need its context.
func NewApplication(profiles ...Package) *Application {
	i := do.NewWithOpts(&do.InjectorOpts{}, basePackages)
	return &Application{i: i, profiles: profiles}
}

func (a *Application) Init(ctx context.Context, withTelemetry bool) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	applicationPackages(ctx)(a.i)
	for _, profile := range a.profiles {
		profile(a.i)
	}

	if withTelemetry {
		if telErr := a.initTelemetry(ctx); telErr != nil {
			return nil, fmt.Errorf("telemetry: %w", telErr)
		}
	}

	return ctx, nil
}

// initTelemetry installs the logger and the OpenTelemetry providers before any service is built, so every service logs and traces through them.
func (a *Application) initTelemetry(ctx context.Context) error {
	cfg, cfgErr := a.invoke[rez.Config]()
	if cfgErr != nil {
		return cfgErr
	}
	telemetry, telemetryErr := opentelemetry.NewOpenTelemetryService(ctx, cfg)
	if telemetryErr != nil {
		return fmt.Errorf("telemetry: %w", telemetryErr)
	}
	if initErr := telemetry.Init(); initErr != nil {
		return fmt.Errorf("init telemetry: %w", initErr)
	}
	a.telemetry = telemetry
	return nil
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
	if prepareErr := a.prepare(); prepareErr != nil {
		return fmt.Errorf("prepare: %w", prepareErr)
	}
	return a.runLifecycleServices(ctx, append(a.invokeBaseLifecycleServices(), svcs...))
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
	// Telemetry stops last, so it flushes what the services recorded while stopping.
	if a.telemetry != nil {
		if telemetryErr := a.telemetry.Shutdown(cancelCtx); telemetryErr != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown telemetry: %w", telemetryErr))
		}
	}
	return shutdownErr
}

func (a *Application) invokeBaseLifecycleServices() []rez.LifecycleService {
	svcs := []rez.LifecycleService{
		a.mustInvoke[*watermill.MessageQueue](),
		a.mustInvoke[*river.JobService](),
	}
	ir := a.mustInvoke[*integrations.Registry]()
	for _, intg := range ir.All[rez.LifecycleServiceProvider]() {
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

// prepare registers the job workers and message handlers, once. Anything that inserts a job needs it first.
func (a *Application) prepare() error {
	if a.prepared {
		return nil
	}
	if jobsErr := a.registerJobWorkers(); jobsErr != nil {
		return fmt.Errorf("register job workers: %w", jobsErr)
	}
	if msgsErr := a.registerMessageHandlers(); msgsErr != nil {
		return fmt.Errorf("register message handlers: %w", msgsErr)
	}
	a.prepared = true
	return nil
}

func (a *Application) registerJobWorkers() error {
	def, defErr := do.Invoke[jobs.Definition](a.i)
	if defErr != nil {
		return fmt.Errorf("get jobs definition: %w", defErr)
	}
	return a.With(func(r jobs.Registrar) error {
		return r.Register(def)
	})
}

func (a *Application) registerMessageHandlers() error {
	def, defErr := do.Invoke[messages.Definition](a.i)
	if defErr != nil {
		return fmt.Errorf("get messages definition: %w", defErr)
	}
	return a.With(func(r messages.Registrar) error {
		return r.Register(def)
	})
}
