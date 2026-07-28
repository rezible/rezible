package rez

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/google/uuid"
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
	ErrConflict             = fmt.Errorf("conflict")
	ErrInvalidInput         = fmt.Errorf("invalid input")
)

type (
	Database interface {
		Client(context.Context) *ent.Client
		WithTx(context.Context, func(context.Context, *ent.Client) error, ...ent.TxOption) error
		AcquireTxLocks(context.Context, string, ...string) error
		IsTransientError(error) bool
		Shutdown() error
	}

	DatabaseNotificationService interface {
		Listen(ctx context.Context, channel string, onConnect func(context.Context) error, onNotify func(context.Context, []byte) error) error
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
	MessageEventHandler interface {
		HandlerName() string
		NewEvent() any
		Handle(context.Context, any) error
	}

	MessageEventWithScopes interface {
		MessageScopes() []string
	}

	MessageEventSubscriptionOpts struct {
		Scopes []string
	}

	MessageService interface {
		AddHandlers(...MessageEventHandler) error
		Publish(context.Context, any) error
		Subscribe(context.Context, MessageEventHandler, *MessageEventSubscriptionOpts) error
	}
)

type (
	JobService interface {
		RegisterPeriodicJob(*river.PeriodicJob)
		Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
		InsertMany(context.Context, []river.InsertManyParams) ([]*rivertype.JobInsertResult, error)
		Cancel(context.Context, int64) error
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

	GetKnowledgeGraphViewParams struct {
		EntityID          uuid.UUID
		RelationshipID    uuid.UUID
		Depth             int
		RelationshipKinds []string
	}

	KnowledgeGraphView struct {
		RootID        uuid.UUID
		Entities      ent.KnowledgeEntities
		Relationships ent.KnowledgeRelationships
		Truncated     bool
	}

	KnowledgeGraphService interface {
		ListEntities(context.Context, ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error)
		GetEntity(context.Context, uuid.UUID) (*ent.KnowledgeEntity, error)

		ListRelationships(context.Context, ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)
		GetRelationship(context.Context, uuid.UUID) (*ent.KnowledgeRelationship, error)

		GetView(context.Context, GetKnowledgeGraphViewParams) (*KnowledgeGraphView, error)
		GetEntityAt(context.Context, uuid.UUID, time.Time) (*ent.KnowledgeEntity, error)
		GetRelationshipAt(context.Context, uuid.UUID, time.Time) (*ent.KnowledgeRelationship, error)

		IngestEntityEvidence(context.Context, *ent.NormalizedEvent, ent.KnowledgeEvidenceRef) (*ent.KnowledgeEntity, error)
		IngestEvidenceBulk(context.Context, *ent.NormalizedEvent, ...ent.KnowledgeEvidenceRef) (ent.KnowledgeSubjectAliasSlice, error)
	}
)

type (
	ProviderEvent struct {
		Provider           string
		ProviderSource     string
		ProviderSubjectRef string
		ProviderEventRef   string
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
		QueryProviderEvents(context.Context, ProviderEventQuerySourceCursors) iter.Seq2[*ProviderEventQueryResult, error]
	}

	ProviderEventProcessor interface {
		ProcessProviderEvent(context.Context, ProviderEvent) (ent.NormalizedEvents, error)
	}

	NormalizedEventProjector interface {
		ProjectEvent(context.Context, *ent.NormalizedEvent) ([]ProjectedEntityRef, error)
	}

	ProjectedEntityRef struct {
		Kind string
		Id   uuid.UUID
	}

	EventProjectorFunc func(context.Context, *ent.NormalizedEvent) ([]ProjectedEntityRef, error)

	EventProjectionService interface {
		GetEventProjectorFunc(*ent.NormalizedEvent) (EventProjectorFunc, bool)
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

	CompleteIntegrationOAuth2FlowParams struct {
		Code           string
		State          *string
		ClientVerifier *string
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
		CompleteOAuth2Flow(ctx context.Context, integrationName string, params CompleteIntegrationOAuth2FlowParams) (*CompleteIntegrationOAuth2FlowResult, error)

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
		WithProjection       bool
		WithAnnotations      bool
		AnnotationPredicates []predicate.EventAnnotation
	}

	GetEventParams struct {
		WithProjection bool
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
		SetPreferences(context.Context, uuid.UUID, func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error)
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
	AgentTurnInput struct {
		Message *ai.Message              `json:"message,omitempty"`
		Resume  *ai.GenerateActionResume `json:"resume,omitempty"`
	}

	AgentTurnChunk struct {
		Artifact            *aix.Artifact          `json:"artifact,omitempty"`
		ModelChunk          *ai.ModelResponseChunk `json:"model_chunk,omitempty"`
		TurnEndFinishReason *aix.AgentFinishReason `json:"finish_reason,omitempty"`
	}

	InvokeAgentTurnParams struct {
		Session *ent.AgentSession
		Parent  *ent.AgentTurn
		Turn    *ent.AgentTurn
		Input   *AgentTurnInput
		OnChunk func(AgentTurnChunk)
	}

	AgentInvocationResult struct {
		State        []byte
		Response     *ai.Message
		FinishReason aix.AgentFinishReason
		Error        *core.GenkitError
	}

	AiService interface {
		MakeInitialAgentTurnInput(context.Context, string, any) (*AgentTurnInput, error)
		InvokeAgentTurn(context.Context, InvokeAgentTurnParams) (*AgentInvocationResult, error)
	}

	CreateAgentSessionParams struct {
		AgentName        string
		OwnerUserID      uuid.UUID
		PermissionScopes []string
		Input            any
		Metadata         map[string]any
	}

	ListAgentSessionsParams struct {
		ent.ListParams
		Predicates []predicate.AgentSession
		Metadata   map[string]any
	}

	RequestAgentTurnParams struct {
		Input        *AgentTurnInput
		ParentTurnID *uuid.UUID
	}

	ListAgentTurnsParams struct {
		ent.ListParams
		Predicates []predicate.AgentTurn
	}

	AgentSessionService interface {
		ListAgentSessions(context.Context, ListAgentSessionsParams) (*ent.ListResult[ent.AgentSession], error)
		CreateAgentSession(context.Context, CreateAgentSessionParams) (*ent.AgentSession, error)
		GetAgentSession(context.Context, uuid.UUID) (*ent.AgentSession, error)

		RequestAgentTurn(context.Context, uuid.UUID, *RequestAgentTurnParams) (*ent.AgentTurn, error)
		//GetLastSuccessfulAgentTurnForSession(context.Context, uuid.UUID) (*ent.AgentTurn, error)

		ListAgentTurns(context.Context, ListAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error)
		GetAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		RetryAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		AbortAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
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
