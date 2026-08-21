package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"

	rez "github.com/rezible/rezible"
)

func withMigrationService(i do.Injector, fn func(rez.MigrationService) error) error {
	ms, msErr := do.Invoke[rez.MigrationService](i)
	if msErr != nil {
		return fmt.Errorf("invoke migration service: %w", msErr)
	}
	return fn(ms)
}

func withConfig(i do.Injector, fn func(rez.Config) error) error {
	cfg, cfgErr := do.Invoke[rez.Config](i)
	if cfgErr != nil {
		return fmt.Errorf("invoke config: %w", cfgErr)
	}
	return fn(cfg)
}

func runServicesFor[Entrypoint rez.LifecycleService](ctx context.Context, i do.Injector) error {
	if regErr := registerBackgroundServicePackages(i); regErr != nil {
		return fmt.Errorf("register background services: %w", regErr)
	}

	lifecycles, lifecyclesErr := getServiceLifecyclesFor[Entrypoint](i)
	if lifecyclesErr != nil {
		return fmt.Errorf("get service lifecycles: %w", lifecyclesErr)
	}

	return lifecycles.run(ctx)
}

type serviceLifecycles map[string]rez.ServiceLifecycle

func getServiceLifecyclesFor[Entrypoint rez.LifecycleService](i do.Injector) (serviceLifecycles, error) {
	// invoke entrypoint service to invoke required service dependencies
	entrySvc, srvErr := do.Invoke[Entrypoint](i)
	if srvErr != nil {
		return nil, fmt.Errorf("invoke entrypoint %T: %w", entrySvc, srvErr)
	}

	services := make(serviceLifecycles)
	for _, desc := range i.ListInvokedServices() {
		svc, invErr := do.InvokeNamed[any](i, desc.Service)
		if invErr != nil {
			return nil, fmt.Errorf("failed to invoke %s: %w", desc.Service, invErr)
		}
		if intgSvc, isIntegration := svc.(rez.IntegrationPackage); isIntegration {
			// skipping unavailable integration
			available, intgErr := intgSvc.IsAvailable()
			if !available {
				if intgErr != nil {
					slog.Warn("integration package unavailable", "err", intgErr)
				}
				continue
			}
		}
		if lifecycleSvc, isLifecycleSvc := svc.(rez.LifecycleService); isLifecycleSvc {
			if ls := lifecycleSvc.Lifecycle(); ls != nil {
				svcName := strings.TrimLeft(fmt.Sprintf("%T", svc), "*")
				services[svcName] = *ls
			}
		}
	}

	return services, nil
}

func (ls serviceLifecycles) run(ctx context.Context) error {
	slog.Info("=== Starting Services ===")

	const lifecycleStopTimeout = 10 * time.Second

	group, runCtx := errgroup.WithContext(ctx)
	for name, lifecycle := range ls {
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

func registerBackgroundServicePackages(i do.Injector) error {
	return errors.Join(registerIntegrationPackages(i), registerJobWorkers(i))
}

func registerIntegrationPackages(i do.Injector) error {
	intgReg := do.MustInvoke[rez.IntegrationPackageRegistry](i)
	eventProcessors := do.MustInvoke[rez.ProviderEventProcessorRegistry](i)

	for _, desc := range i.ListProvidedServices() {
		if !strings.Contains(desc.Service, "internal/integrations") {
			continue
		}
		svc := do.MustInvokeNamed[any](i, desc.Service)
		if pkg, isIntgPkg := svc.(rez.IntegrationPackage); isIntgPkg {
			if regErr := intgReg.RegisterPackage(pkg); regErr != nil {
				return fmt.Errorf("failed to register integration package: %w", regErr)
			}
			if procPkg, isEventProcessor := pkg.(rez.ProviderEventProcessor); isEventProcessor {
				eventProcessors[pkg.Name()] = procPkg
			}
		}
	}

	return nil
}

func registerJobWorkers(i do.Injector) error {
	return errors.Join(
		registerJobWorker[jobs.StartAgentSession](i),
		registerJobWorker[jobs.InvokeAgentTurn](i),
		registerJobWorker[jobs.SyncIntegrationSourceEvents](i),
	)
}

func registerJobWorker[A river.JobArgs](i do.Injector) error {
	jobs.RegisterWorker(do.MustInvoke[jobs.Worker[A]](i))
	return nil
}
