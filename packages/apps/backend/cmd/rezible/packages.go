package main

import (
	"context"
	"fmt"

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
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

type WithProvider func(context.Context) func(do.Injector)

func withEnvironmentConfig(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.Config, error) {
			return koanf.LoadConfig(ctx, koanf.Options{
				LoadEnvironment: true,
			})
		}),
	)
}

func withOpenTelemetry(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.TelemetryService, error) {
			return opentelemetry.NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
		}),
	)
}

func postgresTestDatabaseConfig(i do.Injector) (rez.PostgresConfig, error) {
	return do.MustInvoke[*pgtestdb.Database](i).Config(), nil
}

func withPostgresDatabase(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.PostgresConfig, error) {
			return do.MustInvoke[rez.Config](i).Postgres, nil
		}),

		do.Lazy(func(i do.Injector) (*pgtestdb.Database, error) {
			return pgtestdb.New(do.MustInvoke[rez.Config](i).Postgres)
		}),

		do.Lazy(func(i do.Injector) (*postgres.ConnectionPool, error) {
			return postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), false)
		}),

		do.Lazy(func(i do.Injector) (*postgres.MigrationService, error) {
			mgPool, mgPoolErr := postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), true)
			if mgPoolErr != nil {
				return nil, fmt.Errorf("admin pgx pool: %w", mgPoolErr)
			}
			return postgres.NewMigrationService(mgPool)
		}),

		do.Lazy(func(i do.Injector) (rez.Database, error) {
			return postgres.NewPgxPoolDatabaseClient(do.MustInvoke[*postgres.ConnectionPool](i))
		}),
	)
}

func withGenkitAiService(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.AiService, error) {
			svc := genkit.NewAiService(do.MustInvoke[rez.Config](i))
			opts := do.MustInvoke[[]genkit.AiServiceOption](i)
			return svc, svc.Init(ctx, opts...)
		}),
	)
}

var basePackages = do.Package(
	pkgRiver,
	pkgWatermill,
	pkgGenkit,
	pkgHttp,
	pkgIntegrations,
	pkgDatabase,
)

var pkgGenkit = do.Package(
	do.Lazy(func(i do.Injector) (rezai.ClassifyAgentThreadResponseWorkflowRunner, error) {
		return rezai.GetWorkflowRunner(do.MustInvoke[rez.AiService](i), rezai.ClassifyAgentThreadResponseWorkflow)
	}),

	do.Lazy(func(i do.Injector) ([]genkit.AiServiceOption, error) {
		analysisMw := genkit.WithSystemAnalysisAgentMiddleware(
			do.MustInvoke[rez.SystemAnalysisService](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
		situationSvc := do.MustInvoke[rez.SituationService](i)
		opts := []genkit.AiServiceOption{
			genkit.WithAgent(genkit.NewChatAgent()),
			genkit.WithAgent(genkit.NewInvestigationAgent(situationSvc), analysisMw),
			genkit.WithWorkflow(rezai.ClassifyAgentThreadResponseWorkflow),
		}
		return opts, nil
	}),

	do.Lazy(func(i do.Injector) (*genkit.EvaluationService, error) {
		return genkit.NewEvaluationService(do.MustInvoke[rez.Database](i), do.MustInvoke[*genkit.AiService](i)), nil
	}),
	do.Bind[*genkit.EvaluationService, rezai.EvalScenarioRunner](),
)

type JobsRegistrar interface {
	RegisterWorkers(jobs.Definition) error
}

var pkgRiver = do.Package(
	do.Lazy(func(i do.Injector) (*river.JobService, error) {
		return river.NewJobService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[*postgres.ConnectionPool](i).Pool,
			do.MustInvoke[rez.TelemetryService](i),
		)
	}),
	do.Bind[*river.JobService, rez.JobService](),
	do.Bind[*river.JobService, JobsRegistrar](),
)

type MessageHandlerRegistrar interface {
	AddHandlers(...rez.MessageEventHandler) error
}

var pkgWatermill = do.Package(
	do.Lazy(func(i do.Injector) (watermill.Transport, error) {
		return nil, nil
	}),
	do.Lazy(func(i do.Injector) (*watermill.MessageService, error) {
		return watermill.NewMessageService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[watermill.Transport](i),
		)
	}),
	do.Bind[*watermill.MessageService, rez.MessageService](),
	do.Bind[*watermill.MessageService, MessageHandlerRegistrar](),
)

var pkgIntegrations = do.Package(
	do.Lazy(func(i do.Injector) (rez.IntegrationRegistry, error) {
		return integrations.NewRegistry(), nil
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
			do.MustInvoke[rez.AlertService](i),
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
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.EventsService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*slackintegration.AppServiceDependencies, error) {
		return slackintegration.NewServiceDependencies(
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*slackagent.App, error) {
		return slackagent.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.AgentSessionService](i),
			do.MustInvoke[rez.EventsService](i),
			do.MustInvoke[rezai.ClassifyAgentThreadResponseWorkflowRunner](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (*slackagent.Integration, error) {
		app := do.MustInvoke[*slackagent.App](i)
		deps := do.MustInvoke[*slackintegration.AppServiceDependencies](i)
		return app.MakeIntegration(deps)
	}),

	do.Lazy(func(i do.Injector) (*slackincidents.App, error) {
		return slackincidents.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IncidentService](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (*slackincidents.Integration, error) {
		app := do.MustInvoke[*slackincidents.App](i)
		deps := do.MustInvoke[*slackintegration.AppServiceDependencies](i)
		return app.MakeIntegration(deps)
	}),

	do.Lazy(func(i do.Injector) ([]rez.IntegrationDefinition, error) {
		return []rez.IntegrationDefinition{
			do.MustInvoke[*demoprovider.Integration](i),
			do.MustInvoke[*google.Integration](i),
			do.MustInvoke[*github.Integration](i),
			do.MustInvoke[*slackagent.Integration](i),
			do.MustInvoke[*slackincidents.Integration](i),
		}, nil
	}),
)

var pkgDatabase = do.Package(
	do.Lazy(func(i do.Injector) (*db.ProviderEventPipelineService, error) {
		return db.NewProviderEventPipelineService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.ProviderEventProcessorRegistry](i),
			do.MustInvoke[rez.EventProjectionService](i),
		)
	}),
	do.Bind[*db.ProviderEventPipelineService, rez.ProviderEventPipelineService](),

	do.Lazy(func(i do.Injector) (rez.IntegrationService, error) {
		return db.NewIntegrationsService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationRegistry](i),
		)
	}),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.SyncIntegrationSourceEvents], error) {
		return db.NewIntegrationEventsSyncWorker(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.IntegrationRegistry](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
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
			do.MustInvoke[rez.SituationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.OncallRostersService, error) {
		return db.NewOncallRostersService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),

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

	do.Lazy(func(i do.Injector) (rez.KnowledgeGraphService, error) {
		return db.NewKnowledgeGraphService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.SystemAnalysisService, error) {
		return db.NewSystemAnalysisService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*db.DebriefService, error) {
		return db.NewDebriefService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Bind[*db.DebriefService, rez.DebriefService](),

	do.Lazy(func(i do.Injector) (rez.RetrospectiveService, error) {
		return db.NewRetrospectiveService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.IncidentService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.AlertService, error) {
		return db.NewAlertService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.CloseInactiveAlertEpisodes], error) {
		return db.NewCloseInactiveAlertEpisodesWorker(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.SituationService](i),
		)
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
			do.MustInvoke[rez.MessageService](i),
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

	do.Lazy(func(i do.Injector) (*db.SituationService, error) {
		return db.NewSituationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.AgentSessionService](i),
			do.MustInvoke[rez.KnowledgeGraphService](i),
		)
	}),
	do.Bind[*db.SituationService, rez.SituationService](),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.ReconcileSituationInvestigation], error) {
		return db.NewReconcileSituationInvestigationWorker(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.AgentSessionService](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (rez.SystemHazardService, error) {
		return db.NewSystemHazardService(do.MustInvoke[rez.Database](i), do.MustInvoke[rez.KnowledgeGraphService](i))
	}),
)

var pkgHttp = do.Package(
	do.Lazy(func(i do.Injector) (oapiv1.Handler, error) {
		return apiv1.NewHandler(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiService](i),
			do.MustInvoke[rez.AgentSessionService](i),
			do.MustInvoke[rez.MessageService](i),
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
			do.MustInvoke[rez.SituationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*http.Server, error) {
		return http.NewServer(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[rez.IntegrationRegistry](i).GetAvailableWebhookHandlers(),
			i.HealthCheckWithContext,
		)
	}),
)

func jobWorkerProvider[A jobs.JobArgs]() do.Provider[jobs.WorkerDefinition] {
	return func(i do.Injector) (jobs.WorkerDefinition, error) {
		return jobs.DefineWorker[A](do.MustInvoke[jobs.Worker[A]](i)), nil
	}
}

func jobWorkerFuncProvider[S any](fn func(S) jobs.WorkerDefinition) do.Provider[jobs.WorkerDefinition] {
	return func(i do.Injector) (jobs.WorkerDefinition, error) {
		return fn(do.MustInvoke[S](i)), nil
	}
}

func makeDefaultJobWorkerProviders() []do.Provider[jobs.WorkerDefinition] {
	return []do.Provider[jobs.WorkerDefinition]{
		jobWorkerProvider[jobs.ReconcileSituationInvestigation](),
		jobWorkerProvider[jobs.CloseInactiveAlertEpisodes](),
		jobWorkerProvider[jobs.StartAgentSession](),
		jobWorkerProvider[jobs.InvokeAgentTurn](),
		jobWorkerProvider[jobs.SyncIntegrationSourceEvents](),

		// TODO: maybe convert these to regular workers?
		jobWorkerFuncProvider(db.NewProcessProviderEventWorker),
		jobWorkerFuncProvider(db.NewProjectNormalizedEventWorker),
		jobWorkerFuncProvider(db.NewSendIncidentDebriefRequestsWorker),
		jobWorkerFuncProvider(db.NewGenerateIncidentDebriefResponseWorker),
		jobWorkerFuncProvider(db.NewGenerateIncidentDebriefSuggestionsWorker),
		jobWorkerFuncProvider(db.NewScanOncallShiftsWorker),
		jobWorkerFuncProvider(db.NewEnsureShiftHandoverSentWorker),
		jobWorkerFuncProvider(db.NewEnsureShiftHandoverReminderSentWorker),
		jobWorkerFuncProvider(db.NewGenerateShiftMetricsWorker),

		jobWorkerFuncProvider(slackagent.NewSendMessageWorker),
		jobWorkerFuncProvider(slackagent.NewHandleBoundAgentThreadMessagedWorker),

		jobWorkerFuncProvider(slackincidents.NewCreateIncidentChannelWorker),
		jobWorkerFuncProvider(slackincidents.NewSendIncidentMilestoneMessageWorker),
	}
}
