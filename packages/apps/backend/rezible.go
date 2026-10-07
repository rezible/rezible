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

	"entgo.io/ent/dialect/sql/schema"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/task"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/texm/prosemirror-go"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"

	dt "github.com/rezible/rezible/ent/discussionthread"
	"github.com/rezible/rezible/ent/incident"
	ifvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	ihv "github.com/rezible/rezible/ent/investigationhypothesisversion"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/situation"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/ent/situationsignalattention"
	saent "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
)

var (
	ErrTenantContextMissing = fmt.Errorf("tenant access context not set")
	ErrInvalidUser          = fmt.Errorf("user does not exist")
	ErrDomainNotAllowed     = fmt.Errorf("domain not allowed")
	ErrInvalidTenant        = fmt.Errorf("tenant does not exist")
	ErrForbidden            = fmt.Errorf("forbidden")
	ErrAuthSessionMissing   = fmt.Errorf("no auth session")
	ErrAuthSessionExpired   = fmt.Errorf("auth session expired")
	ErrAuthSessionInvalid   = fmt.Errorf("auth session invalid")
	ErrConflict             = fmt.Errorf("conflict")
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrUnprocessableInput   = fmt.Errorf("unprocessable input")
	ErrNotFound             = fmt.Errorf("not found")
	ErrNotImplemented       = fmt.Errorf("not implemented")
	ErrRateLimited          = fmt.Errorf("rate limited")
)

type (
	// LifecycleService runs until it is shut down or fails. Run closes ready
	// exactly once after the service can accept work.
	LifecycleService interface {
		Run(context.Context, chan<- struct{}) error
		Shutdown(context.Context) error
	}

	LifecycleServiceProvider interface {
		LifecycleService() LifecycleService
	}
)

type (
	Database interface {
		Client(context.Context) *ent.Client
		WithTx(context.Context, func(context.Context, *ent.Client) error, ...ent.TxOption) error
		AcquireTxLocks(context.Context, string, ...string) error
		IsTransientError(error) bool
		IsConstraintError(error) (string, bool)
		Shutdown() error
	}

	MigrationDirection string

	MigrationService interface {
		GetCurrentStatus(context.Context) (*MigrationStatus, error)
		CreateSchemaMigration(ctx context.Context, name string, infraTables ...*schema.Table) error
		Run(context.Context, MigrationDirection) error
		UpdateChecksum() error
	}

	MigrationStatus struct {
		CurrentVersion uint
		LatestVersion  uint
		Dirty          bool
	}
)

type (
	// Clock is the source of processing time for decisions that depend on it.
	Clock interface {
		Now() time.Time
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
	MessageEvent interface {
		MessageName() string
	}

	MessageEventHandler interface {
		HandlerName() string
		NewEvent() any
		Handle(context.Context, any) error
	}

	MessageEventScopes []string

	MessageEventSubscriptionOpts struct {
		Scopes MessageEventScopes
	}

	MessageQueue interface {
		Publish(context.Context, any) error
		PublishLive(context.Context, any) error
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

	KnowledgeSubjectState = schematypes.KnowledgeGraphSubjectState

	KnowledgeEvidenceRef struct {
		Kind         ke.Kind
		Assertion    string
		EffectiveAt  time.Time
		Subject      KnowledgeSubjectRef
		SubjectState KnowledgeSubjectState
	}

	KnowledgeGraphIngestionService interface {
		IngestEvidence(context.Context, *ent.NormalizedEvent, ...KnowledgeEvidenceRef) error
		ResolveInternalSubject(context.Context, KnowledgeSubjectRef) (*ent.KnowledgeSubjectAlias, error)
	}
)

type (
	ListKnowledgeEntitiesParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeEntity
	}

	ListKnowledgeRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeRelationship
	}

	ListKnowledgeSubjectAliasesParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeSubjectAlias
	}

	ListKnowledgeEvidenceParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeEvidence
	}

	KnowledgeGraphQueryService interface {
		ListEntities(context.Context, ListKnowledgeEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error)
		GetEntity(context.Context, uuid.UUID) (*ent.KnowledgeEntity, error)

		ListRelationships(context.Context, ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)
		GetRelationship(context.Context, uuid.UUID) (*ent.KnowledgeRelationship, error)

		ListSubjectAliases(context.Context, ListKnowledgeSubjectAliasesParams) (*ent.ListResult[ent.KnowledgeSubjectAlias], error)
		GetSubjectAlias(context.Context, uuid.UUID) (*ent.KnowledgeSubjectAlias, error)

		ListEvidence(context.Context, ListKnowledgeEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error)
		GetEvidence(context.Context, uuid.UUID) (*ent.KnowledgeEvidence, error)

		SelectGraphEntities(context.Context, SelectKnowledgeGraphEntitiesParams) (*KnowledgeGraphEntitiesPage, error)
		ExpandGraphRelationships(context.Context, ExpandKnowledgeGraphRelationshipsParams) (*KnowledgeGraphRelationshipsPage, error)

		// ResolveStructure returns, for each entity, the entities in TargetCategories that represent it: itself
		// if it is in one, otherwise its nearest ancestors in one, climbing the structure hierarchy at most
		// MaxDepth steps. Entities represented by nothing are omitted.
		ResolveStructure(context.Context, ResolveStructureParams) (map[uuid.UUID][]uuid.UUID, error)
	}

	KnowledgeEntityFilter struct {
		Categories []kne.Category
		Kinds      []string
	}

	EntitySelectionCursor string

	SelectKnowledgeGraphEntitiesParams struct {
		Filter KnowledgeEntityFilter
		Cursor *EntitySelectionCursor
		Limit  *int
	}

	EntitySelectionRef string

	KnowledgeGraphEntitiesPage struct {
		Entities           []*ent.KnowledgeEntity
		EntitySelectionRef EntitySelectionRef
		NextCursor         *EntitySelectionCursor
	}

	RelationshipExpansionCursor string

	ExpandKnowledgeGraphRelationshipsParams struct {
		EntitySelectionRef EntitySelectionRef
		Predicates         []knr.Predicate
		Cursor             *RelationshipExpansionCursor
		Limit              *int
	}

	KnowledgeGraphRelationshipsPage struct {
		Relationships []*ent.KnowledgeRelationship
		NextCursor    *RelationshipExpansionCursor
	}

	ResolveStructureParams struct {
		EntityIDs []uuid.UUID
		// TargetCategories are the categories that represent an entity, for example
		// knowledgegraph.StructureLevelRuntime.Categories().
		TargetCategories []kne.Category
		MaxDepth         int
	}
)

type (
	ListSystemAnalysisEntitiesParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisEntity
		// OrderBy defaults to creation time, then ID ascending.
		OrderBy []saent.OrderOption
	}

	ListSystemAnalysisRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisRelationship
		// OrderBy defaults to creation time, then ID ascending.
		OrderBy       []sarel.OrderOption
		WithEndpoints bool
	}

	ListSystemAnalysisEntriesParams struct {
		ent.ListParams
		Predicates []predicate.SystemAnalysisEntry
		// OrderBy defaults to occurrence time, sequence, then ID ascending.
		OrderBy     []sae.OrderOption
		SummaryOnly bool
		Kinds       []sae.Kind
	}

	SetSystemAnalysisEntryParams struct {
		AnalysisID  uuid.UUID
		Reference   *string
		Kind        sae.Kind
		Title       string
		Body        string
		OccurredAt  *time.Time
		SetSubjects []SetSystemAnalysisEntrySubjectParams
	}

	SetSystemAnalysisEntrySubjectParams struct {
		Role                    string
		KnowledgeEntityID       *uuid.UUID
		KnowledgeRelationshipID *uuid.UUID
		KnowledgeEvidenceID     *uuid.UUID
	}

	ListSystemAnalysisEntrySubjectsParams struct {
		AnalysisID uuid.UUID
		EntryID    uuid.UUID
		ent.ListParams
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
		SetSystemAnalysisEntry(context.Context, uuid.UUID, SetSystemAnalysisEntryParams) (*ent.SystemAnalysisEntry, error)
		DeleteSystemAnalysisEntry(context.Context, uuid.UUID) error

		ListSystemAnalysisEntrySubjects(context.Context, ListSystemAnalysisEntrySubjectsParams) (*ent.ListResult[ent.SystemAnalysisEntrySubject], error)
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

type ProviderEventSourceCursors map[string]string

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
		InstallationLinks() []IntegrationInstallationLink
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

	// IntegrationInstallationLink is an external page that helps users set up or manage an integration installation.
	IntegrationInstallationLink struct {
		Kind  string
		Label string
		URL   string
	}

	OAuth2FlowIntegration interface {
		OAuth2Config() *oauth2.Config
		RetrieveInstallationTargetOptions(context.Context, *oauth2.Token) ([]IntegrationInstallationTarget, error)
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

	GetAvailableAiAgentToolsParams struct {
		AgentName string
	}

	// IntegrationInstallationLookup reads saved installations without resolving integration definitions,
	// so integrations can depend on it without depending on the integration registry.
	IntegrationInstallationLookup interface {
		LookupInstallation(context.Context, predicate.Integration) (*ent.Integration, error)
		ListInstallations(context.Context, ...predicate.Integration) ([]*ent.Integration, error)
	}

	// InstalledIntegrationGetter resolves saved installations through their integration definitions.
	// Unlike IntegrationInstallationLookup, it depends on the integration registry.
	InstalledIntegrationGetter interface {
		GetInstalledIntegration(context.Context, uuid.UUID) (InstalledIntegration, error)
		ListAllInstalled(ctx context.Context, predicates ...predicate.Integration) ([]InstalledIntegration, error)
	}

	IntegrationService interface {
		IntegrationInstallationLookup
		InstalledIntegrationGetter

		ListInstallable(context.Context) ([]IntegrationDefinition, error)

		InstallNew(context.Context, string, []byte) (InstalledIntegration, error)
		ListUserInstallationTargets(ctx context.Context) ([]IntegrationInstallationTarget, error)
		InstallFromTarget(context.Context, IntegrationInstallationTarget) (InstalledIntegration, error)

		UpdateInstallation(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (InstalledIntegration, error)
		DeleteInstalled(ctx context.Context, id uuid.UUID) error

		AsInstalledIntegration(i *ent.Integration) (InstalledIntegration, error)
		GetAvailableAgentTools(context.Context, GetAvailableAiAgentToolsParams) ([]ai.Tool, error)

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
		Kind                 string
		From                 time.Time
		To                   time.Time
		SituationID          uuid.UUID
		AnalysisID           uuid.UUID
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
	CompleteOrgSetupParams struct {
		Name     string
		Timezone string
	}

	OrganizationService interface {
		Get(context.Context, predicate.Organization) (*ent.Organization, error)
		Set(context.Context, uuid.UUID, func(*ent.OrganizationMutation)) (*ent.Organization, error)
		SetPreferences(context.Context, uuid.UUID, func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error)
		CompleteOrgSetup(context.Context, uuid.UUID, CompleteOrgSetupParams) (*ent.Organization, error)
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
	AppAuthSessionCookie interface {
		Set(http.ResponseWriter, *ent.UserAuthSession)
		Get(*http.Request) (uuid.UUID, error)
		Clear(http.ResponseWriter)
	}

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
)

type (
	AiWorkflow[I, O any] interface {
		Run(context.Context, I) (O, error)
	}

	AiWorkflowRunner interface {
		ExecuteWorkflow(context.Context, string, func(context.Context) error) error
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

	InvokeAiAgentTurnParams struct {
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

	AiAgentInvoker interface {
		Config() AiAgentConfig
		DecodeSessionInput([]byte) (ValidatingInput, error)
		MakeInitialTurnInput(context.Context, *ent.AgentSession) (*AiAgentTurnInput, error)
		Invoke(context.Context, InvokeAiAgentTurnParams) (*AiAgentInvocationResult, error)
	}

	AiAgentCatalogue interface {
		GetConfigs() []AiAgentConfig
		ValidateSessionInput(string, []byte) (ValidatingInput, error)
		MakeInitialTurnInput(context.Context, *ent.AgentSession) (*AiAgentTurnInput, error)
	}

	AiAgentRuntime interface {
		AgentCatalogue() AiAgentCatalogue
		InvokeAgentTurn(context.Context, InvokeAiAgentTurnParams) (*AiAgentInvocationResult, error)
	}

	ListAiAgentSessionsParams struct {
		ent.ListParams
		Predicates []predicate.AgentSession
		Metadata   map[string]any
	}

	CreateAiAgentSessionParams struct {
		AgentName        string
		PermissionScopes []string
		Input            ValidatingInput
		Metadata         map[string]any
		Bindings         []AiAgentSessionBindingParams
	}

	AiAgentSessionBindingParams struct {
		ProviderResourceRef
		IntegrationID *uuid.UUID
		Metadata      map[string]any
	}

	ListAiAgentSessionBindingsParams struct {
		ent.ListParams
		Predicates []predicate.AgentSessionBinding
	}

	ListAiAgentTurnsParams struct {
		ent.ListParams
		Predicates []predicate.AgentTurn
	}

	ListAiAgentMessagesParams struct {
		ent.ListParams
		Predicates []predicate.AgentMessage
	}

	ListAiAgentArtifactsParams struct {
		ent.ListParams
		Predicates []predicate.AgentArtifact
	}

	RequestAiAgentTurnParams struct {
		Input *AiAgentTurnInput
	}

	AiAgentSessionService interface {
		ListAgentSessions(context.Context, ListAiAgentSessionsParams) (*ent.ListResult[ent.AgentSession], error)
		CreateAgentSession(context.Context, CreateAiAgentSessionParams) (*ent.AgentSession, error)
		GetAgentSession(context.Context, uuid.UUID) (*ent.AgentSession, error)

		RequestAgentTurn(context.Context, uuid.UUID, *RequestAiAgentTurnParams) (*ent.AgentTurn, error)

		ListAgentTurns(context.Context, ListAiAgentTurnsParams) (*ent.ListResult[ent.AgentTurn], error)
		GetAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		RetryAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)
		AbortAgentTurn(context.Context, uuid.UUID) (*ent.AgentTurn, error)

		ListAgentMessages(context.Context, ListAiAgentMessagesParams) (*ent.ListResult[ent.AgentMessage], error)
		ListAgentArtifacts(context.Context, ListAiAgentArtifactsParams) (*ent.ListResult[ent.AgentArtifact], error)

		ListAgentSessionBindings(context.Context, ListAiAgentSessionBindingsParams) (ent.AgentSessionBindings, error)
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

	RecordAlertInstanceParams struct {
		Event      *ent.NormalizedEvent
		Definition AlertDefinitionValues
		Instance   AlertInstanceValues
	}

	AlertDefinitionValues struct {
		KnowledgeEntityID        uuid.UUID
		Title                    string
		Description              string
		Definition               string
		ResolutionTimeoutSeconds *int
		IdentityGroupLabels      []string
	}

	AlertInstanceValues struct {
		InstanceID string
		Labels     map[string]string
		Summary    string
		Severity   schematypes.SignalSeverity
		Firing     bool
		StartedAt  time.Time
		EndedAt    *time.Time
	}

	ListAlertEpisodesParams struct {
		ent.ListParams
		// SituationID lists the episodes that are signals of the situation.
		SituationID uuid.UUID
	}

	SituationSignalService interface {
		// NotifySignalChanged tells situations that a signal's stored state changed. Call it inside the
		// transaction that changed it.
		NotifySignalChanged(ctx context.Context, signalEntityID uuid.UUID) error
	}

	AlertService interface {
		ListAlerts(context.Context, ListAlertsParams) (*ent.ListResult[ent.AlertDefinition], error)
		ListAlertEpisodes(context.Context, ListAlertEpisodesParams) (*ent.ListResult[ent.AlertEpisode], error)
		GetAlert(context.Context, uuid.UUID) (*ent.AlertDefinition, error)
		GetAlertInstance(context.Context, uuid.UUID) (*ent.AlertInstance, error)
		GetAlertMetrics(context.Context, GetAlertMetricsParams) (*ent.AlertMetrics, error)

		RecordAlertInstance(context.Context, RecordAlertInstanceParams) (*ent.AlertDefinition, error)
		SetAlertIdentityGroupLabels(ctx context.Context, id uuid.UUID, labels []string) (*ent.AlertDefinition, error)
		SetAlertSituationSignalAttention(ctx context.Context, id uuid.UUID, level situationsignalattention.Level) (*ent.AlertDefinition, error)
	}
)

type (
	CreateInvestigationParams struct {
		AnalysisID uuid.UUID
		Query      string
	}

	SubmitInvestigationUserInputParams struct {
		InvestigationID uuid.UUID
		Text            string
		SubmissionKey   string
	}

	ListInvestigationUserInputsParams struct {
		ent.ListParams
		InvestigationID uuid.UUID
	}

	RecordInvestigationEvidenceRevisionParams struct {
		InvestigationID uuid.UUID
		Explanation     string
		CallerKey       string
	}

	ListInvestigationEvidenceRevisionsParams struct {
		ent.ListParams
		InvestigationID uuid.UUID
	}

	InvestigationDetail struct {
		Investigation  *ent.Investigation
		Query          string
		HasPendingWork bool
		LatestTurn     *ent.AgentTurn
		ActiveTurn     *ent.AgentTurn

		// PendingEvidenceRevisions counts revisions not yet assigned to a turn.
		PendingEvidenceRevisions int
		// AutomaticUpdatesPaused is set when revisions are pending and the evidence turn limit is reached.
		AutomaticUpdatesPaused bool
		// EvidenceCurrentAsOf is when the newest revision reflected by a completed turn was recorded, or the
		// investigation's creation if none is.
		EvidenceCurrentAsOf time.Time
	}

	InvestigationService interface {
		CreateInvestigation(context.Context, CreateInvestigationParams) (*ent.Investigation, error)
		LookupInvestigation(context.Context, ...predicate.Investigation) (*ent.Investigation, error)
		ReadInvestigationDetail(context.Context, uuid.UUID) (*InvestigationDetail, error)
		UpdateInvestigation(context.Context, uuid.UUID) error

		SubmitInvestigationUserInput(context.Context, SubmitInvestigationUserInputParams) (*ent.InvestigationUserInput, error)
		ListInvestigationUserInputs(context.Context, ListInvestigationUserInputsParams) (*ent.ListResult[ent.InvestigationUserInput], error)

		RecordInvestigationEvidenceRevision(context.Context, RecordInvestigationEvidenceRevisionParams) (*ent.InvestigationEvidenceRevision, error)
		ListInvestigationEvidenceRevisions(context.Context, ListInvestigationEvidenceRevisionsParams) (*ent.ListResult[ent.InvestigationEvidenceRevision], error)
	}
)

type InvestigationReportSelection string

const (
	InvestigationReportSelectionLatest    InvestigationReportSelection = "latest"
	InvestigationReportSelectionCompleted InvestigationReportSelection = "completed"
)

type (
	InvestigationPublicationScope struct {
		InvestigationID uuid.UUID
		AgentTurnID     uuid.UUID
	}

	InvestigationFindingVersionReference struct {
		VersionID uuid.UUID
		Relation  ifvl.Relation
	}

	PublishInvestigationReportParams struct {
		Text        string
		Summary     string
		EvidenceIDs []uuid.UUID
	}

	PublishInvestigationFindingParams struct {
		Key               string
		Title             string
		Body              string
		EvidenceIDs       []uuid.UUID
		FindingReferences []InvestigationFindingVersionReference
	}

	PublishInvestigationAnswerParams struct {
		Title             string
		Body              string
		EvidenceIDs       []uuid.UUID
		FindingReferences []InvestigationFindingVersionReference
	}

	PublishInvestigationHypothesisParams struct {
		Key           string
		Title         string
		Justification string
		Status        ihv.Status
		EvidenceIDs   []uuid.UUID
	}

	ReadInvestigationReportParams struct {
		InvestigationID uuid.UUID
		Selection       InvestigationReportSelection
	}

	ListInvestigationFindingsParams struct {
		ent.ListParams
		InvestigationID uuid.UUID
	}

	ListInvestigationHypothesesParams struct {
		ent.ListParams
		InvestigationID uuid.UUID
	}

	InvestigationOutputService interface {
		PublishInvestigationReport(context.Context, InvestigationPublicationScope, PublishInvestigationReportParams) (*ent.InvestigationReport, error)
		PublishInvestigationFinding(context.Context, InvestigationPublicationScope, PublishInvestigationFindingParams) (*ent.InvestigationFindingVersion, error)
		PublishInvestigationAnswer(context.Context, InvestigationPublicationScope, PublishInvestigationAnswerParams) (*ent.InvestigationFindingVersion, error)
		PublishInvestigationHypothesis(context.Context, InvestigationPublicationScope, PublishInvestigationHypothesisParams) (*ent.InvestigationHypothesisVersion, error)

		ReadInvestigationReport(context.Context, ReadInvestigationReportParams) (*ent.InvestigationReport, error)
		GetInvestigationFindingVersion(ctx context.Context, investigationID uuid.UUID, versionID uuid.UUID) (*ent.InvestigationFindingVersion, error)
		ListInvestigationFindings(context.Context, ListInvestigationFindingsParams) (*ent.ListResult[ent.InvestigationFindingVersion], error)
		GetInvestigationHypothesisVersion(ctx context.Context, investigationID uuid.UUID, versionID uuid.UUID) (*ent.InvestigationHypothesisVersion, error)
		ListInvestigationHypotheses(context.Context, ListInvestigationHypothesesParams) (*ent.ListResult[ent.InvestigationHypothesisVersion], error)
	}
)

type (
	// SituationStage is derived from stored times: closed if closed, raised if raised, otherwise a candidate.
	SituationStage string

	ListSituationsParams struct {
		ent.ListParams
		Stages      []SituationStage
		Muted       *bool
		OpenedAfter *time.Time
	}

	ListSituationJudgmentsParams struct {
		ent.ListParams
		SituationID uuid.UUID
	}

	CreateSituationParams struct {
		Title   string
		Summary string
		// SeedEntityID is the signal that started the situation; it must be one of the groups' signals.
		SeedEntityID      uuid.UUID
		ObservationGroups []SituationObservationGroupParams
		// Raise raises the new situation; nil leaves it a candidate.
		Raise *RaiseSituationParams
	}

	SituationObservationGroupParams struct {
		Title           string
		Body            *string
		SignalEntityIDs []uuid.UUID
	}

	RaiseSituationParams struct {
		StartInvestigation bool
		Reason             string
	}

	SituationMute struct {
		Reason situation.MuteReason
	}

	SituationHold struct {
		// Until is when the hold ends; nil holds for the default length.
		Until *time.Time
	}

	CloseSituationParams struct {
		// Note is stored on the closed action. The close reason is decided from the situation's state.
		Note string
	}

	MergeSituationsParams struct {
		SourceID    uuid.UUID
		TargetID    uuid.UUID
		Explanation string
	}

	IncidentSituationLinkChanges struct {
		Added   []uuid.UUID
		Removed []uuid.UUID
	}

	ListSituationHazardAssessmentsParams struct {
		ent.ListParams
		SituationID    uuid.UUID
		SystemHazardID uuid.UUID
	}

	AddSituationHazardAssessmentParams struct {
		SituationID    uuid.UUID
		SystemHazardID uuid.UUID
		Status         sha.Status
		Summary        string
		AssessedAt     time.Time
		UserID         *uuid.UUID
		AgentTurnID    *uuid.UUID
	}

	SituationService interface {
		ListSituations(context.Context, ListSituationsParams) (*ent.ListResult[ent.Situation], error)
		GetSituation(context.Context, uuid.UUID) (*ent.Situation, error)
		CreateSituation(context.Context, CreateSituationParams) (*ent.Situation, error)

		RaiseSituation(context.Context, uuid.UUID, RaiseSituationParams) (*ent.Situation, error)
		ListSituationJudgments(context.Context, ListSituationJudgmentsParams) (*ent.ListResult[ent.SituationJudgment], error)
		MergeSituations(context.Context, MergeSituationsParams) (*ent.Situation, error)
		CloseSituation(context.Context, uuid.UUID, CloseSituationParams) (*ent.Situation, error)

		// SetSituationMute mutes the situation, or clears its mute when nil.
		SetSituationMute(context.Context, uuid.UUID, *SituationMute) (*ent.Situation, error)

		// SetSituationHold delays the situation's automatic closure, or clears the hold when nil.
		SetSituationHold(context.Context, uuid.UUID, *SituationHold) (*ent.Situation, error)

		// ListSituationEventIDs lists the events of the situation's signals.
		ListSituationEventIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)

		// SyncIncidentLinks applies an incident write's situation link changes inside its transaction.
		SyncIncidentLinks(ctx context.Context, incidentID uuid.UUID, changes IncidentSituationLinkChanges) error

		ListSituationHazardAssessments(context.Context, ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error)
		AddSituationHazardAssessment(context.Context, AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error)
	}
)

const (
	SituationStageCandidate SituationStage = "candidate"
	SituationStageRaised    SituationStage = "raised"
	SituationStageClosed    SituationStage = "closed"
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

	SetIncidentRoleAssignmentParams struct {
		IncidentID uuid.UUID
		RoleID     uuid.UUID
		UserID     uuid.UUID
	}

	ListIncidentsParams struct {
		ent.ListParams
		ResponseStates []incident.ResponseState
		SeverityId     uuid.UUID
		UserId         uuid.UUID
		OpenedAfter    time.Time
		OpenedBefore   time.Time
	}

	IncidentService interface {
		ListIncidents(context.Context, ListIncidentsParams) (*ent.ListResult[ent.Incident], error)
		Query(context.Context, predicate.Incident, func(*ent.IncidentQuery)) (*ent.Incident, error)
		Get(context.Context, predicate.Incident) (*ent.Incident, error)
		Set(context.Context, uuid.UUID, func(*ent.IncidentMutation)) (*ent.Incident, error)
		Archive(context.Context, uuid.UUID) error

		ListMilestonesForIncident(context.Context, uuid.UUID) (ent.IncidentMilestones, error)
		GetIncidentMilestone(context.Context, uuid.UUID) (*ent.IncidentMilestone, error)
		SetIncidentMilestone(context.Context, uuid.UUID, func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error)

		GetIncidentMetadata(context.Context) (*IncidentMetadata, error)
		ListIncidentTypes(context.Context) (ent.IncidentTypes, error)
		ListIncidentTags(context.Context) (ent.IncidentTags, error)
		ListIncidentSeverities(context.Context) (ent.IncidentSeverities, error)
		ListIncidentRoles(context.Context) (ent.IncidentRoles, error)

		GetIncidentSeverity(context.Context, uuid.UUID) (*ent.IncidentSeverity, error)

		SetIncidentRoleAssignment(context.Context, uuid.UUID, SetIncidentRoleAssignmentParams) (*ent.IncidentRoleAssignment, error)
		GetIncidentRoleAssignment(context.Context, uuid.UUID) (*ent.IncidentRoleAssignment, error)
		DeleteIncidentRoleAssignment(context.Context, uuid.UUID) error
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

func (EventOnIncidentUpdated) MessageName() string          { return "incident.updated.v1" }
func (EventOnIncidentMilestoneUpdated) MessageName() string { return "incident.milestone-updated.v1" }

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
	ListDiscussionThreadsParams struct {
		ent.ListParams
		AnalysisID      uuid.UUID
		RetrospectiveID uuid.UUID
		TargetKind      dt.TargetKind
		TargetID        uuid.UUID
		Resolved        *bool
	}

	ListDiscussionCommentsParams struct {
		ent.ListParams
		ThreadID uuid.UUID
		ParentID uuid.UUID
	}

	CreateDiscussionThreadParams struct {
		AnalysisID      *uuid.UUID
		RetrospectiveID *uuid.UUID
		TargetKind      *dt.TargetKind
		TargetID        *uuid.UUID
		InitialMessage  string
	}

	CreateDiscussionCommentParams struct {
		ThreadID uuid.UUID
		Content  string
	}

	UpdateDiscussionCommentParams struct {
		Content string
	}

	DiscussionService interface {
		ListThreads(context.Context, ListDiscussionThreadsParams) (*ent.ListResult[ent.DiscussionThread], error)
		GetThread(context.Context, uuid.UUID) (*ent.DiscussionThread, error)
		CreateThread(context.Context, CreateDiscussionThreadParams) (*ent.DiscussionThread, error)

		ListComments(context.Context, ListDiscussionCommentsParams) (*ent.ListResult[ent.DiscussionComment], error)
		GetComment(context.Context, uuid.UUID) (*ent.DiscussionComment, error)
		CreateComment(context.Context, CreateDiscussionCommentParams) (*ent.DiscussionComment, error)
		UpdateComment(context.Context, uuid.UUID, UpdateDiscussionCommentParams) (*ent.DiscussionComment, error)

		ReviewService
	}
)

type (
	ListReviewsParams struct {
		ent.ListParams
		RetrospectiveID uuid.UUID
		AnalysisEntryID uuid.UUID
	}

	CreateReviewRequestParams struct {
		ReviewerID uuid.UUID
		Message    *string
	}

	ReviewService interface {
		ListReviews(context.Context, ListReviewsParams) (*ent.ListResult[ent.Review], error)
		GetReview(context.Context, uuid.UUID) (*ent.Review, error)
		CreateReviewRequest(context.Context, CreateReviewRequestParams) (*ent.Review, error)
	}
)

type (
	RetrospectiveReportComposition struct {
		SelectedFindingEntries ent.SystemAnalysisEntries
		Tasks                  ent.Tasks
	}

	RetrospectiveReport struct {
		SelectedFindings []uuid.UUID
	}

	SetRetrospectiveReportParams struct {
		SelectedFindingEntryIDs []uuid.UUID
	}

	RetrospectiveService interface {
		Get(context.Context, predicate.Retrospective) (*ent.Retrospective, error)
		Set(context.Context, uuid.UUID, func(*ent.RetrospectiveMutation)) (*ent.Retrospective, error)
		CreateForIncident(context.Context, uuid.UUID) (*ent.Retrospective, error)

		GetReportComposition(context.Context, uuid.UUID) (*RetrospectiveReportComposition, error)
		SetReport(context.Context, uuid.UUID, SetRetrospectiveReportParams) (*RetrospectiveReport, error)
	}
)

type (
	ListTasksParams struct {
		ent.ListParams
		IncidentID    uuid.UUID
		OwnerID       uuid.UUID
		SourceEntryID uuid.UUID
		State         task.State
	}

	SetTaskParams struct {
		Title         *string
		Description   *string
		IncidentID    *uuid.UUID
		OwnerID       *uuid.UUID
		DueAt         *time.Time
		SourceEntryID *uuid.UUID
		TicketIDs     []uuid.UUID
	}

	TaskService interface {
		ListTasks(context.Context, ListTasksParams) (*ent.ListResult[ent.Task], error)
		GetTask(context.Context, uuid.UUID) (*ent.Task, error)
		SetTask(context.Context, uuid.UUID, SetTaskParams) (*ent.Task, error)
		ArchiveTask(context.Context, uuid.UUID) error
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
