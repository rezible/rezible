package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/rezible/rezible/internal/genkit"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	demoprovider "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/slack/slackagent"
	"github.com/rezible/rezible/internal/integrations/slack/slackincidents"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/pkg/integrations"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/pkg/projections"
)

type startable interface {
	Start(context.Context) error
}

func runServicesFor[Entrypoint startable](ctx context.Context, i do.Injector) error {
	if initErr := doRegistrations(i); initErr != nil {
		return fmt.Errorf("failed to initialize services: %w", initErr)
	}

	// invoke entrypoint service to load required service dependencies
	es, srvErr := do.Invoke[Entrypoint](i)
	if srvErr != nil {
		return fmt.Errorf("failed to initialize %T: %v", es, srvErr)
	}
	var services []startable
	for _, desc := range i.ListInvokedServices() {
		s, invErr := do.InvokeNamed[any](i, desc.Service)
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

func shutdownServices(baseCtx context.Context, i do.Injector) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), 5*time.Second)
	defer cancel()
	shutdown := i.ShutdownWithContext(ctx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

func doRegistrations(i do.Injector) error {
	intgReg := do.MustInvoke[*integrations.PackageRegistry](i)
	pipelineReg := do.MustInvoke[*projections.PipelineRegistry](i)

	for _, desc := range i.ListProvidedServices() {
		svc := desc.Service
		if strings.Contains(svc, "internal/integrations") {
			if pkg, ok := do.MustInvokeNamed[any](i, svc).(rez.IntegrationPackage); ok {
				if regErr := intgReg.RegisterPackage(pkg); regErr != nil {
					return fmt.Errorf("failed to register integration package: %w", regErr)
				}
				if procPkg, isEventProcessor := pkg.(rez.ProviderEventProcessor); isEventProcessor {
					pipelineReg.RegisterProviderEventProcessors(procPkg, pkg.Name())
				}
			}
		}
	}

	pipelineReg.RegisterEventProjector(do.MustInvoke[*db.KnowledgeService](i),
		projections.SubjectKindChatMessage,
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
	)
	pipelineReg.RegisterEventProjector(do.MustInvoke[*db.UserService](i), projections.SubjectKindUser)
	pipelineReg.RegisterEventProjector(do.MustInvoke[*db.IncidentService](i),
		projections.SubjectKindIncident,
		projections.SubjectKindIncidentImpact,
	)
	pipelineReg.RegisterEventProjector(do.MustInvoke[*db.AlertService](i), projections.SubjectKindAlert)
	pipelineReg.RegisterEventProjector(do.MustInvoke[*db.PlaybookService](i), projections.SubjectKindPlaybook)

	return nil
}

func declareServices(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (*projections.PipelineRegistry, error) {
		return projections.NewPipelineRegistry(), nil
	})

	do.Provide(i, func(i do.Injector) (*integrations.PackageRegistry, error) {
		return integrations.NewPackageRegistry(), nil
	})

	do.Provide(i, func(i do.Injector) (rez.AgentRegistry, error) {
		r := genkit.NewAgentRegistry(ctx, do.MustInvoke[rez.Config](i), do.MustInvoke[rez.AgentRunSnapshotService](i))
		genkit.RegisterWorkflowAgent(r, genkit.NewAlertInvestigationAgent(do.MustInvoke[rez.AlertService](i)))
		return r, nil
	})

	do.Provide(i, func(i do.Injector) (rez.MigrationService, error) {
		pgPool, poolErr := postgres.MakePgxPool(ctx, do.MustInvoke[rez.Config](i).Postgres, true)
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

	provideServices(i)
	provideIntegrations(i)

	do.Provide(i, func(i do.Injector) (oapiv1.Handler, error) {
		return apiv1.NewHandler(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AgentService](i),
			do.MustInvoke[rez.AlertService](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.DocumentsService](i),
			do.MustInvoke[rez.DebriefService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.EventsService](i),
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
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[*integrations.PackageRegistry](i).GetWebhookHandlers(),
		)
	})
}

var provideIntegrations = do.Package(
	do.Lazy(func(i do.Injector) (*demoprovider.Integration, error) {
		return demoprovider.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*github.Integration, error) {
		return github.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*google.Integration, error) {
		return google.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.EventsService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*slackagent.Integration, error) {
		app, appErr := slackagent.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.AgentService](i),
			do.MustInvoke[rez.EventsService](i),
		)
		if appErr != nil {
			return nil, fmt.Errorf("making slackagent app: %w", appErr)
		}
		return slackagent.MakeIntegration(
			app,
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*slackincidents.Integration, error) {
		app, appErr := slackincidents.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
		)
		if appErr != nil {
			return nil, appErr
		}
		return slackincidents.MakeIntegration(
			app,
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),
)

var provideServices = do.Package(
	do.Lazy(func(i do.Injector) (*db.ProviderEventPipelineService, error) {
		return db.NewProviderEventPipelineService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[*projections.PipelineRegistry](i),
		)
	}),
	do.Bind[*db.ProviderEventPipelineService, rez.ProviderEventPipelineService](),

	do.Lazy(func(i do.Injector) (*db.KnowledgeService, error) {
		return db.NewKnowledgeService(do.MustInvoke[rez.Database](i)), nil
	}),
	do.Bind[*db.KnowledgeService, rez.KnowledgeService](),

	do.Lazy(func(i do.Injector) (*db.IntegrationsService, error) {
		return db.NewIntegrationsService(
			do.MustInvoke[rez.Config](i).App,
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[*integrations.PackageRegistry](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),
	do.Bind[*db.IntegrationsService, rez.IntegrationService](),

	do.Lazy(func(i do.Injector) (*db.OrganizationService, error) {
		return db.NewOrganizationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		), nil
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

	do.Lazy(func(i do.Injector) (*db.EventService, error) {
		return db.NewEventService(
			do.MustInvoke[rez.Database](i),
		)
	}),
	do.Bind[*db.EventService, rez.EventsService](),

	do.Lazy(func(i do.Injector) (*db.AuthSessionService, error) {
		return db.NewAuthSessionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		), nil
	}),
	do.Bind[*db.AuthSessionService, rez.AuthSessionService](),

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
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.TeamService](i),
		)
	}),
	do.Bind[*db.DocumentsService, rez.DocumentsService](),

	do.Lazy(func(i do.Injector) (*db.AgentRunSnapshotService, error) {
		return db.NewAgentRunSnapshotService(do.MustInvoke[rez.Database](i))
	}),
	do.Bind[*db.AgentRunSnapshotService, rez.AgentRunSnapshotService](),

	do.Lazy(func(i do.Injector) (*db.AgentService, error) {
		return db.NewAgentService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.AgentRegistry](i),
		)
	}),
	do.Bind[*db.AgentService, rez.AgentService](),
)
