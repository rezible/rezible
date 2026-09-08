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

func makePackageInjector() do.Injector {
	return do.NewWithOpts(&do.InjectorOpts{})
}

func with[T any](i do.Injector, fn func(T) error) error {
	t, invErr := do.Invoke[T](i)
	if invErr != nil {
		return fmt.Errorf("failed to invoke %T: %w", t, invErr)
	}
	return fn(t)
}

func useEnvironmentConfig(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (rez.Config, error) {
		return koanf.LoadConfig(ctx, koanf.Options{
			LoadEnvironment: true,
		})
	})
}

func useOpenTelemetry(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (rez.TelemetryService, error) {
		return opentelemetry.NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
	})
}

func usePostgresTestDatabase(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*pgtestdb.Database, error) {
		return pgtestdb.New(do.MustInvoke[rez.Config](i).Postgres)
	})

	do.Override(i, func(i do.Injector) (rez.PostgresConfig, error) {
		return do.MustInvoke[*pgtestdb.Database](i).Config(), nil
	})
}

func usePostgresDatabase(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (rez.PostgresConfig, error) {
		return do.MustInvoke[rez.Config](i).Postgres, nil
	})

	do.Provide(i, func(i do.Injector) (*postgres.ConnectionPool, error) {
		return postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), false)
	})

	do.Provide(i, func(i do.Injector) (*postgres.MigrationService, error) {
		mgPool, mgPoolErr := postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), true)
		if mgPoolErr != nil {
			return nil, fmt.Errorf("admin pgx pool: %w", mgPoolErr)
		}
		return postgres.NewMigrationService(mgPool)
	})

	do.Provide(i, func(i do.Injector) (rez.Database, error) {
		return postgres.NewPgxPoolDatabaseClient(do.MustInvoke[*postgres.ConnectionPool](i))
	})
}

func useGenkitAiService(ctx context.Context, i do.Injector) {
	do.Provide(i, func(i do.Injector) (rez.AiService, error) {
		svc := genkit.NewAiService(do.MustInvoke[rez.Config](i))
		opts := do.MustInvoke[[]genkit.AiServiceOption](i)
		return svc, svc.Init(ctx, opts...)
	})
}

var useBasePackages = do.Package(
	useRegistries,
	useRiverJobService,
	useWatermillMessageService,
	useGenkit,
	useDatabaseServices,
	useIntegrations,
	useDefaultMessageHandlers,
	useJobDefinitions,
)

var useRegistries = do.Package(
	do.Lazy(func(i do.Injector) (rez.IntegrationRegistry, error) {
		return integrations.NewRegistry(), nil
	}),

	do.Lazy(func(i do.Injector) (rez.ProviderEventProcessorRegistry, error) {
		return rez.ProviderEventProcessorRegistry{}, nil
	}),
)

var useGenkit = do.Package(
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

	do.Lazy(func(i do.Injector) (*genkit.DevServer, error) {
		return genkit.NewDevServer(do.MustInvoke[*genkit.AiService](i), do.MustInvoke[*genkit.EvaluationService](i))
	}),
)

type jobsRegistrar interface {
	Init(jobs.Definition) error
}

var useRiverJobService = do.Package(
	do.Lazy(func(i do.Injector) (*river.JobService, error) {
		return river.NewJobService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[*postgres.ConnectionPool](i),
			do.MustInvoke[rez.TelemetryService](i),
		)
	}),
	do.Bind[*river.JobService, rez.JobService](),
	do.Bind[*river.JobService, jobsRegistrar](),
)

type messageHandlerRegistrar interface {
	AddHandlers(...rez.MessageEventHandler) error
}

var useWatermillMessageService = do.Package(
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
	do.Bind[*watermill.MessageService, messageHandlerRegistrar](),
)

var useIntegrations = do.Package(
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
		var defs []rez.IntegrationDefinition
		cfg := do.MustInvoke[rez.Config](i)
		if cfg.App.DebugMode {
			defs = append(defs, do.MustInvoke[*demoprovider.Integration](i))
		}
		if cfg.Integrations.Google.Enabled {
			defs = append(defs, do.MustInvoke[*google.Integration](i))
		}
		if cfg.Integrations.Github.Enabled {
			defs = append(defs, do.MustInvoke[*github.Integration](i))
		}
		if cfg.Integrations.Slack.Agent.Enabled {
			defs = append(defs, do.MustInvoke[*slackagent.Integration](i))
		}
		if cfg.Integrations.Slack.Incidents.Enabled {
			defs = append(defs, do.MustInvoke[*slackincidents.Integration](i))
		}
		return defs, nil
	}),
)

var useDatabaseServices = do.Package(
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

func provideLifecycleServicesWith[T rez.LifecycleService](i do.Injector) ([]rez.LifecycleService, error) {
	svcs := []rez.LifecycleService{
		do.MustInvoke[*river.JobService](i),
		do.MustInvoke[*watermill.MessageService](i),
	}
	for _, intg := range do.MustInvoke[[]rez.IntegrationDefinition](i) {
		if ls, ok := intg.(rez.LifecycleService); ok {
			svcs = append(svcs, ls)
		}
	}
	svcs = append(svcs, do.MustInvoke[T](i))
	return svcs, nil
}

var useHttpServer = do.Package(
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

	do.Lazy(func(i do.Injector) (http.WebhookHandlers, error) {
		reg := do.MustInvoke[rez.IntegrationRegistry](i)
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

	do.Lazy(func(i do.Injector) ([]rez.LifecycleService, error) {
		return provideLifecycleServicesWith[*http.Server](i)
	}),
)

var useGenkitDevServer = do.Package(
	do.Lazy(func(i do.Injector) ([]rez.LifecycleService, error) {
		return provideLifecycleServicesWith[*genkit.DevServer](i)
	}),
)

var useDefaultMessageHandlers = do.Lazy(func(i do.Injector) ([]rez.MessageEventHandler, error) {
	type ProvidesMessageHandlers interface {
		GetMessageHandlers() []rez.MessageEventHandler
	}
	handlers := do.MustInvoke[*db.SituationService](i).GetMessageHandlers()
	for _, pkg := range do.MustInvoke[[]rez.IntegrationDefinition](i) {
		if hp, ok := pkg.(ProvidesMessageHandlers); ok {
			handlers = append(handlers, hp.GetMessageHandlers()...)
		}
	}
	return handlers, nil
})

func makeJobWorkerProvider[A jobs.JobArgs]() do.Provider[jobs.WorkerDefinition] {
	return func(i do.Injector) (jobs.WorkerDefinition, error) {
		return jobs.DefineWorker[A](do.MustInvoke[jobs.Worker[A]](i)), nil
	}
}

func makeServiceFuncJobWorkerProvider[S any](fn func(S) jobs.WorkerDefinition) do.Provider[jobs.WorkerDefinition] {
	return func(i do.Injector) (jobs.WorkerDefinition, error) {
		return fn(do.MustInvoke[S](i)), nil
	}
}

var defaultJobWorkerProviders = []do.Provider[jobs.WorkerDefinition]{
	makeJobWorkerProvider[jobs.ReconcileSituationInvestigation](),
	makeJobWorkerProvider[jobs.CloseInactiveAlertEpisodes](),
	makeJobWorkerProvider[jobs.StartAgentSession](),
	makeJobWorkerProvider[jobs.InvokeAgentTurn](),
	makeJobWorkerProvider[jobs.SyncIntegrationSourceEvents](),

	// TODO: convert these to regular workers
	makeServiceFuncJobWorkerProvider(db.NewProcessProviderEventWorker),
	makeServiceFuncJobWorkerProvider(db.NewProjectNormalizedEventWorker),
	makeServiceFuncJobWorkerProvider(db.NewSendIncidentDebriefRequestsWorker),
	makeServiceFuncJobWorkerProvider(db.NewGenerateIncidentDebriefResponseWorker),
	makeServiceFuncJobWorkerProvider(db.NewGenerateIncidentDebriefSuggestionsWorker),
	makeServiceFuncJobWorkerProvider(db.NewScanOncallShiftsWorker),
	makeServiceFuncJobWorkerProvider(db.NewEnsureShiftHandoverSentWorker),
	makeServiceFuncJobWorkerProvider(db.NewEnsureShiftHandoverReminderSentWorker),
	makeServiceFuncJobWorkerProvider(db.NewGenerateShiftMetricsWorker),
	makeServiceFuncJobWorkerProvider(slackagent.NewSendMessageWorker),
	makeServiceFuncJobWorkerProvider(slackagent.NewHandleBoundAgentThreadMessagedWorker),
	makeServiceFuncJobWorkerProvider(slackincidents.NewCreateIncidentChannelWorker),
	makeServiceFuncJobWorkerProvider(slackincidents.NewSendIncidentMilestoneMessageWorker),
}

var useJobDefinitions = do.Package(
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

	do.Lazy(func(i do.Injector) ([]jobs.WorkerDefinition, error) {
		defs := make([]jobs.WorkerDefinition, len(defaultJobWorkerProviders))
		for idx, provider := range defaultJobWorkerProviders {
			def, provideErr := provider(i)
			if provideErr != nil {
				return nil, fmt.Errorf("provide job worker: %w", provideErr)
			}
			defs[idx] = def
		}
		return defs, nil
	}),
	do.Lazy(func(i do.Injector) ([]*jobs.PeriodicJob, error) {
		return []*jobs.PeriodicJob{
			jobs.CloseInactiveAlertEpisodesPeriodicJob,
		}, nil
	}),
	do.Lazy(func(i do.Injector) (jobs.Definition, error) {
		return jobs.Definition{
			Workers:      do.MustInvoke[[]jobs.WorkerDefinition](i),
			PeriodicJobs: do.MustInvoke[[]*jobs.PeriodicJob](i),
		}, nil
	}),
)
