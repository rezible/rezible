package main

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/samber/do/v2"

	rez "github.com/rezible/rezible"
	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/eventprojection"
	"github.com/rezible/rezible/internal/genkit"
	"github.com/rezible/rezible/internal/http"
	demoprovider "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/internal/integrations/slack/slackagent"
	"github.com/rezible/rezible/internal/integrations/slack/slackincidents"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

func makePackageInjector() do.Injector {
	opts := &do.InjectorOpts{}
	return do.NewWithOpts(opts)
}

func initPackages(ctx context.Context, i do.Injector) error {
	makePackageProvider(ctx)(i)
	return nil
}

type Provider = func(do.Injector)

func makePackageProvider(ctx context.Context) Provider {
	return do.Package(
		makeConfigProvider(ctx),
		makeOpenTelemetryProvider(ctx),
		makePostgresProvider(ctx),
		makeGenkitProvider(ctx),
		provideWatermillMessageService,
		provideDatabaseServices,
		provideIntegrations,
		provideJobWorkers,
		provideHttpApiServer,
	)
}

func makeConfigProvider(ctx context.Context) Provider {
	return do.Lazy(func(i do.Injector) (rez.Config, error) {
		return koanf.LoadConfig(ctx, koanf.Options{LoadEnvironment: true})
	})
}

func makeOpenTelemetryProvider(ctx context.Context) Provider {
	return do.Lazy(func(i do.Injector) (rez.TelemetryService, error) {
		return opentelemetry.NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
	})
}

func makePostgresProvider(ctx context.Context) Provider {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.PostgresConfig, error) {
			return do.MustInvoke[rez.Config](i).Postgres, nil
		}),

		do.Lazy(func(i do.Injector) (rez.MigrationService, error) {
			mgPool, mgPoolErr := postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), true)
			if mgPoolErr != nil {
				return nil, fmt.Errorf("admin pgx pool: %w", mgPoolErr)
			}
			return postgres.NewMigrationService(mgPool)
		}),

		do.Lazy(func(i do.Injector) (*postgres.ConnectionPool, error) {
			return postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), false)
		}),

		do.Lazy(func(i do.Injector) (rez.Database, error) {
			return postgres.NewPgxPoolDatabaseClient(do.MustInvoke[*postgres.ConnectionPool](i))
		}),

		do.Lazy(func(i do.Injector) (rez.JobService, error) {
			return river.NewJobService(
				do.MustInvoke[rez.Config](i),
				do.MustInvoke[*postgres.ConnectionPool](i),
				do.MustInvoke[rez.TelemetryService](i),
			)
		}),
	)
}

func makeGenkitProvider(ctx context.Context) Provider {
	initService := func(initCtx context.Context, i do.Injector, svc *genkit.AiService) error {
		intgToolsMw := genkit.WithIntegrationToolsMiddleware(do.MustInvoke[rez.IntegrationService](i))
		return svc.Init(initCtx,
			genkit.WithAgent(genkit.NewChatAgent(), intgToolsMw),
			genkit.WithAgent(genkit.NewAlertsAgent(do.MustInvoke[rez.AlertService](i)), intgToolsMw),
			genkit.WithWorkflow(rezai.ClassifyAgentThreadResponseWorkflow),
		)
	}
	return do.Package(
		do.Lazy(func(i do.Injector) (*genkit.AiService, error) {
			svc := genkit.NewAiService(do.MustInvoke[rez.Config](i))
			return svc, initService(ctx, i, svc)
		}),
		do.Bind[*genkit.AiService, rez.AiService](),

		do.Lazy(func(i do.Injector) (rezai.ClassifyAgentThreadResponseWorkflowRunner, error) {
			return rezai.ClassifyAgentThreadResponseWorkflow.GetRunner(do.MustInvoke[rez.AiService](i))
		}),

		do.Lazy(func(i do.Injector) (*genkit.DevServer, error) {
			svc := genkit.NewAiService(do.MustInvoke[rez.Config](i))
			devCtx := execution.NewTenantContext(ctx, 1)
			return svc.MakeDevServer(), initService(devCtx, i, svc)
		}),
	)
}

var provideWatermillMessageService = do.Package(
	do.Lazy(func(i do.Injector) (watermill.Transport, error) {
		return nil, nil
	}),
	do.Lazy(func(i do.Injector) (rez.MessageService, error) {
		return watermill.NewMessageService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[watermill.Transport](i),
		)
	}),
)

var provideIntegrations = do.Package(
	do.Lazy(func(i do.Injector) (rez.IntegrationPackageRegistry, error) {
		return integrations.NewPackageRegistry(), nil
	}),

	do.Lazy(func(i do.Injector) (rez.ProviderEventProcessorRegistry, error) {
		return rez.ProviderEventProcessorRegistry{}, nil
	}),

	do.Lazy(func(i do.Injector) (rez.EventProjectionService, error) {
		return eventprojection.NewProjectionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),

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
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.AgentSessionService](i),
			do.MustInvoke[rez.EventsService](i),
			do.MustInvoke[rezai.ClassifyAgentThreadResponseWorkflowRunner](i),
		)
		if appErr != nil {
			return nil, fmt.Errorf("making slackagent app: %w", appErr)
		}
		svc, svcErr := slackintegration.NewAppService(app,
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i))
		if svcErr != nil {
			return nil, fmt.Errorf("making slackagent app service: %w", svcErr)
		}
		return slackagent.MakeIntegration(svc), nil
	}),

	do.Lazy(func(i do.Injector) (*slackincidents.Integration, error) {
		app, appErr := slackincidents.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IncidentService](i),
		)
		if appErr != nil {
			return nil, fmt.Errorf("making slackincidents app: %w", appErr)
		}
		svc, svcErr := slackintegration.NewAppService(app,
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i))
		if svcErr != nil {
			return nil, fmt.Errorf("making slackincidents app service: %w", svcErr)
		}
		return slackincidents.MakeIntegration(svc), nil
	}),
)

var provideDatabaseServices = do.Package(
	do.Lazy(func(i do.Injector) (rez.ProviderEventPipelineService, error) {
		return db.NewProviderEventPipelineService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.ProviderEventProcessorRegistry](i),
			do.MustInvoke[rez.EventProjectionService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.IntegrationService, error) {
		return db.NewIntegrationsService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationPackageRegistry](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.OrganizationService, error) {
		return db.NewOrganizationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.UserService, error) {
		return db.NewUserService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OrganizationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.TeamService, error) {
		return db.NewTeamService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.EventsService, error) {
		return db.NewEventsService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.AuthSessionService, error) {
		return db.NewAuthSessionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.IncidentService, error) {
		return db.NewIncidentService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.OncallRostersService, error) {
		return db.NewOncallRostersService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.OncallShiftsService, error) {
		return db.NewOncallShiftsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.OncallMetricsService, error) {
		return db.NewOncallMetricsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OncallShiftsService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.KnowledgeGraphService, error) {
		return db.NewKnowledgeGraphService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.SystemAnalysisService, error) {
		return db.NewSystemAnalysisService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.DebriefService, error) {
		return db.NewDebriefService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.RetrospectiveService, error) {
		return db.NewRetrospectiveService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.AlertService, error) {
		return db.NewAlertService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.PlaybookService, error) {
		return db.NewPlaybookService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.DocumentsService, error) {
		return db.NewDocumentsService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.TeamService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.AgentSessionService, error) {
		return db.NewAgentSessionService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.InvestigationService, error) {
		return db.NewInvestigationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.AlertService](i),
			do.MustInvoke[rez.AgentSessionService](i),
		)
	}),
)

var provideJobWorkers = do.Package(
	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.InvokeAgentTurn], error) {
		return db.NewInvokeAgentTurnWorker(
			do.MustInvoke[rez.Config](i).AI,
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.AiService](i),
			do.MustInvoke[rez.AgentSessionService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.StartAgentSession], error) {
		return db.NewStartAgentSessionWorker(
			do.MustInvoke[rez.Config](i).AI,
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiService](i),
			do.MustInvoke[rez.AgentSessionService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.SyncIntegrationSourceEvents], error) {
		return db.NewIntegrationEventsSyncWorker(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.IntegrationPackageRegistry](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),
)

var provideHttpApiServer = do.Package(
	do.Lazy(func(i do.Injector) (oapiv1.Handler, error) {
		return apiv1.NewHandler(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiService](i),
			do.MustInvoke[rez.AgentSessionService](i),
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
			do.MustInvoke[rez.SystemAnalysisService](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (http.WebhookHandlers, error) {
		reg := do.MustInvoke[rez.IntegrationPackageRegistry](i)
		return reg.GetAvailableWebhookHandlers(), nil
	}),

	do.Lazy(func(i do.Injector) (*http.Server, error) {
		return http.NewServer(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[http.WebhookHandlers](i),
		)
	}),
)

//var provideRootServices = do.Package(
//	do.Bind[*db.ProviderEventPipelineService, rez.ProviderEventPipelineService](),
//	do.Bind[*db.IntegrationsService, rez.IntegrationService](),
//	do.Bind[*db.OrganizationService, rez.OrganizationService](),
//	do.Bind[*db.UserService, rez.UserService](),
//	do.Bind[*db.TeamService, rez.TeamService](),
//	do.Bind[*db.EventService, rez.EventsService](),
//	do.Bind[*db.AuthSessionService, rez.AuthSessionService](),
//	do.Bind[*db.IncidentService, rez.IncidentService](),
//	do.Bind[*db.OncallRostersService, rez.OncallRostersService](),
//	do.Bind[*db.OncallShiftsService, rez.OncallShiftsService](),
//	do.Bind[*db.OncallMetricsService, rez.OncallMetricsService](),
//	do.Bind[*db.KnowledgeGraphService, rez.KnowledgeGraphService](),
//	do.Bind[*db.DebriefService, rez.DebriefService](),
//	do.Bind[*db.RetrospectiveService, rez.RetrospectiveService](),
//	do.Bind[*db.AlertService, rez.AlertService](),
//	do.Bind[*db.PlaybookService, rez.PlaybookService](),
//	do.Bind[*db.DocumentsService, rez.DocumentsService](),
//	do.Bind[*db.AgentSessionService, rez.AgentSessionService](),
//	do.Bind[*db.InvestigationService, rez.InvestigationService](),
//)
