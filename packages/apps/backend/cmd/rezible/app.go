package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/integrations/projections"
	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/http/oidc"
	fakeprovider "github.com/rezible/rezible/internal/integrations/fake"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/slack/slackagent"
	"github.com/rezible/rezible/internal/integrations/slack/slackincidents"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type commandRunner struct {
	i do.Injector
}

func makeCommandRunner() *commandRunner {
	return &commandRunner{i: do.New()}
}

func (r *commandRunner) setupContext(ctx context.Context) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	cfg, cfgErr := koanf.LoadConfig(ctx, koanf.Options{LoadEnvironment: true})
	if cfgErr != nil {
		return nil, fmt.Errorf("load config: %w", cfgErr)
	}
	do.ProvideValue(r.i, *cfg)
	declareServices(ctx, r.i)

	return ctx, nil
}

func (r *commandRunner) printConfig() error {
	fmt.Printf("%+v\n", do.MustInvoke[rez.Config](r.i))
	return nil
}

func (r *commandRunner) runServer(ctx context.Context) error {
	return runService[*http.Server](ctx, r)
}

func (r *commandRunner) runSchemaMigration(ctx context.Context, direction string) error {
	ms := do.MustInvoke[*postgres.MigrationService](r.i)
	return ms.Run(ctx, direction)
}

func (r *commandRunner) createSchemaMigration(ctx context.Context, name string) error {
	ms := do.MustInvoke[*postgres.MigrationService](r.i)
	return ms.CreateSchemaMigration(ctx, name)
}

func (r *commandRunner) updateMigrationChecksumFile() error {
	return postgres.UpdateMigrationsChecksum()
}

func (r *commandRunner) printOpenApiSpec(asJson bool) error {
	spec, specErr := oapiv1.GetSpec(asJson)
	if specErr != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", specErr)
	}
	fmt.Printf("%s", spec)
	return nil
}

func (r *commandRunner) shutdown(baseCtx context.Context) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), 5*time.Second)
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

type startable interface {
	Start(context.Context) error
}

func runService[Entrypoint startable](ctx context.Context, r *commandRunner) error {
	if initErr := registerIntegrations(r.i); initErr != nil {
		return fmt.Errorf("failed to initialize services: %w", initErr)
	}
	// invoke entrypoint service to load required service dependencies
	es, srvErr := do.Invoke[Entrypoint](r.i)
	if srvErr != nil {
		return fmt.Errorf("failed to initialize %T: %v", es, srvErr)
	}
	var services []startable
	for _, desc := range r.i.ListInvokedServices() {
		s, invErr := do.InvokeNamed[any](r.i, desc.Service)
		if invErr != nil {
			return fmt.Errorf("failed to invoke: %v", invErr)
		}
		if intgSvc, ok := s.(rez.IntegrationPackage); ok {
			if available, _ := intgSvc.IsAvailable(); !available {
				continue
			}
		}
		if svc, ok := s.(startable); ok {
			services = append(services, svc)
		}
	}

	errChan := make(chan error)
	go func() {
		p := pool.New().
			WithErrors().
			WithContext(ctx).
			WithFirstError()
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

func registerIntegrations(i do.Injector) error {
	pes := do.MustInvoke[rez.ProviderEventService](i)

	pes.RegisterProjectionHandler(do.MustInvoke[*db.KnowledgeService](i),
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
	)
	pes.RegisterProjectionHandler(do.MustInvoke[*db.UserService](i), projections.SubjectKindUser)
	pes.RegisterProjectionHandler(do.MustInvoke[*db.IncidentService](i), projections.SubjectKindIncident)
	pes.RegisterProjectionHandler(do.MustInvoke[*db.AlertService](i), projections.SubjectKindAlert)

	intgReg := do.MustInvoke[*integrations.PackageRegistry](i)
	for _, desc := range i.ListProvidedServices() {
		svc := desc.Service
		if strings.Contains(svc, "internal/integrations") {
			if pkg, ok := do.MustInvokeNamed[any](i, svc).(rez.IntegrationPackage); ok {
				if regErr := intgReg.RegisterPackage(pkg); regErr != nil {
					return fmt.Errorf("failed to register integration package: %w", regErr)
				}
			}
		}
	}

	return nil
}

func declareServices(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (*postgres.MigrationService, error) {
		pgCfg := do.MustInvoke[rez.Config](i).Postgres
		pgPool, poolErr := postgres.MakePgxPool(ctx, pgCfg, true)
		if poolErr != nil {
			return nil, fmt.Errorf("making admin pgx pool: %w", poolErr)
		}
		return postgres.NewMigrationService(pgPool)
	})

	do.Provide(i, func(i do.Injector) (*postgres.PgxPool, error) {
		return postgres.MakePgxPool(ctx, do.MustInvoke[rez.Config](i).Postgres, false)
	})

	do.Provide(i, func(i do.Injector) (rez.Database, error) {
		return postgres.NewPgxPoolDatabaseClient(do.MustInvoke[*postgres.PgxPool](i))
	})

	do.Provide(i, func(i do.Injector) (rez.TelemetryService, error) {
		return opentelemetry.NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
	})

	do.Provide(i, func(i do.Injector) (rez.JobService, error) {
		return river.NewJobService(
			do.MustInvoke[*postgres.PgxPool](i),
			do.MustInvoke[rez.TelemetryService](i),
		)
	})

	do.Provide(i, func(i do.Injector) (rez.MessageService, error) {
		return watermill.NewMessageService(do.MustInvoke[rez.TelemetryService](i))
	})

	do.Provide(i, func(i do.Injector) (*integrations.PackageRegistry, error) {
		return integrations.NewPackageRegistry(do.MustInvoke[rez.TelemetryService](i)), nil
	})

	do.Provide(i, func(i do.Injector) (http.UserAuthSessionService, error) {
		return oidc.NewAuthSessionService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		)
	})

	provideServices(i)
	provideIntegrations(i)

	do.Provide(i, func(i do.Injector) (oapiv1.Handler, error) {
		return apiv1.NewHandler(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AlertService](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.DocumentsService](i),
			do.MustInvoke[rez.DebriefService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.ProviderEventService](i),
			do.MustInvoke[rez.EventAnnotationsService](i),
			do.MustInvoke[rez.OncallRostersService](i),
			do.MustInvoke[rez.OncallShiftsService](i),
			do.MustInvoke[rez.OncallMetricsService](i),
			do.MustInvoke[rez.PlaybookService](i),
			do.MustInvoke[rez.RetrospectiveService](i),
			do.MustInvoke[rez.SystemTopologyService](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) (*http.Server, error) {
		return http.NewServer(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[http.UserAuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[*integrations.PackageRegistry](i).GetWebhookHandlers(),
		)
	})
}

var provideIntegrations = do.Package(
	do.Lazy(func(i do.Injector) (*slackagent.Integration, error) {
		return slackagent.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.EventAnnotationsService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.ProviderEventService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (*slackincidents.Integration, error) {
		return slackincidents.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.EventAnnotationsService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.ProviderEventService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (*fakeprovider.Integration, error) {
		return fakeprovider.MakeIntegration(do.MustInvoke[rez.Config](i)), nil
	}),
	do.Lazy(func(i do.Injector) (*github.Integration, error) {
		return github.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.ProviderEventService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (*google.Integration, error) {
		return google.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.EventAnnotationsService](i),
		)
	}),
)

var provideServices = do.Package(
	do.Lazy(func(i do.Injector) (*db.KnowledgeService, error) {
		return db.NewKnowledgeService(do.MustInvoke[rez.Database](i)), nil
	}),
	do.Bind[*db.KnowledgeService, rez.KnowledgeService](),

	do.Lazy(func(i do.Injector) (*db.IntegrationsService, error) {
		return db.NewIntegrationsService(
			do.MustInvoke[rez.Config](i).App,
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[*integrations.PackageRegistry](i))
	}),
	do.Bind[*db.IntegrationsService, rez.IntegrationService](),

	do.Lazy(func(i do.Injector) (*db.ProviderEventService, error) {
		return db.NewProviderEventService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationService](i),
		)
	}),
	do.Bind[*db.ProviderEventService, rez.ProviderEventService](),

	do.Lazy(func(i do.Injector) (*db.OrganizationService, error) {
		return db.NewOrganizationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Bind[*db.OrganizationService, rez.OrganizationService](),

	do.Lazy(func(i do.Injector) (*db.UserService, error) {
		return db.NewUserService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.KnowledgeService](i),
		)
	}),
	do.Bind[*db.UserService, rez.UserService](),

	do.Lazy(func(i do.Injector) (*db.TeamService, error) {
		return db.NewTeamService(do.MustInvoke[rez.Database](i))
	}),
	do.Bind[*db.TeamService, rez.TeamService](),

	do.Lazy(func(i do.Injector) (*db.EventAnnotationsService, error) {
		return db.NewEventAnnotationsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.ProviderEventService](i),
		)
	}),
	do.Bind[*db.EventAnnotationsService, rez.EventAnnotationsService](),

	do.Lazy(func(i do.Injector) (*db.IncidentService, error) {
		return db.NewIncidentService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.KnowledgeService](i),
		)
	}),
	do.Bind[*db.IncidentService, rez.IncidentService](),

	do.Lazy(func(i do.Injector) (*db.OncallRostersService, error) {
		return db.NewOncallRostersService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Bind[*db.OncallRostersService, rez.OncallRostersService](),

	do.Lazy(func(i do.Injector) (*db.OncallShiftsService, error) {
		return db.NewOncallShiftsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationService](i),
		)
	}),
	do.Bind[*db.OncallShiftsService, rez.OncallShiftsService](),

	do.Lazy(func(i do.Injector) (*db.OncallMetricsService, error) {
		return db.NewOncallMetricsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OncallShiftsService](i),
		)
	}),
	do.Bind[*db.OncallMetricsService, rez.OncallMetricsService](),

	do.Lazy(func(i do.Injector) (*db.SystemTopologyService, error) {
		return db.NewSystemTopologyService(
			do.MustInvoke[rez.Database](i),
		)
	}),
	do.Bind[*db.SystemTopologyService, rez.SystemTopologyService](),

	do.Lazy(func(i do.Injector) (*db.DebriefService, error) {
		return db.NewDebriefService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Bind[*db.DebriefService, rez.DebriefService](),

	do.Lazy(func(i do.Injector) (*db.RetrospectiveService, error) {
		return db.NewRetrospectiveService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
		)
	}),
	do.Bind[*db.RetrospectiveService, rez.RetrospectiveService](),

	do.Lazy(func(i do.Injector) (*db.AlertService, error) {
		return db.NewAlertService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeService](i),
		)
	}),
	do.Bind[*db.AlertService, rez.AlertService](),

	do.Lazy(func(i do.Injector) (*db.PlaybookService, error) {
		return db.NewPlaybookService(do.MustInvoke[rez.Database](i))
	}),
	do.Bind[*db.PlaybookService, rez.PlaybookService](),

	do.Lazy(func(i do.Injector) (*db.DocumentsService, error) {
		return db.NewDocumentsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.TeamService](i),
		)
	}),
	do.Bind[*db.DocumentsService, rez.DocumentsService](),
)
