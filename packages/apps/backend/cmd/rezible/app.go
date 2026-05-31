package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/http/oidc"
	fakeprovider "github.com/rezible/rezible/internal/integrations/fake"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/slack"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/telemetry"
)

type lifecycleService interface {
	Start(context.Context) error
	Shutdown(context.Context) error
}

type appRunner struct {
	services []lifecycleService
}

func makeAppRunner(i do.Injector) (*appRunner, error) {
	// invoke http server
	if _, srvErr := do.Invoke[*http.Server](i); srvErr != nil {
		return nil, fmt.Errorf("failed to initialize server: %v", srvErr)
	}
	var services []lifecycleService
	for _, desc := range i.ListInvokedServices() {
		s, invErr := do.InvokeNamed[any](i, desc.Service)
		if invErr != nil {
			return nil, fmt.Errorf("failed to invoke: %v", invErr)
		}
		if svc, ok := s.(lifecycleService); ok {
			services = append(services, svc)
		}
	}
	return &appRunner{services: services}, nil
}

func (r *appRunner) start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx).WithFirstError()
		for _, l := range r.services {
			slog.Info("Starting " + strings.TrimLeft(fmt.Sprintf("%T", l), "*"))
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()
	fmt.Println("=== Starting Services ===")
	var servicesErr error
	select {
	case <-ctx.Done():
		servicesErr = ctx.Err()
	case poolErr := <-errChan:
		servicesErr = poolErr
	}
	fmt.Println("=== Stopping Services ===")
	if servicesErr != nil && !errors.Is(servicesErr, context.Canceled) {
		return fmt.Errorf("run services: %s", servicesErr.Error())
	}
	return nil
}

func provideDependencies(ctx context.Context, i do.Injector) error {
	do.Provide(i, makeAppRunner)

	cl, clErr := koanf.NewConfigLoader(koanf.ConfigLoaderOptions{LoadEnvironment: true})
	if clErr != nil {
		return fmt.Errorf("config loader: %w", clErr)
	}
	do.ProvideValue(i, cl)
	do.ProvideValue[rez.ConfigLoader](i, cl)

	cfg, cfgErr := cl.LoadConfig(ctx)
	if cfgErr != nil {
		return fmt.Errorf("load config: %w", cfgErr)
	}
	do.ProvideValue[rez.Config](i, cfg)

	do.Provide(i, func(i do.Injector) (*telemetry.Service, error) {
		return telemetry.NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
	})
	do.MustAs[*telemetry.Service, rez.TelemetryService](i)

	do.Provide(i, func(i do.Injector) (*postgres.MigratorClient, error) {
		return postgres.NewAdminMigratorClient(do.MustInvoke[rez.Config](i).Postgres)
	})
	do.Provide(i, func(i do.Injector) (*postgres.DatabaseClient, error) {
		return postgres.NewDatabaseClient(ctx, do.MustInvoke[rez.Config](i).Postgres)
	})
	do.Provide(i, func(i do.Injector) (*ent.Client, error) {
		pgDbc := do.MustInvoke[*postgres.DatabaseClient](i)
		return pgDbc.Client(), nil
	})
	do.Provide(i, func(i do.Injector) (rez.JobService, error) {
		pgDbc := do.MustInvoke[*postgres.DatabaseClient](i)
		return river.NewJobService(do.MustInvoke[rez.TelemetryService](i), pgDbc.Pool())
	})

	do.Provide(i, func(i do.Injector) (rez.MessageService, error) {
		return watermill.NewMessageService(do.MustInvoke[rez.TelemetryService](i))
	})

	do.ProvideValue(i, projections.NewEventProjectionHandlerRegistry())

	do.Provide(i, func(i do.Injector) (*integrations.PackageRegistry, error) {
		return integrations.NewPackageRegistry(do.MustInvoke[rez.TelemetryService](i)), nil
	})

	do.Provide(i, func(i do.Injector) (*oidc.AuthSessionService, error) {
		return oidc.NewAuthSessionService(
			cfg,
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		)
	})
	do.MustAs[*oidc.AuthSessionService, http.UserAuthSessionService](i)

	do.Package(provideServices()...)(i)

	do.Provide(i, func(i do.Injector) (oapiv1.Handler, error) {
		return makeOpenApiHandler(i), nil
	})

	do.Provide(i, func(i do.Injector) (*http.Server, error) {
		return http.NewServer(
			cfg,
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[http.UserAuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[*integrations.PackageRegistry](i).GetWebhookHandlers(),
		)
	})

	if intgsErr := registerIntegrationPackages(i); intgsErr != nil {
		return fmt.Errorf("register integration packages: %w", intgsErr)
	}

	return nil
}

func registerIntegrationPackages(i do.Injector) error {
	do.Package(
		fakeprovider.Package,
		slack.Package,
		github.Package,
		google.Package,
	)(i)

	var pkgs []rez.IntegrationPackage
	for _, desc := range i.ListProvidedServices() {
		svc := desc.Service
		if strings.Contains(svc, "internal/integrations") {
			if pkg, ok := do.MustInvokeNamed[any](i, svc).(rez.IntegrationPackage); ok {
				pkgs = append(pkgs, pkg)
			}
		}
	}

	intgReg := do.MustInvoke[*integrations.PackageRegistry](i)
	for _, pkg := range pkgs {
		if regErr := intgReg.RegisterPackage(pkg); regErr != nil {
			return fmt.Errorf("failed to register integration package: %w", regErr)
		}
	}

	return nil
}

func makeOpenApiHandler(i do.Injector) oapiv1.Handler {
	return apiv1.NewHandler(
		do.MustInvoke[*ent.Client](i),
		do.MustInvoke[rez.AlertService](i),
		do.MustInvoke[rez.OrganizationService](i),
		do.MustInvoke[rez.UserService](i),
		do.MustInvoke[rez.DocumentsService](i),
		do.MustInvoke[rez.DebriefService](i),
		do.MustInvoke[rez.IncidentService](i),
		do.MustInvoke[rez.IntegrationsService](i),
		do.MustInvoke[rez.EventsService](i),
		do.MustInvoke[rez.EventAnnotationsService](i),
		do.MustInvoke[rez.OncallRostersService](i),
		do.MustInvoke[rez.OncallShiftsService](i),
		do.MustInvoke[rez.OncallMetricsService](i),
		do.MustInvoke[rez.PlaybookService](i),
		do.MustInvoke[rez.RetrospectiveService](i),
		do.MustInvoke[rez.SystemTopologyService](i),
	)
}

func provideServices() []func(do.Injector) {
	return []func(do.Injector){
		do.Lazy(func(i do.Injector) (rez.IntegrationsService, error) {
			return db.NewIntegrationsService(
				do.MustInvoke[rez.Config](i).App,
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
				do.MustInvoke[*integrations.PackageRegistry](i))
		}),
		do.Lazy(func(i do.Injector) (rez.ProviderEventService, error) {
			return db.NewProviderEventService(
				do.MustInvoke[rez.TelemetryService](i),
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
				do.MustInvoke[rez.IntegrationsService](i),
				do.MustInvoke[*projections.EventProjectionHandlerRegistry](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.OrganizationService, error) {
			return db.NewOrganizationsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.UserService, error) {
			return db.NewUserService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.OrganizationService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.TeamService, error) {
			return db.NewTeamService(do.MustInvoke[*ent.Client](i))
		}),
		do.Lazy(func(i do.Injector) (rez.EventsService, error) {
			return db.NewEventsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.UserService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.EventAnnotationsService, error) {
			return db.NewEventAnnotationsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.EventsService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.IncidentService, error) {
			return db.NewIncidentService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
				do.MustInvoke[rez.MessageService](i),
				do.MustInvoke[rez.UserService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.OncallRostersService, error) {
			return db.NewOncallRostersService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.SystemTopologyService, error) {
			return db.NewSystemTopologyService(
				do.MustInvoke[*ent.Client](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.OncallShiftsService, error) {
			return db.NewOncallShiftsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
				do.MustInvoke[rez.IntegrationsService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.OncallMetricsService, error) {
			return db.NewOncallMetricsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.OncallShiftsService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.DebriefService, error) {
			return db.NewDebriefService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.JobService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.RetrospectiveService, error) {
			return db.NewRetrospectiveService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.MessageService](i),
				do.MustInvoke[rez.IncidentService](i),
			)
		}),
		do.Lazy(func(i do.Injector) (rez.AlertService, error) {
			return db.NewAlertService(do.MustInvoke[*ent.Client](i))
		}),
		do.Lazy(func(i do.Injector) (rez.PlaybookService, error) {
			return db.NewPlaybookService(do.MustInvoke[*ent.Client](i))
		}),
		do.Lazy(func(i do.Injector) (rez.DocumentsService, error) {
			return db.NewDocumentsService(
				do.MustInvoke[*ent.Client](i),
				do.MustInvoke[rez.TeamService](i),
			)
		}),
	}
}
