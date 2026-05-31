package rez

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
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
	ErrTenantContextMissing     = fmt.Errorf("tenant access context not set")
	ErrInvalidUser              = fmt.Errorf("user does not exist")
	ErrDomainNotAllowed         = fmt.Errorf("domain not allowed")
	ErrInvalidTenant            = fmt.Errorf("tenant does not exist")
	ErrAuthSessionMissing       = fmt.Errorf("no auth session")
	ErrAuthSessionExpired       = fmt.Errorf("auth session expired")
	ErrAuthSessionInvalid       = fmt.Errorf("auth session invalid")
	ErrNoConfiguredIntegrations = fmt.Errorf("no configured integrations")
)

type ConfigLoader interface {
	LoadConfig(ctx context.Context) (*Config, error)
}

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
		InsertTx(context.Context, *ent.Tx, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
		InsertMany(context.Context, []river.InsertManyParams) ([]*rivertype.JobInsertResult, error)
		InsertManyTx(ctx context.Context, tx *ent.Tx, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error)
	}
)

type (
	IntegrationPackage interface {
		Name() string
		IsAvailable() (bool, error)
		SupportedDataKinds() []string
		ValidateUserConfig(map[string]any) error
		ValidateUserPreferences(map[string]any) error
		OAuthConfigRequired() bool
		GetConfiguredIntegration(*ent.Integration) ConfiguredIntegration
	}

	ConfiguredIntegration interface {
		ID() uuid.UUID
		Provider() string
		DisplayName() string
		ExternalRef() string
		GetSanitizedConfig() map[string]any
		GetUserPreferences() map[string]any
		GetAvailableDataKinds() map[string]bool
	}

	ListIntegrationsParams struct {
		IDs          []uuid.UUID
		Providers    []string
		ExternalRefs []string
		ConfigValues map[string]any
	}

	ConfigureIntegrationParams struct {
		Provider    string
		DisplayName string
		ExternalRef string
		Config      map[string]any
		Preferences map[string]any
	}

	CompleteIntegrationOAuth2Result struct {
		Status         string
		Configured     []ConfiguredIntegration
		SelectionToken string
		Options        []ExternalIntegrationOption
	}

	ExternalIntegrationOption struct {
		ExternalRef string         `json:"externalRef"`
		DisplayName string         `json:"displayName"`
		Config      map[string]any `json:"config"`
	}

	SelectIntegrationOAuth2Params struct {
		SelectionToken string
		ExternalRefs   []string
	}

	CompleteIntegrationOAuth2Params struct {
		Code           string
		State          *string
		ClientVerifier *string
	}

	ListProviderDataSyncStatusParams struct {
		Provider string
	}

	IntegrationsService interface {
		GetAvailable() []IntegrationPackage

		Configure(ctx context.Context, params ConfigureIntegrationParams) (ConfiguredIntegration, error)
		ListConfigured(ctx context.Context, params ListIntegrationsParams) ([]ConfiguredIntegration, error)
		GetConfigured(ctx context.Context, id uuid.UUID) (ConfiguredIntegration, error)
		UpdateConfiguredPreferences(ctx context.Context, id uuid.UUID, prefs map[string]any) (ConfiguredIntegration, error)
		DeleteConfigured(ctx context.Context, id uuid.UUID) error

		GetProviderEventProcessor(provider string) (ProviderEventProcessor, error)
		GetProviderEventQueriers(ctx context.Context, provider string) ([]ProviderEventQuerier, error)

		StartOAuth2Flow(ctx context.Context, provider string, callbackPath string) (string, error)
		SelectOAuth2Flow(ctx context.Context, provider string, params SelectIntegrationOAuth2Params) (*CompleteIntegrationOAuth2Result, error)
		CompleteOAuth2Flow(ctx context.Context, provider string, params CompleteIntegrationOAuth2Params) (*CompleteIntegrationOAuth2Result, error)

		GetChatService(ctx context.Context) (ChatService, error)

		RequestDataSync(ctx context.Context, provider string, sources []string) error
		GetDataSyncStatus(ctx context.Context, provider string) (*ent.ListResult[ent.ProviderEventSyncRun], error)
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

	ProviderEventProcessor interface {
		Process(context.Context, ProviderEvent) (ent.NormalizedEvents, error)
	}

	ProviderEventQueryRequest struct {
		SourceCursors map[string]string
	}

	ProviderEventQuerier interface {
		Provider() string
		PullEvents(context.Context, ProviderEventQueryRequest) iter.Seq2[*ProviderEventQueryResult, error]
	}

	ProviderEventQueryResult struct {
		Event             ProviderEvent
		SourceCursorAfter *string
	}

	ProviderEventSyncOptions struct {
		SyncReason   string
		QueryRequest ProviderEventQueryRequest
	}

	ProviderEventService interface {
		Ingest(context.Context, ProviderEvent) error
		SyncEvents(context.Context, ProviderEventQuerier, ProviderEventSyncOptions) error
	}

	ProviderEventIngestResult struct {
		Duplicate bool
	}
)

func (req ProviderEventQueryRequest) GetSourceCursor(src string) (string, bool) {
	cursor, exists := req.SourceCursors[src]
	return cursor, exists || len(req.SourceCursors) == 0
}

type (
	OrganizationService interface {
		SyncFromAuthProvider(context.Context, ent.Organization) (*ent.Organization, error)
		Get(context.Context, predicate.Organization) (*ent.Organization, error)
		CompleteSetup(context.Context, *ent.Organization) error
	}
)

type (
	ListUsersParams = struct {
		ent.ListParams
		TeamID uuid.UUID
	}

	UserService interface {
		SyncFromAuthProvider(context.Context, ent.Organization, ent.User) (*ent.User, error)

		Get(context.Context, predicate.User) (*ent.User, error)
		Set(context.Context, uuid.UUID, func(*ent.UserMutation)) (*ent.User, error)
		List(context.Context, ListUsersParams) ([]*ent.User, error)
	}
)

type (
	KnowledgeService interface {
		GetEntity(context.Context, predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error)
	}
)

type (
	ListSystemTopologyEntitiesParams struct {
		ent.ListParams
		Kinds []string
	}

	ListSystemTopologyRelationshipsParams struct {
		ent.ListParams
		Kinds          []string
		EntityID       uuid.UUID
		SourceEntityID uuid.UUID
		TargetEntityID uuid.UUID
	}

	SystemTopologyNeighborhoodParams struct {
		Depth             int
		RelationshipKinds []string
	}

	CreateSystemTopologySnapshotParams struct {
		Name              string
		AsOf              time.Time
		Scope             string
		ScopeProperties   map[string]any
		EntityIDs         []uuid.UUID
		RootEntityIDs     []uuid.UUID
		Depth             int
		EntityKinds       []string
		RelationshipKinds []string
		IncludeIncidents  bool
		IncludeChanges    bool
		IncludeAlerts     bool
	}

	SystemTopologyGraph struct {
		Entities      []*ent.KnowledgeEntity
		Relationships []*ent.KnowledgeRelationship
	}

	SystemTopologyService interface {
		ListEntities(context.Context, ListSystemTopologyEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error)
		GetEntity(context.Context, uuid.UUID) (*ent.KnowledgeEntity, error)
		GetNeighborhood(context.Context, uuid.UUID, SystemTopologyNeighborhoodParams) (*SystemTopologyGraph, error)
		ListRelationships(context.Context, ListSystemTopologyRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)

		CreateSnapshot(context.Context, CreateSystemTopologySnapshotParams) (*ent.SystemTopologySnapshot, error)
		GetSnapshot(context.Context, uuid.UUID) (*ent.SystemTopologySnapshot, error)
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

	DocumentsService interface {
		GetDocument(context.Context, uuid.UUID) (*ent.Document, error)
		SetDocument(context.Context, uuid.UUID, func(*ent.DocumentMutation)) (*ent.Document, error)
		GetDocumentAccess(context.Context, uuid.UUID) (*ent.DocumentAccess, error)
	}
)

//type (
//	AiAgentService interface {
//	}
//)

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
		Set(context.Context, uuid.UUID, func(*ent.IncidentMutation) []ent.Mutation) (*ent.Incident, error)
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
	ListEventsParams struct {
		ent.ListParams
		Predicates      []predicate.NormalizedEvent
		WithAnnotations bool
	}

	EventsService interface {
		GetEvent(ctx context.Context, id uuid.UUID) (*ent.NormalizedEvent, error)
		ListEvents(ctx context.Context, params ListEventsParams) (*ent.ListResult[ent.NormalizedEvent], error)
	}

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

	EventAnnotationsService interface {
		ListAnnotations(ctx context.Context, params ListAnnotationsParams) (*ent.ListResult[ent.EventAnnotation], error)

		Lookup(context.Context, predicate.EventAnnotation) (*ent.EventAnnotation, error)

		GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error)
		SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error)
		DeleteAnnotation(ctx context.Context, id uuid.UUID) error
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

		GetCurrentOncallForComponent(context.Context, uuid.UUID) ([]*ent.User, error)
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
