package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rezible/rezible/telemetry"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"
)

type app struct {
	i do.Injector
}

func newApp() *app {
	return &app{i: do.New()}
}

func (r *app) setup(ctx context.Context) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	if cfgErr := r.loadConfig(ctx); cfgErr != nil {
		return nil, fmt.Errorf("failed to load config: %w", cfgErr)
	}

	provideDependencies(r.i)

	return ctx, nil
}

func (r *app) loadConfig(ctx context.Context) error {
	loader := koanf.NewConfigLoader(koanf.ConfigLoaderOptions{
		LoadEnvironment: true,
	})
	cfg, cfgErr := loader.LoadConfig(ctx)
	if cfgErr != nil {
		return cfgErr
	}
	do.ProvideValue[rez.Config](r.i, cfg)
	return nil
}

func (r *app) printConfig() error {
	cfg := do.MustInvoke[rez.Config](r.i)
	fmt.Printf("%+v\n", cfg)
	return nil
}

func (r *app) start(ctx context.Context, services []rez.LifecycleService) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx).WithFirstError()
		for _, l := range services {
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

func (r *app) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdown := r.i.ShutdownWithContext(ctx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

func initializeLifecycleServicesFor[T any](ctx context.Context, a *app) ([]rez.LifecycleService, error) {
	if initErr := a.initializeServices(ctx); initErr != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", initErr)
	}
	// invoke entrypoint service to load required service dependencies
	if _, srvErr := do.Invoke[T](a.i); srvErr != nil {
		return nil, fmt.Errorf("failed to initialize http server: %v", srvErr)
	}
	var services []rez.LifecycleService
	for _, desc := range a.i.ListInvokedServices() {
		s, invErr := do.InvokeNamed[any](a.i, desc.Service)
		if invErr != nil {
			return nil, fmt.Errorf("failed to invoke: %v", invErr)
		}
		if svc, ok := s.(rez.LifecycleService); ok {
			services = append(services, svc)
		}
	}
	return services, nil
}

func (r *app) serveHttp(ctx context.Context) error {
	services, svcErr := initializeLifecycleServicesFor[*http.Server](ctx, r)
	if svcErr != nil {
		return fmt.Errorf("failed to initialize http server: %w", svcErr)
	}
	return r.start(ctx, services)
}

func (r *app) applySchemaMigrations(ctx context.Context, direction string) error {
	// TODO: this should run as a lifecycle service
	//services, svcErr := initializeLifecycleServicesFor[*postgres.MigrationService](ctx, r)
	//if svcErr != nil {
	//	return fmt.Errorf("failed to initialize http server: %w", svcErr)
	//}
	//return r.start(ctx, services)

	mc := do.MustInvoke[*postgres.MigrationService](r.i)
	return mc.Run(ctx, direction)
}

func (r *app) createSchemaMigration(ctx context.Context, name string) error {
	cfg := do.MustInvoke[rez.Config](r.i)
	return postgres.CreateSchemaMigration(ctx, cfg.Postgres, name)
}

func (r *app) updateMigrationChecksumFile() error {
	return postgres.UpdateMigrationsChecksum()
}

func (r *app) printOpenApiSpec(asJson bool) error {
	spec, specErr := oapiv1.GetSpec(asJson)
	if specErr != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", specErr)
	}
	fmt.Printf("%s", spec)
	return nil
}

func (r *app) initializeServices(ctx context.Context) error {
	cfg := do.MustInvoke[rez.Config](r.i)

	tel, telErr := telemetry.NewOpenTelemetryService(ctx, cfg)
	if telErr != nil {
		return fmt.Errorf("create telemetry service: %w", telErr)
	}
	do.ProvideValue[rez.TelemetryService](r.i, tel)

	do.Provide(r.i, func(i do.Injector) (*postgres.PgxPool, error) {
		return postgres.MakePgxPool(ctx, cfg.Postgres, false)
	})

	eventSvc := do.MustInvoke[rez.ProviderEventService](r.i)
	r.registerEventProjectionServices(eventSvc)

	intgReg := do.MustInvoke[*integrations.PackageRegistry](r.i)
	if pkgsErr := r.registerIntegrationPackages(intgReg); pkgsErr != nil {
		return fmt.Errorf("register integration packages: %w", pkgsErr)
	}

	return nil
}

func (r *app) registerEventProjectionServices(s rez.ProviderEventService) {
	s.RegisterProjectionHandler("knowledge", do.MustInvoke[*db.KnowledgeService](r.i),
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
	)
	s.RegisterProjectionHandler("users", do.MustInvoke[*db.UserService](r.i),
		projections.SubjectKindUser)
	s.RegisterProjectionHandler("incidents", do.MustInvoke[*db.IncidentService](r.i),
		projections.SubjectKindIncident)
	s.RegisterProjectionHandler("alerts", do.MustInvoke[*db.AlertService](r.i),
		projections.SubjectKindAlert)
}

func (r *app) registerIntegrationPackages(reg *integrations.PackageRegistry) error {
	for _, desc := range r.i.ListProvidedServices() {
		svc := desc.Service
		if strings.Contains(svc, "internal/integrations") {
			if pkg, ok := do.MustInvokeNamed[any](r.i, svc).(rez.IntegrationPackage); ok {
				if regErr := reg.RegisterPackage(pkg); regErr != nil {
					return fmt.Errorf("failed to register integration package: %w", regErr)
				}
			}
		}
	}
	return nil
}
