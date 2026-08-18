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
	"github.com/sourcegraph/conc/pool"

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

type startable interface {
	Start(context.Context) error
}

func startServicesFor[Entrypoint startable](ctx context.Context, i do.Injector) error {
	if regErr := registerServerPackages(i); regErr != nil {
		return fmt.Errorf("register packages: %w", regErr)
	}

	// invoke entrypoint service to load required service dependencies
	entrySvc, srvErr := do.Invoke[Entrypoint](i)
	if srvErr != nil {
		return fmt.Errorf("initialize entrypoint %T: %v", entrySvc, srvErr)
	}

	var services []startable
	for _, desc := range i.ListInvokedServices() {
		svc, invErr := do.InvokeNamed[any](i, desc.Service)
		if invErr != nil {
			return fmt.Errorf("failed to invoke %s: %v", desc.Service, invErr)
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
		if startableSvc, isStartable := svc.(startable); isStartable {
			services = append(services, startableSvc)
		}
	}

	return startServices(ctx, services)
}

func startServices(ctx context.Context, svcs []startable) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx).WithFirstError()
		for _, l := range svcs {
			slog.Info("Starting " + strings.TrimLeft(fmt.Sprintf("%T", l), "*"))
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()

	slog.Info("=== Starting Services ===")
	var servicesErr error
	select {
	case <-ctx.Done():
		servicesErr = ctx.Err()
	case poolErr := <-errChan:
		servicesErr = poolErr
	}
	slog.Info("=== Stopping Services ===")

	if servicesErr != nil && !errors.Is(servicesErr, context.Canceled) {
		return fmt.Errorf("run services: %s", servicesErr.Error())
	}
	return nil
}

func shutdownServices(ctx context.Context, i do.Injector) error {
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

func registerServerPackages(i do.Injector) error {
	if intgErr := registerIntegrationPackages(i); intgErr != nil {
		return fmt.Errorf("auto-register integration packages: %w", intgErr)
	}
	if jobsErr := registerJobWorkers(i); jobsErr != nil {
		return fmt.Errorf("register job workers: %w", jobsErr)
	}
	return nil
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

func registerJobWorker[A river.JobArgs](i do.Injector) error {
	jobs.RegisterWorker(do.MustInvoke[jobs.Worker[A]](i))
	return nil
}

func registerJobWorkers(i do.Injector) error {
	return errors.Join(
		registerJobWorker[jobs.StartAgentSession](i),
		registerJobWorker[jobs.InvokeAgentTurn](i),
		registerJobWorker[jobs.SyncIntegrationSourceEvents](i),
	)
}
