package rez

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/texm/prosemirror-go"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationhazardassessment"
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
	// LifecycleService runs until it is shut down or fails. Run closes ready
	// exactly once after the service can accept work.
	LifecycleService interface {
		Run(context.Context, chan<- struct{}) error
		Shutdown(context.Context) error
	}
)

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
		Parent *slog.Logger
		Name   string
		Level  slog.Leveler
		Attrs  []slog.Attr
		Groups []string
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
		Publish(context.Context, any) error
		Subscribe(context.Context, *MessageEventSubscriptionOpts, ...MessageEventHandler) error
	}
)

type (
	JobService interface {
		Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
		InsertMany(context.Context, []river.InsertManyParams) ([]*rivertype.JobInsertResult, error)
		Cancel(context.Context, int64) error
	}
)

type ProviderResourceRef struct {
	Provider          string `json:"provider"`
	ProviderNamespace string `json:"provider_namespace"`
	ResourceRef       string `json:"resource_ref"`
}

func (ref ProviderResourceRef) Validate() error {
	if strings.TrimSpace(ref.Provider) == "" {
		return fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(ref.ResourceRef) == "" {
		return fmt.Errorf("resource_ref is required")
	}
	if ref.Provider != "rezible" && strings.TrimSpace(ref.ProviderNamespace) == "" {
		return fmt.Errorf("provider_namespace is required for provider %q", ref.Provider)
	}
	return nil
}

type (
	// KnowledgeEntityLinkingAttributes contains stable, provider-independent values that
	// may identify the same knowledge entity across provider aliases.
	KnowledgeEntityLinkingAttributes interface {
		Values() map[string]string
	}

	KnowledgeEntityRef struct {
		Category            kne.Category
		Kind                string
		ProviderResourceRef ProviderResourceRef
		LinkingAttributes   KnowledgeEntityLinkingAttributes
	}

	KnowledgeRelationshipRef struct {
		Predicate           knr.Predicate
		ProviderResourceRef ProviderResourceRef
		Source              KnowledgeEntityRef
		Target              KnowledgeEntityRef
	}

	KnowledgeSubjectRef struct {
		Entity       *KnowledgeEntityRef
		Relationship *KnowledgeRelationshipRef
	}

	KnowledgeEvidenceRef struct {
		Kind         ke.Kind
		Assertion    string
		EffectiveAt  time.Time
		SubjectState schematypes.KnowledgeGraphSubjectState
		Subject      KnowledgeSubjectRef
	}

	ListKnowledgeGraphEntitiesParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeEntity
	}

	ListKnowledgeGraphRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeRelationship
	}

	ListKnowledgeGraphEvidenceParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeEvidence
	}

	QueryKnowledgeEntityNeighborhoodParams struct {
		EntityID                 *uuid.UUID
		SourceEntityID           *uuid.UUID
		TargetEntityID           *uuid.UUID
		NeighborEntityCategories []string
		RelationshipPredicates   []string
		Depth                    int

		Offset int
		Limit  int
	}

	KnowledgeGraphNeighborhoodSlice struct {
		RootEntityID  uuid.UUID
		Entities      ent.KnowledgeEntities
		Relationships ent.KnowledgeRelationships

		RelationshipCount int
	}

	GetKnowledgeGraphViewParams struct {
		EntityID               uuid.UUID
		Depth                  int
		RelationshipPredicates []string
	}

	KnowledgeGraphView struct {
		RootID        uuid.UUID
		Entities      ent.KnowledgeEntities
		Relationships ent.KnowledgeRelationships
		Truncated     bool
	}

	KnowledgeGraphEntityNeighborhoodSummary struct {
		OutgoingRelationships map[string]KnowledgeGraphNeighborhoodGroupSummary
		IncomingRelationships map[string]KnowledgeGraphNeighborhoodGroupSummary
	}
	KnowledgeGraphNeighborhoodGroupSummary struct {
		Count int
	}

	KnowledgeGraphService interface {
		ListEntities(context.Context, ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error)
		GetEntity(context.Context, uuid.UUID) (*ent.KnowledgeEntity, error)

		ListRelationships(context.Context, ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)
		GetRelationship(context.Context, uuid.UUID) (*ent.KnowledgeRelationship, error)

		ListEvidence(context.Context, ListKnowledgeGraphEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error)
		GetEvidence(context.Context, uuid.UUID) (*ent.KnowledgeEvidence, error)

		QueryEntityNeighborhood(context.Context, QueryKnowledgeEntityNeighborhoodParams) (*KnowledgeGraphNeighborhoodSlice, error)
		SummarizeEntityNeighborhood(context.Context, uuid.UUID) (*KnowledgeGraphEntityNeighborhoodSummary, error)

		GetView(context.Context, GetKnowledgeGraphViewParams) (*KnowledgeGraphView, error)

		IngestEvidence(context.Context, *ent.NormalizedEvent, ...KnowledgeEvidenceRef) error
		ResolveInternalSubject(context.Context, KnowledgeSubjectRef) (*ent.KnowledgeSubjectAlias, error)
	}
)

type (
	ListSystemAnalysisEntitiesParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisEntity
	}

	ListSystemAnalysisRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisRelationship
	}

	ListSystemAnalysisEntriesParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisEntry
	}

	IncludeSystemAnalysisSubjectsParams struct {
		AnalysisId      uuid.UUID
		EntityIds       []uuid.UUID
		RelationshipIds []uuid.UUID
	}

	SystemAnalysisService interface {
		GetSystemAnalysis(context.Context, uuid.UUID) (*ent.SystemAnalysis, error)
		SetSystemAnalysis(context.Context, uuid.UUID, func(*ent.SystemAnalysisMutation)) (*ent.SystemAnalysis, error)

		IncludeSystemAnalysisSubjects(context.Context, IncludeSystemAnalysisSubjectsParams) error

		ListSystemAnalysisEntities(context.Context, ListSystemAnalysisEntitiesParams) (*ent.ListResult[ent.SystemAnalysisEntity], error)
		SetSystemAnalysisEntity(context.Context, uuid.UUID, func(*ent.SystemAnalysisEntityMutation)) (*ent.SystemAnalysisEntity, error)
		HasSystemAnalysisEntity(context.Context, uuid.UUID, uuid.UUID) (bool, error)
		DeleteSystemAnalysisEntity(context.Context, uuid.UUID) error

		ListSystemAnalysisRelationships(context.Context, ListSystemAnalysisRelationshipsParams) (*ent.ListResult[ent.SystemAnalysisRelationship], error)
		SetSystemAnalysisRelationship(context.Context, uuid.UUID, func(*ent.SystemAnalysisRelationshipMutation)) (*ent.SystemAnalysisRelationship, error)
		DeleteSystemAnalysisRelationship(context.Context, uuid.UUID) error

		ListSystemAnalysisEntries(context.Context, ListSystemAnalysisEntriesParams) (*ent.ListResult[ent.SystemAnalysisEntry], error)
		LookupSystemAnalysisEntry(context.Context, predicate.SystemAnalysisEntry) (*ent.SystemAnalysisEntry, error)
		SetSystemAnalysisEntry(context.Context, uuid.UUID, func(*ent.SystemAnalysisEntryMutation), ...func(*ent.SystemAnalysisEntrySubjectMutation)) (*ent.SystemAnalysisEntry, error)
		DeleteSystemAnalysisEntry(context.Context, uuid.UUID) error

		SetSystemAnalysisEntrySubject(context.Context, uuid.UUID, func(*ent.SystemAnalysisEntrySubjectMutation)) (*ent.SystemAnalysisEntrySubject, error)
		DeleteSystemAnalysisEntrySubject(context.Context, uuid.UUID) error
	}
)

type (
	ProviderEvent struct {
		Provider            string
		ProviderNamespace   string
		ProviderEventSource string
		ProviderEventRef    string
		Attributes          json.RawMessage
		ReceivedAt          time.Time
	}

	ProviderEventQueryResult struct {
		Event                          ProviderEvent
		ProviderEventSourceCursorAfter *string
	}

	ProviderEventSourceCursors map[string]string

	ProviderEventQuerier interface {
		QueryProviderEvents(context.Context, ProviderEventSourceCursors) iter.Seq2[*ProviderEventQueryResult, error]
	}

	ProviderEventProcessor interface {
		ProcessProviderEvent(context.Context, ProviderEvent) (ent.NormalizedEvents, error)
	}

	ProviderEventProcessorRegistry map[string]ProviderEventProcessor

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
		SourceSyncDurations map[string]time.Duration
		SourceCursorsAfter  ProviderEventSourceCursors
		SyncErrors          []error
		EventsPulled        int
		EventsIngested      int
		NumDuplicates       int
	}

	ProviderEventPipelineService interface {
		Ingest(context.Context, ProviderEvent) error
		SyncEvents(context.Context, ProviderEventQuerier, ProviderEventSourceCursors) ProviderEventSyncResult
	}
)

func (c ProviderEventSourceCursors) GetForSource(source string) (string, bool) {
	sc, ok := c[source]
	return sc, ok || len(c) == 0
}

type (
	IntegrationDefinition interface {
		Name() string
		Provider() string
		DisplayName() string
		Description() string
		Capabilities() []string
		IsAvailable() (bool, error)
		MaxInstalls() *int
		OAuthInstallRequired() bool
		ValidateInstallationConfig([]byte) (IntegrationInstallationConfig, error)
		ValidateUserSettings(map[string]any) error
		GetInstalledIntegration(*ent.Integration) (InstalledIntegration, error)
	}

	IntegrationInstallationConfig interface {
		Encode() ([]byte, error)
		InstallationTargetRef() ProviderResourceRef
	}

	InstalledIntegration interface {
		Integration() *ent.Integration
		Config() IntegrationInstallationConfig
		Capabilities() []string
	}

	OAuth2FlowIntegration interface {
		OAuth2Config() *oauth2.Config
		RetrieveInstallationTargetOptions(context.Context, *oauth2.Token) ([]IntegrationInstallationTarget, error)
	}

	IntegrationRegistry interface {
		Register(IntegrationDefinition) error
		GetAvailable() []IntegrationDefinition
		Get(string) (IntegrationDefinition, error)
		GetAvailableWebhookHandlers() map[string]http.Handler
		GetOAuth2FlowIntegration(string) (OAuth2FlowIntegration, error)
		GetProviderEventQuerier(InstalledIntegration) (ProviderEventQuerier, error)
		GetAvailableAgentTools(context.Context, []InstalledIntegration, GetAvailableAgentToolsParams) (map[IntegrationDefinition][]ai.Tool, error)
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
		Installed                           []InstalledIntegration
		InstallationTargetSelectionRequired bool
		InstallationTargetOptions           []IntegrationInstallationTarget
	}

	IntegrationInstallationTarget struct {
		DisplayName string
		Config      IntegrationInstallationConfig
	}

	GetAvailableAgentToolsParams struct {
		AgentName string
	}

	IntegrationService interface {
		GetAvailable() []IntegrationDefinition

		InstallNew(context.Context, string, []byte) (InstalledIntegration, error)
		ListUserInstallationTargets(ctx context.Context) ([]IntegrationInstallationTarget, error)
		InstallFromTarget(context.Context, IntegrationInstallationTarget) (InstalledIntegration, error)

		LookupInstallation(context.Context, predicate.Integration) (*ent.Integration, error)
		ListAllInstalled(ctx context.Context, predicates ...predicate.Integration) ([]InstalledIntegration, error)
		UpdateInstallation(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (InstalledIntegration, error)
		DeleteInstalled(ctx context.Context, id uuid.UUID) error

		AsInstalledIntegration(i *ent.Integration) (InstalledIntegration, error)
		GetAvailableAgentTools(context.Context, GetAvailableAgentToolsParams) ([]ai.Tool, error)

		StartOAuth2Flow(ctx context.Context, integrationName string) (string, error)
		CompleteOAuth2Flow(ctx context.Context, integrationName string, params CompleteIntegrationOAuth2FlowParams) (*CompleteIntegrationOAuth2FlowResult, error)

		RequestIntegrationEventSync(ctx context.Context, id uuid.UUID, sources []string) error
		ListIntegrationEventSyncRuns(ctx context.Context, id uuid.UUID) ([]*ent.IntegrationEventSyncRun, error)
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
		List(context.Context, ListUsersParams) (*ent.ListResult[ent.User], error)
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
	ValidatingInput interface {
		Validate() error
	}

	AiWorkflowInput = ValidatingInput

	AiWorkflowOutput interface{}

	AiWorkflowRunner = interface {
		Run(context.Context, AiWorkflowInput) (AiWorkflowOutput, error)
	}

	AiAgentTurnInput struct {
		Message *ai.Message     `json:"message,omitempty"`
		Resume  *aix.ToolResume `json:"resume,omitempty"`
	}

	AiAgentTurnChunk struct {
		Artifact            *aix.Artifact          `json:"artifact,omitempty"`
		ModelChunk          *ai.ModelResponseChunk `json:"model_chunk,omitempty"`
		TurnEndFinishReason *aix.AgentFinishReason `json:"finish_reason,omitempty"`
	}

	AiAgentTurnState struct {
		Messages  []*ai.Message
		Artifacts []*aix.Artifact
	}

	InvokeAgentTurnParams struct {
		Session *ent.AgentSession
		Turn    *ent.AgentTurn
		State   AiAgentTurnState
		Input   *AiAgentTurnInput
		OnChunk func(AiAgentTurnChunk)
	}

	AiAgentInvocationResult struct {
		State        AiAgentTurnState
		Response     *ai.Message
		FinishReason aix.AgentFinishReason
		Error        error
	}

	AiAgentConfig struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Model       string `json:"model"`
	}

	AiService interface {
		GetWorkflowRunner(string) (AiWorkflowRunner, error)
		GetAgents() []AiAgentConfig
		ValidateAgentSessionInput(string, []byte) (ValidatingInput, error)
		MakeInitialAgentTurnInput(context.Context, *ent.AgentSession) (*AiAgentTurnInput, error)
		InvokeAgentTurn(context.Context, InvokeAgentTurnParams) (*AiAgentInvocationResult, error)
	}

	ListAgentSessionsParams struct {
		ent.ListParams
		Predicates []predicate.AgentSession
		Metadata   map[string]any
	}

	CreateAgentSessionParams struct {
		AgentName        string
		PermissionScopes []string
		Input            ValidatingInput
		SystemAnalysisID *uuid.UUID
		Metadata         map[string]any
		Bindings         []AgentSessionBindingParams
	}

	AgentSessionBindingParams struct {
		ProviderResourceRef
		IntegrationID *uuid.UUID
		Metadata      map[string]any
	}

	ListAgentSessionBindingsParams struct {
		ent.ListParams
		Predicates []predicate.AgentSessionBinding
	}

	RequestAgentTurnParams struct {
		Input *AiAgentTurnInput
	}

	ListAgentTurnsParams struct {
		ent.ListParams
		Predicates []predicate.AgentTurn
	}

	ListAgentMessagesParams struct {
		ent.ListParams
		Predicates []predicate.AgentMessage
	}

	ListAgentArtifactsParams struct {
		ent.ListParams
		Predicates []predicate.AgentArtifact
	}

	AgentSessionService interface {
		ListAgentSessions(context.Context, ListAgentSessionsParams) (*ent.ListResult[ent.AgentSession], error)
		CreateAgentSession(context.Context, CreateAgentSessionParams) (*ent.AgentSession, error)
		GetAgentSession(context.Context, uuid.UUID) (*ent.AgentSession, error)

		RequestAgentTurn(context.Context, uuid.UUID, *RequestAgentTurnParams) (*ent.AgentTurn, error)

		ListAgentTurns(context.Context, ListAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error)
		GetAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		RetryAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		AbortAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)

		ListAgentMessages(context.Context, ListAgentMessagesParams) (*ent.ListResult[ent.AgentMessage], error)
		ListAgentArtifacts(context.Context, ListAgentArtifactsParams) (*ent.ListResult[ent.AgentArtifact], error)

		ListAgentSessionBindings(context.Context, ListAgentSessionBindingsParams) (ent.AgentSessionBindings, error)
		LookupAgentSessionBinding(context.Context, ...predicate.AgentSessionBinding) (*ent.AgentSessionBinding, error)
		SetAgentSessionBinding(context.Context, uuid.UUID, func(*ent.AgentSessionBindingMutation)) (*ent.AgentSessionBinding, error)
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
		ListAlerts(context.Context, ListAlertsParams) (*ent.ListResult[ent.AlertDefinition], error)
		GetAlert(context.Context, uuid.UUID) (*ent.AlertDefinition, error)
		GetAlertInstance(context.Context, uuid.UUID) (*ent.AlertInstance, error)
		GetAlertMetrics(context.Context, GetAlertMetricsParams) (*ent.AlertMetrics, error)

		RecordAlertDefinitionInstance(context.Context, uuid.UUID, *ent.NormalizedEvent) (*ent.AlertInstance, error)
	}
)

type (
	ListSituationsParams struct {
		ent.ListParams
		Status      situation.Status
		OpenedAfter *time.Time
	}

	CreateSituationParams struct {
		Title         string
		Summary       string
		OpenedAt      time.Time
		EvidenceItems []SituationEvidenceItemParams
	}

	SituationEvidenceItemParams struct {
		AlertEpisodeID *uuid.UUID
		IncidentID     *uuid.UUID
	}

	CloseSituationParams struct {
		SituationID uuid.UUID
		Reason      situation.CloseReason
	}

	SetSituationInvestigationReportParams struct {
		AgentTurnID uuid.UUID
		Report      schematypes.SituationInvestigationReport
		Assessments []SituationInvestigationHazardAssessment
	}

	SituationInvestigationHazardAssessment struct {
		SystemHazardID uuid.UUID                        `json:"systemHazardId"`
		Status         situationhazardassessment.Status `json:"status"`
		Summary        string                           `json:"summary"`
	}

	ListSituationHazardAssessmentsParams struct {
		ent.ListParams
		SituationID    uuid.UUID
		SystemHazardID uuid.UUID
	}

	AddSituationHazardAssessmentParams struct {
		SituationID    uuid.UUID
		SystemHazardID uuid.UUID
		Status         situationhazardassessment.Status
		Summary        string
		AssessedAt     time.Time
		UserID         *uuid.UUID
		AgentTurnID    *uuid.UUID
	}

	SituationService interface {
		ListSituations(context.Context, ListSituationsParams) (*ent.ListResult[ent.Situation], error)
		CreateSituation(context.Context, CreateSituationParams) (*ent.Situation, error)
		GetSituation(context.Context, uuid.UUID) (*ent.Situation, error)
		CloseSituation(context.Context, CloseSituationParams) (*ent.Situation, error)

		AddSituationEvidenceItem(context.Context, uuid.UUID, SituationEvidenceItemParams) error
		NotifySituationEvidenceItemUpdated(context.Context, uuid.UUID, SituationEvidenceItemParams) error
		RemoveSituationEvidenceItem(context.Context, uuid.UUID, SituationEvidenceItemParams) error

		CreateSituationInvestigation(context.Context, uuid.UUID) (*ent.SituationInvestigation, error)
		GetSituationInvestigation(context.Context, uuid.UUID) (*ent.SituationInvestigation, error)
		SetSituationInvestigationReport(context.Context, SetSituationInvestigationReportParams) (*ent.SituationInvestigation, error)

		ListSituationHazardAssessments(context.Context, ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error)
		AddSituationHazardAssessment(context.Context, AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error)
	}
)

type (
	CreateSystemHazardParams struct {
		Title                 string
		Description           string
		PotentialConsequences string
	}

	AddSystemHazardRiskAssessmentParams struct {
		SystemHazardID uuid.UUID
		Likelihood     string
		Consequence    string
		RiskLevel      string
		Rationale      string
		AssessedAt     time.Time
	}

	ListSystemHazardRiskAssessmentsParams struct {
		ent.ListParams
		SystemHazardID uuid.UUID
	}

	SystemHazardService interface {
		CreateSystemHazard(context.Context, CreateSystemHazardParams) (*ent.SystemHazard, error)
		GetSystemHazard(context.Context, uuid.UUID) (*ent.SystemHazard, error)
		RetireSystemHazard(context.Context, uuid.UUID) (*ent.SystemHazard, error)

		AddSystemHazardRiskAssessment(context.Context, AddSystemHazardRiskAssessmentParams) (*ent.SystemHazardRiskAssessment, error)
		ListSystemHazardRiskAssessments(context.Context, ListSystemHazardRiskAssessmentsParams) (*ent.ListResult[ent.SystemHazardRiskAssessment], error)
		GetLatestSystemHazardRiskAssessment(context.Context, uuid.UUID) (*ent.SystemHazardRiskAssessment, error)
	}
)

type (
	ListPlaybooksParams struct {
		ent.ListParams
		AlertID uuid.UUID
	}

	PlaybookService interface {
		ListPlaybooks(context.Context, ListPlaybooksParams) (*ent.ListResult[ent.Playbook], error)
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

		ListComments(context.Context, ListRetrospectiveCommentsParams) (*ent.ListResult[ent.RetrospectiveComment], error)
		GetComment(context.Context, uuid.UUID) (*ent.RetrospectiveComment, error)
		SetComment(context.Context, *ent.RetrospectiveComment) (*ent.RetrospectiveComment, error)
	}
)

type (
	ListOncallRostersParams = struct {
		ent.ListParams
		TeamID uuid.UUID
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
