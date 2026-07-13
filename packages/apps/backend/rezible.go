package rez

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/texm/prosemirror-go"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
)

var (
	ErrTenantContextMissing = fmt.Errorf("tenant access context not set")
	ErrInvalidUser          = fmt.Errorf("user does not exist")
	ErrDomainNotAllowed     = fmt.Errorf("domain not allowed")
	ErrInvalidTenant        = fmt.Errorf("tenant does not exist")
	ErrAuthSessionMissing   = fmt.Errorf("no auth session")
	ErrAuthSessionExpired   = fmt.Errorf("auth session expired")
	ErrAuthSessionInvalid   = fmt.Errorf("auth session invalid")
)

type ConfigLoader interface {
	LoadConfig(ctx context.Context) (*Config, []error)
}

type (
	Database interface {
		Client(context.Context) *ent.Client
		WithTx(context.Context, func(context.Context, *ent.Client) error, ...ent.TxOption) error
		AcquireTxLocks(context.Context, string, ...string) error
		IsTransientError(error) bool
		Shutdown() error
	}

	MigrationService interface {
		GetCurrentStatus(context.Context) (*MigrationStatus, error)
		CreateSchemaMigration(ctx context.Context, name string) error
		Run(ctx context.Context, direction string) error
		UpdateChecksum() error
	}

	MigrationStatus struct {
		CurrentVersion uint
		LatestVersion  uint
		Dirty          bool
	}
)

type (
	NewLoggerOptions struct {
		Parent      *slog.Logger
		PackageName string
		Level       slog.Leveler
		Attrs       []slog.Attr
		Groups      []string
	}

	TelemetryService interface {
		NewLogger(opts NewLoggerOptions) *slog.Logger
		Logger() *slog.Logger

		TracerProvider() trace.TracerProvider
		Tracer(name string, opts ...trace.TracerOption) trace.Tracer
		DefaultTracer() trace.Tracer

		MeterProvider() metric.MeterProvider
		Meter(name string, opts ...metric.MeterOption) metric.Meter
		DefaultMeter() metric.Meter
	}
)

type (
	MessageService interface {
		AddCommandHandlers(handlers ...cqrs.CommandHandler) error
		SendCommand(ctx context.Context, cmd any) error

		AddEventHandlers(handlers ...cqrs.EventHandler) error
		PublishEvent(ctx context.Context, event any) error
	}
)

func NewCommandHandler[T any](name string, handleFn func(context.Context, *T) error) cqrs.CommandHandler {
	return cqrs.NewCommandHandler[T](name, handleFn)
}

func NewEventHandler[T any](name string, handleFn func(context.Context, *T) error) cqrs.EventHandler {
	return cqrs.NewEventHandler[T](name, handleFn)
}

type (
	JobService interface {
		Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
		InsertMany(context.Context, []river.InsertManyParams) ([]*rivertype.JobInsertResult, error)
	}
)

type (
	ProviderEvent struct {
		Provider           string
		ProviderSource     string
		ProviderEventRef   string
		ProviderSubjectRef string
		ReceivedAt         time.Time
		Payload            []byte
		ContentType        string
		RequestMetadata    map[string]string
	}

	ProviderEventQueryResult struct {
		Event             ProviderEvent
		SourceCursorAfter *string
	}

	ProviderEventQuerySourceCursors map[string]string

	ProviderEventQuerier interface {
		QueryProviderEvents(ctx context.Context, sourceCursors ProviderEventQuerySourceCursors) iter.Seq2[*ProviderEventQueryResult, error]
	}

	ProviderEventProcessor interface {
		ProcessProviderEvent(context.Context, ProviderEvent) (ent.NormalizedEvents, error)
	}

	ProjectedEntityRef struct {
		Kind string
		Id   uuid.UUID
	}

	NormalizedEventProjector interface {
		HandleEventProjection(context.Context, *ent.NormalizedEvent) ([]ProjectedEntityRef, error)
	}

	ProviderEventSyncResult struct {
		SyncErrors         []error
		SourceCursorsAfter ProviderEventQuerySourceCursors
		EventsPulled       int
		EventsIngested     int
		NumDuplicates      int
	}

	ProviderEventPipelineService interface {
		Ingest(context.Context, ProviderEvent) error
		SyncEvents(context.Context, ProviderEventQuerier, ProviderEventQuerySourceCursors) ProviderEventSyncResult
	}
)

type (
	ProjectedKnowledgeEvidence struct {
		Kind         ke.EvidenceKind
		Assertion    string
		EffectiveAt  time.Time
		Properties   map[string]any
		SubjectAlias ProjectedKnowledgeEvidenceSubjectAlias
	}

	ProjectedKnowledgeEvidenceSubjectAlias struct {
		Description            string
		AliasRef               ent.KnowledgeSubjectAliasRef
		SubjectEntityRef       *ent.KnowledgeEntityRef
		SubjectRelationshipRef *ent.KnowledgeRelationshipRef
	}

	KnowledgeIngestionService interface {
		IngestProjectedEvidence(context.Context, *ent.NormalizedEvent, ...ProjectedKnowledgeEvidence) error
		IngestDomainEntityEvidence(context.Context, *ent.NormalizedEvent, ProjectedKnowledgeEvidence) (uuid.UUID, error)
	}
)

type (
	IntegrationPackage interface {
		Name() string
		DisplayName() string
		Description() string
		Provider() string
		IsAvailable() (bool, error)
		MaxInstalls() *int
		OAuthInstallRequired() bool
		ValidateInstallationConfig([]byte) (IntegrationInstallationConfig, error)
		ValidateUserSettings(map[string]any) error
		GetInstalledIntegration(*ent.Integration) (InstalledIntegration, error)
	}

	InstalledIntegration interface {
		Integration() *ent.Integration
		Config() IntegrationInstallationConfig
	}

	IntegrationInstallationConfig interface {
		Encode() ([]byte, error)
		ExternalRef() string
	}

	ListIntegrationsParams struct {
		ent.ListParams
		Predicates []predicate.Integration
	}

	CompleteIntegrationOAuth2FlowResult struct {
		InstallationTargetSelectionRequired bool
		Installed                           []InstalledIntegration
		InstallationTargetOptions           []IntegrationInstallationTarget
	}

	IntegrationInstallationTarget struct {
		IntegrationName string
		DisplayName     string
		Config          IntegrationInstallationConfig
	}

	CompleteIntegrationOAuth2Params struct {
		Code           string
		State          *string
		ClientVerifier *string
	}

	IntegrationService interface {
		GetAvailable() []IntegrationPackage

		InstallNew(context.Context, string, []byte) (InstalledIntegration, error)
		ListUserInstallationTargets(ctx context.Context) ([]IntegrationInstallationTarget, error)
		InstallFromTarget(context.Context, IntegrationInstallationTarget) (InstalledIntegration, error)

		LookupInstallation(context.Context, predicate.Integration) (*ent.Integration, error)
		ListInstalled(ctx context.Context, params ListIntegrationsParams) ([]InstalledIntegration, error)
		UpdateInstallation(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (InstalledIntegration, error)
		DeleteInstalled(ctx context.Context, id uuid.UUID) error

		AsInstalledIntegration(i *ent.Integration) (InstalledIntegration, error)

		StartOAuth2Flow(ctx context.Context, integrationName string) (string, error)
		CompleteOAuth2Flow(ctx context.Context, integrationName string, params CompleteIntegrationOAuth2Params) (*CompleteIntegrationOAuth2FlowResult, error)

		RequestIntegrationEventSync(ctx context.Context, id uuid.UUID, sources []string) error
		ListIntegrationEventSyncRuns(ctx context.Context, id uuid.UUID) (*ent.ListResult[ent.IntegrationEventSyncRun], error)
	}
)

type (
	ExpandAnnotationsParams struct {
		WithCreator       bool
		WithRoster        bool
		WithAlertFeedback bool
		WithEvent         bool
	}

	ListAnnotationsParams struct {
		ent.ListParams
		From     time.Time
		To       time.Time
		UserIds  []uuid.UUID
		EventIds []uuid.UUID
		Expand   ExpandAnnotationsParams
	}

	ListEventsParams struct {
		ent.ListParams
		Predicates           []predicate.NormalizedEvent
		WithProjections      bool
		ProjectionPredicates []predicate.NormalizedEventProjection
		WithAnnotations      bool
		AnnotationPredicates []predicate.EventAnnotation
	}

	GetEventParams struct {
		WithProjections      bool
		ProjectionPredicates []predicate.NormalizedEventProjection
	}

	EventsService interface {
		GetEvent(ctx context.Context, id uuid.UUID, params GetEventParams) (*ent.NormalizedEvent, error)
		ListEvents(ctx context.Context, params ListEventsParams) (*ent.ListResult[ent.NormalizedEvent], error)

		ListAnnotations(ctx context.Context, params ListAnnotationsParams) (*ent.ListResult[ent.EventAnnotation], error)

		QueryAnnotation(context.Context, predicate.EventAnnotation) (*ent.EventAnnotation, error)

		GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error)
		SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error)
		DeleteAnnotation(ctx context.Context, id uuid.UUID) error
	}
)

type (
	OrganizationService interface {
		Get(context.Context, predicate.Organization) (*ent.Organization, error)
		Set(context.Context, uuid.UUID, func(*ent.OrganizationMutation)) (*ent.Organization, error)
		SetPreferences(ctx context.Context, orgId uuid.UUID, setFn func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error)
	}
)

type (
	ListUsersParams = struct {
		ent.ListParams
		TeamID uuid.UUID
	}

	UserService interface {
		Get(context.Context, predicate.User) (*ent.User, error)
		Set(context.Context, uuid.UUID, func(*ent.UserMutation)) (*ent.User, error)
		List(context.Context, ListUsersParams) ([]*ent.User, error)
	}
)

type (
	UserAuthProviderSession struct {
		User      ent.User
		Org       ent.Organization
		ExpiresAt time.Time
	}

	AuthSessionService interface {
		CreateFromUserAuthResponse(context.Context, *UserAuthProviderSession) (*ent.UserAuthSession, error)
		CreateForToken(context.Context, string) (*ent.UserAuthSession, error)
		LookupSession(context.Context, uuid.UUID) (*ent.UserAuthSession, error)
		DeleteSession(context.Context, uuid.UUID) error
	}
)

type (
	ListKnowledgeGraphEntitiesParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeEntity
	}

	ListKnowledgeGraphRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeRelationship
	}

	CreateKnowledgeGraphSnapshotParams struct {
		Name              string
		AsOf              time.Time
		Scope             string
		ScopeProperties   map[string]any
		EntityIDs         []uuid.UUID
		RootEntityIDs     []uuid.UUID
		Depth             int
		EntityKinds       []string
		RelationshipKinds []string
	}

	GetKnowledgeGraphViewParams struct {
		Depth             int
		RelationshipKinds []string
	}

	KnowledgeGraphView struct {
		Entities      ent.KnowledgeEntities
		Relationships ent.KnowledgeRelationships
	}

	KnowledgeGraphService interface {
		ListEntities(context.Context, ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error)
		GetEntity(context.Context, uuid.UUID) (*ent.KnowledgeEntity, error)
		GetView(context.Context, uuid.UUID, GetKnowledgeGraphViewParams) (*KnowledgeGraphView, error)
		ListRelationships(context.Context, ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)

		CreateSnapshot(context.Context, CreateKnowledgeGraphSnapshotParams) (*ent.KnowledgeGraphSnapshot, error)
		GetSnapshot(context.Context, uuid.UUID) (*ent.KnowledgeGraphSnapshot, error)
	}
)

type (
	ChatService interface {
		SendMessage(ctx context.Context, id string, msg *ContentNode) (string, error)
		SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error)
		SendTextMessage(ctx context.Context, id string, text string) (string, error)
	}
)

type (
	VideoConferenceService interface {
		CreateIncidentVideoConference(context.Context, *ent.Incident) error
	}
)

type (
	ListTeamsParams struct {
		ent.ListParams
		TeamIds []uuid.UUID
		UserIds []uuid.UUID
	}
	TeamService interface {
		GetById(context.Context, uuid.UUID) (*ent.Team, error)
		List(context.Context, ListTeamsParams) (ent.Teams, error)
	}
)

type (
	ContentNode = prosemirror.Node

	DocumentSessionAuth struct {
		ServerUrl    *url.URL
		DocumentName string
		Token        string
	}

	DocumentsService interface {
		GetDocument(context.Context, uuid.UUID) (*ent.Document, error)
		SetDocument(context.Context, uuid.UUID, func(*ent.DocumentMutation)) (*ent.Document, error)
		GetUserDocumentAccess(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*ent.DocumentAccess, error)
		CreateDocumentEditorSessionAuth(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*DocumentSessionAuth, error)
	}
)

type (
	AiSessionStateService interface {
		GetLatestAgentRunSnapshot(ctx context.Context, runId uuid.UUID) (*ent.AiAgentRunSnapshot, error)
		GetAgentRunSnapshot(context.Context, uuid.UUID) (*ent.AiAgentRunSnapshot, error)
		SetAgentRunSnapshot(context.Context, uuid.UUID, func(*ent.AiAgentRunSnapshotMutation)) (*ent.AiAgentRunSnapshot, error)
		UpdateAgentRunSnapshot(context.Context, uuid.UUID, func(*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation) error) (*ent.AiAgentRunSnapshot, error)
		WriteAgentRunOutput(ctx context.Context, runId uuid.UUID, output any) error
	}

	AiAgentInvoker interface {
		Invoke(ctx context.Context, parentId *uuid.UUID, msg *ai.Message, resume *ai.GenerateActionResume) (uuid.UUID, error)
	}

	AiWorkflowInvoker interface {
		Run(context.Context, any) (any, error)
	}

	AiService interface {
		ValidateAgentRunInput(name string, input []byte) error
		GetAgentRunner(run *ent.AiAgentRun) (AiAgentInvoker, error)
		GetWorkflowInvoker(name string) (AiWorkflowInvoker, error)
	}

	CreateAgentRunParams struct {
		OwnerUserID      uuid.UUID
		PermissionScopes []string
		Input            any
		Metadata         map[string]any
	}

	ListAgentRunsParams struct {
		ent.ListParams
		Predicates []predicate.AiAgentRun
	}

	InvokeAgentRunParams struct {
		ParentSnapshotID uuid.UUID
		Message          *ai.Message
		Resume           *ai.GenerateActionResume
	}

	AiAgentService interface {
		AiSessionStateService
		LookupAgentRunsByMetadata(context.Context, map[string]any) (ent.AiAgentRuns, error)
		ListAgentRuns(context.Context, ListAgentRunsParams) (*ent.ListResult[ent.AiAgentRun], error)
		CreateAgentRun(context.Context, string, CreateAgentRunParams) (*ent.AiAgentRun, error)
		InvokeAgentRun(context.Context, uuid.UUID, InvokeAgentRunParams) error
		GetAgentRun(context.Context, uuid.UUID) (*ent.AiAgentRun, error)
		GetAgentRunOutput(context.Context, uuid.UUID) (*ent.AiAgentRunOutput, error)
		//ClaimAgentRunOutput(context.Context, uuid.UUID, func(context.Context, []byte) (map[string]any, error)) error
	}

	EventOnAiAgentRunOutput struct {
		AgentName        string
		AgentRunMetadata map[string]any
		AgentRunId       uuid.UUID
		AgentOutputId    uuid.UUID
	}
)

type (
	ListAlertsParams struct {
		ent.ListParams
	}

	GetAlertMetricsParams struct {
		AlertId  uuid.UUID
		RosterId uuid.UUID
		From     time.Time
		To       time.Time
	}

	AlertService interface {
		ListAlerts(context.Context, ListAlertsParams) ([]*ent.Alert, int, error)
		GetAlert(context.Context, uuid.UUID) (*ent.Alert, error)
		GetAlertInstance(context.Context, uuid.UUID) (*ent.AlertInstance, error)
		GetAlertMetrics(context.Context, GetAlertMetricsParams) (*ent.AlertMetrics, error)
		GetActiveAlertsForComponents(context.Context, []uuid.UUID) ([]*ent.Alert, error)
	}
)

type (
	ListPlaybooksParams struct {
		ent.ListParams
	}

	PlaybookService interface {
		ListPlaybooks(context.Context, ListPlaybooksParams) ([]*ent.Playbook, int, error)
		GetPlaybook(context.Context, uuid.UUID) (*ent.Playbook, error)
		SetPlaybook(context.Context, *ent.Playbook) (*ent.Playbook, error)
	}
)

type (
	IncidentMetadata struct {
		Roles      ent.IncidentRoles
		Types      ent.IncidentTypes
		Fields     ent.IncidentFields
		Severities ent.IncidentSeverities
		Tags       ent.IncidentTags
	}

	ListIncidentsParams struct {
		ent.ListParams
		UserId       uuid.UUID
		OpenedAfter  time.Time
		OpenedBefore time.Time
	}

	IncidentService interface {
		ListIncidents(context.Context, ListIncidentsParams) (*ent.ListResult[ent.Incident], error)
		Query(context.Context, predicate.Incident, func(*ent.IncidentQuery)) (*ent.Incident, error)
		Get(context.Context, predicate.Incident) (*ent.Incident, error)
		Set(context.Context, uuid.UUID, func(*ent.IncidentMutation)) (*ent.Incident, error)
		Archive(context.Context, uuid.UUID) error

		GetIncidentMilestone(context.Context, uuid.UUID) (*ent.IncidentMilestone, error)
		SetIncidentMilestone(context.Context, uuid.UUID, func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error)

		GetIncidentMetadata(context.Context) (*IncidentMetadata, error)

		ListIncidentRoles(context.Context) ([]*ent.IncidentRole, error)
		ListIncidentSeverities(context.Context) ([]*ent.IncidentSeverity, error)
		ListIncidentTypes(context.Context) ([]*ent.IncidentType, error)
		ListIncidentTags(context.Context) ([]*ent.IncidentTag, error)

		GetIncidentSeverity(context.Context, uuid.UUID) (*ent.IncidentSeverity, error)
	}

	EventOnIncidentUpdated struct {
		Created    bool
		IncidentId uuid.UUID
	}

	EventOnIncidentMilestoneUpdated struct {
		Created     bool
		IncidentId  uuid.UUID
		MilestoneId uuid.UUID
	}

	EventOnIncidentImpactsUpdated struct {
		IncidentId uuid.UUID
	}
)

type (
	DebriefService interface {
		CreateDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error)
		GetDebrief(ctx context.Context, id uuid.UUID) (*ent.IncidentDebrief, error)
		GetUserDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error)
		AddDebriefMessage(ctx context.Context, debriefId uuid.UUID, text string) (*ent.IncidentDebriefMessage, error)

		StartDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error)
		CompleteDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error)
	}
)

type (
	ListRetrospectiveCommentsParams struct {
		ent.ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	ListRetrospectiveReviewsParams struct {
		ent.ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	RetrospectiveService interface {
		Get(context.Context, predicate.Retrospective) (*ent.Retrospective, error)
		Set(context.Context, uuid.UUID, func(*ent.RetrospectiveMutation)) (*ent.Retrospective, error)

		ListComments(context.Context, ListRetrospectiveCommentsParams) ([]*ent.RetrospectiveComment, error)
		GetComment(context.Context, uuid.UUID) (*ent.RetrospectiveComment, error)
		SetComment(context.Context, *ent.RetrospectiveComment) (*ent.RetrospectiveComment, error)
	}
)

type (
	ListOncallRostersParams = struct {
		ent.ListParams
		UserID uuid.UUID
	}

	ListOncallSchedulesParams = struct {
		ent.ListParams
		UserID uuid.UUID
	}

	OncallRostersService interface {
		ListRosters(context.Context, ListOncallRostersParams) (*ent.ListResult[ent.OncallRoster], error)
		GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error)
		GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error)
		GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error)

		ListSchedules(ctx context.Context, params ListOncallSchedulesParams) (*ent.ListResult[ent.OncallSchedule], error)
	}

	ListOncallShiftsParams struct {
		ent.ListParams
		UserID uuid.UUID
		Anchor time.Time
		Window time.Duration
	}

	OncallShiftHandoverSection struct {
		Header  string            `json:"header"`
		Kind    string            `json:"kind"`
		Content *prosemirror.Node `json:"jsonContent,omitempty"`
	}

	OncallShiftsService interface {
		ListShifts(ctx context.Context, params ListOncallShiftsParams) (*ent.ListResult[ent.OncallShift], error)
		GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallShift, error)
		GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallShift, *ent.OncallShift, error)

		GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
		GetHandoverForShift(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error)
		UpdateShiftHandover(ctx context.Context, handover *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error)
		SendShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
	}

	OncallMetricsService interface {
		GetShiftMetrics(ctx context.Context, id uuid.UUID) (*ent.OncallShiftMetrics, error)
		GetComparisonShiftMetrics(ctx context.Context, from, to time.Time) (*ent.OncallShiftMetrics, error)
	}
)
