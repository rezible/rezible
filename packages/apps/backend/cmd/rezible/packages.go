package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/do/v2"

	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/rezible/rezible/pkg/openapi"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"

	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/evals"
	"github.com/rezible/rezible/internal/db/eventprojection"
	"github.com/rezible/rezible/internal/genkit"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/integrations/alertmanager"
	"github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/grafana"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/internal/integrations/slack/slackagent"
	"github.com/rezible/rezible/internal/integrations/slack/slackincidents"
	"github.com/rezible/rezible/internal/integrations/webhook"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
)

func applicationPackages(ctx context.Context) Package {
	return do.Package(
		withEnvironmentConfig(ctx),
		withPostgresDatabase(ctx),
		withGenkitAiRuntime(ctx),
	)
}

// AppProfile adjusts the application for one way of running it. It is a package of overrides, applied once when
// the application is built and before anything is resolved. An override that wraps a base service invokes its
// concrete type, which the base wiring binds to the interface being overridden.
type AppProfile = Package

// override replaces the provider of T.
func override[T any](provider Provider[T]) AppProfile {
	return func(i do.Injector) {
		do.Override(i, provider)
	}
}

// overrideValue replaces the provider of T with a value.
func overrideValue[T any](value T) AppProfile {
	return func(i do.Injector) {
		do.OverrideValue(i, value)
	}
}

// testDatabaseProfile runs against a fresh, migrated database that is dropped when the application shuts down.
var testDatabaseProfile = do.Package(
	do.Lazy(func(i do.Injector) (*pgtestdb.Database, error) {
		return pgtestdb.New(do.MustInvoke[rez.Config](i).Postgres)
	}),
	override(func(i do.Injector) (rez.PostgresConfig, error) {
		return do.MustInvoke[*pgtestdb.Database](i).Config(), nil
	}),
)

// noPeriodicJobsProfile schedules no periodic jobs, so only the work a run causes happens.
var noPeriodicJobsProfile = overrideValue[[]*jobs.PeriodicJob](nil)

var basePackages = do.Package(
	pkgClock,
	pkgRiver,
	pkgWatermill,
	pkgGenkit,
	pkgHttp,
	pkgIntegrations,
	pkgDatabase,
	pkgRedis,
	pkgJobs,
	pkgMessages,
	pkgEvals,
)

func withEnvironmentConfig(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.Config, error) {
			return koanf.LoadConfig(ctx, koanf.Options{
				LoadEnvironment: true,
			})
		}),
	)
}

func withPostgresDatabase(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (rez.PostgresConfig, error) {
			return do.MustInvoke[rez.Config](i).Postgres, nil
		}),

		do.Lazy(func(i do.Injector) (*postgres.ConnectionPool, error) {
			return postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), false)
		}),

		do.Lazy(func(i do.Injector) (rez.MigrationService, error) {
			mgPool, mgPoolErr := postgres.MakePgxPool(ctx, do.MustInvoke[rez.PostgresConfig](i), true)
			if mgPoolErr != nil {
				return nil, fmt.Errorf("admin pgx pool: %w", mgPoolErr)
			}
			return postgres.NewMigrationService(mgPool)
		}),

		do.LazyNamed("db_field_encryption_key", func(i do.Injector) ([]byte, error) {
			return do.MustInvoke[rez.PostgresConfig](i).FieldEncryptionKey()
		}),

		do.Lazy(func(i do.Injector) (rez.Database, error) {
			return postgres.NewPgxPoolDatabaseClient(
				do.MustInvoke[*postgres.ConnectionPool](i),
				do.MustInvokeNamed[[]byte](i, "db_field_encryption_key"),
			)
		}),
	)
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

var pkgClock = do.Package(
	do.Lazy(func(i do.Injector) (rez.Clock, error) { return systemClock{}, nil }),
)

var pkgRiver = do.Package(
	do.Lazy(func(i do.Injector) (*river.JobService, error) {
		return river.NewJobService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[*postgres.ConnectionPool](i).Pool,
		)
	}),
	do.Bind[*river.JobService, rez.JobService](),
	do.Bind[*river.JobService, jobs.Registrar](),
)

var pkgGenkit = do.Package(
	do.Lazy(func(i do.Injector) (rezai.AiClassifyAgentThreadResponseWorkflow, error) {
		builder := do.MustInvoke[*genkit.WorkflowBuilder](i)
		return builder.DefinePromptWorkflow(rezai.ClassifyAgentThreadResponseDefinition)
	}),

	do.Lazy(func(i do.Injector) (rezai.JudgeSituationCandidateWorkflow, error) {
		builder := do.MustInvoke[*genkit.WorkflowBuilder](i)
		return builder.DefinePromptWorkflow(rezai.JudgeSituationCandidateDefinition)
	}),

	do.Lazy(func(i do.Injector) (*genkit.WorkflowBuilder, error) {
		return genkit.NewWorkflowBuilder(
			do.MustInvoke[*genkit.AiRuntime](i),
			do.MustInvoke[rez.AiWorkflowRunner](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (*genkit.AgentCatalogue, error) {
		air := do.MustInvoke[*genkit.AiRuntime](i)
		investigationAgent := genkit.NewInvestigationAgent(
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.InvestigationService](i),
			do.MustInvoke[rez.InvestigationOutputService](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.ChangesService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.AgentProviderReadService](i),
			do.MustInvoke[rez.KnowledgeProposalService](i),
		)
		return genkit.NewAgentCatalogue(
			air,
			genkit.NewChatAgent(),
			investigationAgent,
		)
	}),
	do.Bind[*genkit.AgentCatalogue, rez.AiAgentCatalogue](),
)

func withGenkitAiRuntime(ctx context.Context) func(do.Injector) {
	return do.Package(
		do.Lazy(func(i do.Injector) (*genkit.AiRuntime, error) {
			svc := genkit.NewAiRuntime(do.MustInvoke[rez.Config](i))
			return svc, svc.Init(ctx)
		}),
	)
}

var pkgRedis = do.Package( /* TODO */ )

var pkgWatermill = do.Package(
	do.Lazy(func(i do.Injector) (messages.Transport, error) {
		return watermill.NewGoChannelTransport(), nil
	}),

	do.Lazy(func(i do.Injector) (*watermill.MessageQueue, error) {
		return watermill.NewMessageQueue(
			do.MustInvoke[rez.MessageQueueConfig](i),
			do.MustInvoke[messages.Transport](i),
			do.MustInvoke[*postgres.ConnectionPool](i).Pool,
		)
	}),
	do.Bind[*watermill.MessageQueue, rez.MessageQueue](),
	do.Bind[*watermill.MessageQueue, messages.Registrar](),
)

var pkgIntegrations = do.Package(
	do.Lazy(func(i do.Injector) (*integrations.ProviderHttpClient, error) {
		cfg := do.MustInvoke[rez.Config](i)
		return integrations.NewProviderHTTPClient(cfg.Integrations.AllowPrivateProviderUrls), nil
	}),

	do.Lazy(func(i do.Injector) (*demoprovider.Integration, error) {
		return demoprovider.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*alertmanager.Integration, error) {
		return alertmanager.MakeIntegration(
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*webhook.Integration, error) {
		return webhook.MakeIntegration(
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*grafana.Integration, error) {
		return grafana.MakeIntegration(
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[*integrations.ProviderHttpClient](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*github.Integration, error) {
		return github.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.ProviderEventPipelineService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*google.Integration, error) {
		return google.MakeIntegration(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
			do.MustInvoke[rez.IncidentDetailsService](i),
			do.MustInvoke[rez.EventsService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*slackintegration.AppServiceDependencies, error) {
		return &slackintegration.AppServiceDependencies{
			MessageQueue:                 do.MustInvoke[rez.MessageQueue](i),
			Installations:                do.MustInvoke[rez.IntegrationInstallationLookup](i),
			UserService:                  do.MustInvoke[rez.UserService](i),
			ProviderEventPipelineService: do.MustInvoke[rez.ProviderEventPipelineService](i),
		}, nil
	}),

	do.Lazy(func(i do.Injector) (*slackagent.App, error) {
		return slackagent.MakeApp(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.AiAgentSessionService](i),
			do.MustInvoke[rez.EventsService](i),
			do.MustInvoke[rezai.AiClassifyAgentThreadResponseWorkflow](i),
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
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.IncidentDetailsService](i),
			do.MustInvoke[rez.IncidentManagementService](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (*slackincidents.Integration, error) {
		app := do.MustInvoke[*slackincidents.App](i)
		deps := do.MustInvoke[*slackintegration.AppServiceDependencies](i)
		return app.MakeIntegration(deps)
	}),

	do.Lazy(func(i do.Injector) ([]rez.IntegrationDefinition, error) {
		return []rez.IntegrationDefinition{
			do.MustInvoke[*alertmanager.Integration](i),
			do.MustInvoke[*demoprovider.Integration](i),
			do.MustInvoke[*google.Integration](i),
			do.MustInvoke[*github.Integration](i),
			do.MustInvoke[*grafana.Integration](i),
			do.MustInvoke[*slackagent.Integration](i),
			do.MustInvoke[*slackincidents.Integration](i),
			do.MustInvoke[*webhook.Integration](i),
			do.MustInvoke[*evals.Integration](i),
		}, nil
	}),

	// Event processors have no dependencies, so the pipeline can be built before the integrations that use it.
	do.Lazy(func(do.Injector) (rez.ProviderEventProcessorRegistry, error) {
		return rez.ProviderEventProcessorRegistry{
			alertmanager.ProviderName:     alertmanager.EventProcessor{},
			demoprovider.ProviderName:     demoprovider.EventProcessor{},
			github.ProviderName:           github.EventProcessor{},
			grafana.ProviderName:          grafana.EventProcessor{},
			slackintegration.ProviderName: slackagent.EventProcessor{},
			webhook.ProviderName:          webhook.EventProcessor{},
			evals.EvaluationProviderName:  evals.EventProcessor{},
		}, nil
	}),

	do.Lazy(func(i do.Injector) (rez.EventProjectionService, error) {
		return eventprojection.NewProjectionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphIngestionService](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.IncidentDetailsService](i),
			do.MustInvoke[rez.AlertService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (*integrations.Registry, error) {
		intgDefs := do.MustInvoke[[]rez.IntegrationDefinition](i)
		return integrations.NewRegistry(intgDefs...)
	}),
)

var pkgDatabase = do.Package(
	do.Lazy(func(i do.Injector) (*db.AiWorkflowRunner, error) {
		return db.NewAiWorkflowRunner(), nil
	}),
	do.Bind[*db.AiWorkflowRunner, rez.AiWorkflowRunner](),

	do.Lazy(func(i do.Injector) (*db.ProviderEventIngester, error) {
		return db.NewProviderEventIngester(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.ProviderEventProcessorRegistry](i),
			do.MustInvoke[rez.EventProjectionService](i),
		), nil
	}),
	do.Bind[*db.ProviderEventIngester, rez.ProviderEventIngester](),

	do.Lazy(func(i do.Injector) (*db.ProviderEventPipelineService, error) {
		return db.NewProviderEventPipelineService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.ProviderEventIngester](i),
		)
	}),
	do.Bind[*db.ProviderEventPipelineService, rez.ProviderEventPipelineService](),

	do.Lazy(func(i do.Injector) (*db.IntegrationInstallationsService, error) {
		return db.NewIntegrationInstallationsService(do.MustInvoke[rez.Database](i))
	}),
	do.Bind[*db.IntegrationInstallationsService, rez.IntegrationInstallationLookup](),

	do.Lazy(func(i do.Injector) (*db.IntegrationsService, error) {
		return db.NewIntegrationsService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationInstallationLookup](i),
			do.MustInvoke[*integrations.Registry](i),
		)
	}),
	do.Bind[*db.IntegrationsService, rez.IntegrationService](),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.SyncIntegrationSourceEvents], error) {
		return db.NewIntegrationEventsSyncWorker(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[*integrations.Registry](i),
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
		return db.NewEventsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.SituationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.AuthSessionService, error) {
		return db.NewAuthSessionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.IncidentDetailsService, error) {
		return db.NewIncidentDetailsService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.SituationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.IncidentManagementService, error) {
		return db.NewIncidentManagementService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.RetrospectiveService](i),
			do.MustInvoke[rez.IncidentDetailsService](i),
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

	do.Lazy(func(i do.Injector) (rez.KnowledgeGraphQueryService, error) {
		return db.NewKnowledgeGraphQueryService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.KnowledgeEntityAliasService, error) {
		return db.NewKnowledgeEntityAliasService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[*integrations.Registry](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.ChangesService, error) {
		return db.NewChangesService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.KnowledgeGraphIngestionService, error) {
		return db.NewKnowledgeGraphIngestionService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.KnowledgeProposalService, error) {
		return db.NewKnowledgeProposalService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphIngestionService](i),
			do.MustInvoke[rez.KnowledgeEntityAliasService](i),
			do.MustInvoke[*integrations.Registry](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.SituationBriefService, error) {
		return db.NewSituationBriefService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
			do.MustInvoke[rez.ChangesService](i),
			do.MustInvoke[rez.KnowledgeProposalService](i),
			do.MustInvoke[rez.IntegrationService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.SystemAnalysisService, error) {
		return db.NewSystemAnalysisService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
		)
	}),

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
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Bind[*db.RetrospectiveService, rez.RetrospectiveService](),

	do.Lazy(func(i do.Injector) (rez.TaskService, error) {
		return db.NewTaskService(do.MustInvoke[rez.Database](i))
	}),

	do.Lazy(func(i do.Injector) (rez.DiscussionService, error) {
		return db.NewDiscussionService(do.MustInvoke[rez.Database](i)), nil
	}),

	do.Lazy(func(i do.Injector) (*db.AlertService, error) {
		return db.NewAlertService(
			do.MustInvoke[rez.Config](i).Alerts,
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.KnowledgeGraphIngestionService](i),
			do.MustInvoke[rez.SituationSignalService](i),
		)
	}),
	do.Bind[*db.AlertService, rez.AlertService](),

	do.Lazy(func(i do.Injector) (*db.SituationSignalService, error) {
		return db.NewSituationSignalService(do.MustInvoke[rez.JobService](i)), nil
	}),
	do.Bind[*db.SituationSignalService, rez.SituationSignalService](),

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

	do.Lazy(func(i do.Injector) (rez.AiAgentSessionService, error) {
		return db.NewAiAgentSessionService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageQueue](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.AgentProviderReadService, error) {
		return db.NewAgentProviderReadService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Clock](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.StartAgentSession], error) {
		return db.NewStartAgentSessionWorker(
			do.MustInvoke[rez.Config](i).AI,
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiAgentCatalogue](i),
			do.MustInvoke[rez.AiAgentSessionService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.InvokeAgentTurn], error) {
		return db.NewInvokeAgentTurnWorker(
			do.MustInvoke[rez.Config](i).AI,
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.AiAgentCatalogue](i),
			do.MustInvoke[rez.AiAgentSessionService](i),
		)
	}),

	do.Lazy(func(i do.Injector) (rez.SituationsConfig, error) {
		return do.MustInvoke[rez.Config](i).Situations, nil
	}),

	do.Lazy(func(i do.Injector) (*db.SituationService, error) {
		return db.NewSituationService(
			do.MustInvoke[rez.SituationsConfig](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.InvestigationService](i),
			do.MustInvoke[rez.SystemAnalysisService](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
			do.MustInvoke[rezai.JudgeSituationCandidateWorkflow](i),
			do.MustInvoke[*db.AlertService](i),
		)
	}),
	do.Bind[*db.SituationService, rez.SituationService](),
	do.Lazy(func(i do.Injector) (*db.InvestigationService, error) {
		return db.NewInvestigationService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiAgentSessionService](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.SystemAnalysisService](i),
		), nil
	}),
	do.Bind[*db.InvestigationService, rez.InvestigationService](),
	do.Bind[*db.InvestigationService, rez.InvestigationOutputService](),
	do.Lazy(func(i do.Injector) (*db.SituationNoticeService, error) {
		return db.NewSituationNoticeService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.SituationBriefService](i),
			do.MustInvoke[rez.InvestigationOutputService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.OrganizationService](i),
		), nil
	}),
	do.Bind[*db.SituationNoticeService, rez.SituationNoticeService](),
	do.Lazy(func(i do.Injector) (jobs.Worker[jobs.ReconcileInvestigation], error) {
		return db.NewReconcileInvestigationWorker(
			do.MustInvoke[*db.InvestigationService](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (rez.SystemHazardService, error) {
		return db.NewSystemHazardService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.KnowledgeGraphIngestionService](i),
		)
	}),
)

var pkgOpenApiV1 = do.Package(
	do.Lazy(func(i do.Injector) (rez.AppAuthSessionCookie, error) {
		return oapiv1.NewAppAuthSessionCookie(do.MustInvoke[rez.Config](i).App.FrontendApiPath), nil
	}),

	do.Lazy(func(i do.Injector) (oapiv1.SecurityProvider, error) {
		sp := apiv1.NewRequestSecurityProvider(
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[rez.AppAuthSessionCookie](i),
		)
		if do.MustInvoke[rez.Config](i).HttpServer.Auth.EnableDevSkipMode {
			return apiv1.NewDevelopmentSecurityProvider(sp, makeDevelopmentAuthSession())
		}
		return sp, nil
	}),

	do.Lazy(func(i do.Injector) (oapiv1.Handler, error) {
		return apiv1.NewHandler(
			do.MustInvoke[oapiv1.SecurityProvider](i),
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.AiAgentCatalogue](i),
			do.MustInvoke[rez.AiAgentSessionService](i),
			do.MustInvoke[rez.MessageQueue](i),
			do.MustInvoke[rez.AlertService](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.DocumentsService](i),
			do.MustInvoke[rez.DebriefService](i),
			do.MustInvoke[rez.IncidentDetailsService](i),
			do.MustInvoke[rez.IncidentManagementService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.InvestigationService](i),
			do.MustInvoke[rez.InvestigationOutputService](i),
			do.MustInvoke[rez.EventsService](i),
			do.MustInvoke[rez.OncallRostersService](i),
			do.MustInvoke[rez.OncallShiftsService](i),
			do.MustInvoke[rez.OncallMetricsService](i),
			do.MustInvoke[rez.PlaybookService](i),
			do.MustInvoke[rez.RetrospectiveService](i),
			do.MustInvoke[rez.TaskService](i),
			do.MustInvoke[rez.DiscussionService](i),
			do.MustInvoke[rez.SystemAnalysisService](i),
			do.MustInvoke[rez.KnowledgeGraphQueryService](i),
			do.MustInvoke[rez.KnowledgeEntityAliasService](i),
			do.MustInvoke[rez.KnowledgeProposalService](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.SituationBriefService](i),
			do.MustInvoke[rez.SituationNoticeService](i),
		), nil
	}),

	do.Lazy(func(i do.Injector) (oapiv1.API, error) {
		return oapiv1.MakeApi(do.MustInvoke[oapiv1.Handler](i)), nil
	}),

	do.Lazy(func(i do.Injector) (openapi.Adapter, error) {
		return do.MustInvoke[oapiv1.API](i).Adapter(), nil
	}),
)

var pkgHttp = do.Package(
	do.Lazy(func(i do.Injector) (*oidc.UserAuthProvider, error) {
		return oidc.NewUserAuthProvider(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[rez.AppAuthSessionCookie](i),
		)
	}),
	do.Bind[*oidc.UserAuthProvider, http.UserAuthProvider](),

	pkgOpenApiV1,

	do.Lazy(func(i do.Injector) (http.WebhookHandlers, error) {
		return do.MustInvoke[*integrations.Registry](i).GetAvailableWebhookHandlers(), nil
	}),

	do.Lazy(func(i do.Injector) (http.HealthCheckFunc, error) {
		return i.HealthCheckWithContext, nil
	}),

	do.Lazy(func(i do.Injector) (http.ReadyCheckFunc, error) {
		pool := do.MustInvoke[*postgres.ConnectionPool](i)
		return func(ctx context.Context) map[string]error {
			return map[string]error{
				"postgres.ConnectionPool": pool.Ping(ctx),
			}
		}, nil
	}),

	do.Lazy(func(i do.Injector) (*http.Server, error) {
		return http.NewServer(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[http.UserAuthProvider](i),
			do.MustInvoke[oapiv1.API](i),
			do.MustInvoke[http.WebhookHandlers](i),
			do.MustInvoke[http.HealthCheckFunc](i),
			do.MustInvoke[http.ReadyCheckFunc](i),
		)
	}),
)

type jobDefinitionProvider struct {
	i       do.Injector
	workers []jobs.WorkerDefinition
}

func newJobDefinitionProvider(i do.Injector) *jobDefinitionProvider {
	return &jobDefinitionProvider{i: i}
}

func (p *jobDefinitionProvider) add(workers ...jobs.WorkerDefinition) {
	p.workers = append(p.workers, workers...)
}

func (p *jobDefinitionProvider) addProvidedArgs[A jobs.JobArgs]() {
	p.add(jobs.DefineWorker[A](do.MustInvoke[jobs.Worker[A]](p.i)))
}

func (p *jobDefinitionProvider) addInvokedProviderFunc[S any](fn func(S) jobs.WorkerDefinition) {
	p.add(fn(do.MustInvoke[S](p.i)))
}

func (p *jobDefinitionProvider) addInvokedProvider[P jobs.WorkerProvider]() {
	p.add(do.MustInvoke[P](p.i).JobWorkers()...)
}

func (p *jobDefinitionProvider) getWorkerDefinitions() ([]jobs.WorkerDefinition, error) {
	return p.workers, nil
}

var pkgJobs = do.Package(
	do.Lazy(func(i do.Injector) ([]jobs.WorkerDefinition, error) {
		p := newJobDefinitionProvider(i)

		p.addProvidedArgs[jobs.ReconcileInvestigation]()
		p.addProvidedArgs[jobs.StartAgentSession]()
		p.addProvidedArgs[jobs.InvokeAgentTurn]()
		p.addProvidedArgs[jobs.SyncIntegrationSourceEvents]()

		// TODO: maybe convert these to jobs.WorkerProvider
		p.addInvokedProviderFunc(db.NewProcessProviderEventWorker)
		p.addInvokedProviderFunc(db.NewProjectNormalizedEventWorker)
		p.addInvokedProviderFunc(db.NewSendIncidentDebriefRequestsWorker)
		p.addInvokedProviderFunc(db.NewGenerateIncidentDebriefResponseWorker)
		p.addInvokedProviderFunc(db.NewGenerateIncidentDebriefSuggestionsWorker)
		p.addInvokedProviderFunc(db.NewScanOncallShiftsWorker)
		p.addInvokedProviderFunc(db.NewSettleAlertEpisodeWorker)
		p.addInvokedProviderFunc(db.NewProcessSituationSignalWorker)
		p.addInvokedProviderFunc(db.NewEvaluateSituationWorker)
		p.addInvokedProviderFunc(db.NewEnsureShiftHandoverSentWorker)
		p.addInvokedProviderFunc(db.NewEnsureShiftHandoverReminderSentWorker)
		p.addInvokedProviderFunc(db.NewGenerateShiftMetricsWorker)
		p.addInvokedProviderFunc(db.NewSyncScheduledIntegrationSourcesWorker)
		p.addInvokedProviderFunc(db.NewSyncSituationNoticesWorker)
		p.add(db.NewSendReviewRequestMessageWorker(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.IntegrationService](i),
		))

		for _, intgProv := range do.MustInvoke[*integrations.Registry](i).All[jobs.WorkerProvider]() {
			p.add(intgProv.JobWorkers()...)
		}

		return p.getWorkerDefinitions()
	}),

	do.Lazy(func(i do.Injector) ([]*jobs.PeriodicJob, error) {
		periodicJobs := []*jobs.PeriodicJob{
			jobs.MakePeriodicIntervalJob[jobs.SyncScheduledIntegrationSources](time.Hour, nil),
		}
		return periodicJobs, nil
	}),

	do.Lazy(func(i do.Injector) (jobs.Definition, error) {
		return jobs.Definition{
			Workers:      do.MustInvoke[[]jobs.WorkerDefinition](i),
			PeriodicJobs: do.MustInvoke[[]*jobs.PeriodicJob](i),
		}, nil
	}),
)

type messagesDefinitionProvider struct {
	i        do.Injector
	handlers []rez.MessageEventHandler
}

func newMessagesDefinitionProvider(i do.Injector) *messagesDefinitionProvider {
	return &messagesDefinitionProvider{i: i}
}

func (p *messagesDefinitionProvider) add(handlers ...rez.MessageEventHandler) {
	p.handlers = append(p.handlers, handlers...)
}

func (p *messagesDefinitionProvider) addInvokedProvider[P messages.MessageHandlerProvider]() {
	p.add(do.MustInvoke[P](p.i).MessageHandlers()...)
}

func (p *messagesDefinitionProvider) getDefinition() (messages.Definition, error) {
	return messages.Definition{Handlers: p.handlers}, nil
}

var pkgMessages = do.Package(
	do.Lazy(func(i do.Injector) (rez.MessageQueueConfig, error) {
		return do.MustInvoke[rez.Config](i).MessageQueue, nil
	}),

	do.Lazy(func(i do.Injector) (messages.Definition, error) {
		p := newMessagesDefinitionProvider(i)

		p.addInvokedProvider[*db.InvestigationService]()
		p.addInvokedProvider[*db.SituationNoticeService]()

		for _, intgProv := range do.MustInvoke[*integrations.Registry](i).All[messages.MessageHandlerProvider]() {
			p.add(intgProv.MessageHandlers()...)
		}

		return p.getDefinition()
	}),
)
