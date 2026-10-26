package rez

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/texm/prosemirror-go"
	"golang.org/x/oauth2"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/user"

	"github.com/rezible/rezible/ent/incidentmilestone"
	ifvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	ihv "github.com/rezible/rezible/ent/investigationhypothesisversion"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/review"
	"github.com/rezible/rezible/ent/situation"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/ent/situationsignalattention"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	"github.com/rezible/rezible/ent/task"
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
		// IsForeignKeyError reports whether an error violates one of the named foreign key constraints.
		IsForeignKeyError(error, ...string) bool
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
		// MergeEntities merges the two entities holding a proposal's values into the older one, in the caller's
		// transaction when there is one. It refuses, with ErrConflict, entities that are not services or
		// repositories of one kind.
		MergeEntities(context.Context, MergeEntitiesParams) (*MergeEntitiesResult, error)
	}

	MergeEntitiesParams struct {
		// ProposalID is the proposal being confirmed. Each moved alias is marked with it.
		ProposalID uuid.UUID
		// EntityIDs are the two entities the caller found holding the proposal's values.
		EntityIDs [2]uuid.UUID
	}

	MergeEntitiesResult struct {
		// SurvivorID is the entity that holds both values.
		SurvivorID uuid.UUID
		// Merged is false when the values were already on one entity, so nothing was merged.
		Merged bool
	}
)

type (
	ListKnowledgeEntitiesParams struct {
		// ListParams.Search keeps entities whose display name or a linking value contains it, ignoring case.
		ent.ListParams
		Predicates []predicate.KnowledgeEntity
		// Categories, when set, limits the entities to these categories.
		Categories []kne.Category
		// LinkingAttribute, when set, limits the entities to those holding that linking attribute, and
		// LinkingValueContains, when also set, to those whose value contains it.
		LinkingAttribute, LinkingValueContains string
		// WithLinkingAttributes loads each entity's linking attributes.
		WithLinkingAttributes bool
	}

	// KnowledgeEntityExpand names what GetEntity loads beyond the entity and its aliases.
	KnowledgeEntityExpand struct {
		// AliasEvidence loads every evidence row on each alias. SummarizeEvidence is cheaper when only counts
		// and the latest evidence are needed.
		AliasEvidence     bool
		LinkingAttributes bool
	}

	ListKnowledgeRelationshipsParams struct {
		ent.ListParams
		Predicates []predicate.KnowledgeRelationship
		// EntityID lists the entity's relationships in either direction.
		EntityID uuid.UUID
		// NeighbourCategories and ExcludedNeighbourCategories filter on the category of the other endpoint
		// of the entity's relationships. They need EntityID.
		NeighbourCategories         []kne.Category
		ExcludedNeighbourCategories []kne.Category
		// NewestEvidenceFirst orders by the effective time of each relationship's latest evidence, then ID,
		// instead of by ID.
		NewestEvidenceFirst bool
		Expand              KnowledgeRelationshipExpand
	}

	// KnowledgeRelationshipExpand names what ListRelationships loads beyond the relationships and their aliases.
	KnowledgeRelationshipExpand struct {
		Endpoints bool
	}

	// KnowledgeEvidenceExpand names what GetEvidence loads beyond the evidence and its alias.
	KnowledgeEvidenceExpand struct {
		// Subject loads the alias's entity, or its relationship and the relationship's endpoints.
		Subject bool
	}

	SummarizeKnowledgeEvidenceParams struct {
		ent.ListParams
		// EntityID summarizes the evidence on the entity and on its relationships. The entity must exist.
		EntityID uuid.UUID
		AliasIDs []uuid.UUID
		// Start and End bound effective time. A zero time leaves that side unbounded.
		Start, End time.Time
		// ByAssertion summarizes each assertion on an alias separately.
		ByAssertion bool
		Expand      KnowledgeEvidenceSummaryExpand
	}

	// KnowledgeEvidenceSummaryExpand names what SummarizeEvidence loads beyond the summaries.
	KnowledgeEvidenceSummaryExpand struct {
		// Alias loads each summary's alias, and a relationship alias's relationship and its endpoints.
		Alias bool
	}

	// KnowledgeEvidenceSummary describes the evidence recorded against one alias, or one alias and assertion.
	KnowledgeEvidenceSummary struct {
		SubjectAliasID uuid.UUID
		// Alias is loaded when asked for.
		Alias *ent.KnowledgeSubjectAlias
		// Assertion is set when summarizing by assertion.
		Assertion   string
		Count       int
		EarliestAt  time.Time
		LatestAt    time.Time
		LatestID    uuid.UUID
		LatestState KnowledgeSubjectState
	}

	// KnowledgeEntityAlias is one of an entity's aliases, with the name its provider uses for the entity and when
	// Rezible last recorded evidence from it.
	KnowledgeEntityAlias struct {
		// Alias is loaded with the proposal whose merge moved it, if any, and that proposal's decisions and their
		// users.
		Alias *ent.KnowledgeSubjectAlias
		// IntegrationDisplayName is the display name of the integration named by the alias's provider, empty
		// when none is registered by that name.
		IntegrationDisplayName string
		// Name is the display name the alias's newest evidence states, or its resource ref when it states none.
		Name string
		// LastRecordedAt is the effective time of the alias's newest evidence, zero when it has none.
		LastRecordedAt time.Time
	}

	// KnowledgeEntityWithAliases is an entity with its aliases, most recently recorded first.
	KnowledgeEntityWithAliases struct {
		Entity  *ent.KnowledgeEntity
		Aliases []KnowledgeEntityAlias
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
		GetEntity(context.Context, uuid.UUID, KnowledgeEntityExpand) (*ent.KnowledgeEntity, error)

		ListRelationships(context.Context, ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error)
		GetRelationship(context.Context, uuid.UUID) (*ent.KnowledgeRelationship, error)

		ListSubjectAliases(context.Context, ListKnowledgeSubjectAliasesParams) (*ent.ListResult[ent.KnowledgeSubjectAlias], error)
		GetSubjectAlias(context.Context, uuid.UUID) (*ent.KnowledgeSubjectAlias, error)

		ListEvidence(context.Context, ListKnowledgeEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error)
		GetEvidence(context.Context, uuid.UUID, KnowledgeEvidenceExpand) (*ent.KnowledgeEvidence, error)
		// SummarizeEvidence groups evidence by alias, and by assertion when asked, newest group first.
		SummarizeEvidence(context.Context, SummarizeKnowledgeEvidenceParams) (*ent.ListResult[KnowledgeEvidenceSummary], error)

		SelectGraphEntities(context.Context, SelectKnowledgeGraphEntitiesParams) (*KnowledgeGraphEntitiesPage, error)
		ExpandGraphRelationships(context.Context, ExpandKnowledgeGraphRelationshipsParams) (*KnowledgeGraphRelationshipsPage, error)

		// ResolveStructure returns, for each entity, the entities in TargetCategories that represent it: itself
		// if it is in one, otherwise its nearest ancestors in one, climbing the structure hierarchy at most
		// MaxDepth steps. Entities represented by nothing are omitted.
		ResolveStructure(context.Context, ResolveStructureParams) (map[uuid.UUID][]uuid.UUID, error)

		// ListRelatedEvents lists event entities of one kind that relate to an entity, newest first.
		ListRelatedEvents(context.Context, ListRelatedEventsParams) (*RelatedEvents, error)
	}

	// KnowledgeEntityAliasService reads entities' aliases with what people need to review them.
	KnowledgeEntityAliasService interface {
		// ListEntityAliases lists, by entity ID, each entity's aliases, most recently recorded first.
		ListEntityAliases(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]KnowledgeEntityAlias, error)
		// ListEntitiesWithAliases lists entities as KnowledgeGraphQueryService.ListEntities does, with their aliases.
		ListEntitiesWithAliases(context.Context, ListKnowledgeEntitiesParams) (*ent.ListResult[KnowledgeEntityWithAliases], error)
		// GetEntityWithAliases gets an entity as KnowledgeGraphQueryService.GetEntity does, with its aliases.
		GetEntityWithAliases(context.Context, uuid.UUID, KnowledgeEntityExpand) (*KnowledgeEntityWithAliases, error)
	}

	ListRelatedEventsParams struct {
		// EntityID is the related entity; the events are relationship sources.
		EntityID uuid.UUID
		// Predicate relates the events to the entity, for example impacts or touches.
		Predicate knr.Predicate
		// Kind is the event entities' kind, for example deployment.
		Kind string
		// From and To bound state_effective_at. A zero time leaves that side unbounded.
		From, To                   time.Time
		FromInclusive, ToInclusive bool
		// PropertyEquals are state properties that must equal these values, compared as text.
		PropertyEquals map[string]string
		// PropertyAbsent are state properties that must be absent. An absent property never matches
		// PropertyEquals.
		PropertyAbsent []string
		Limit          int
	}

	RelatedEvents struct {
		Events []*ent.KnowledgeEntity
		// Cut is set when more events matched than Limit.
		Cut bool
	}

	ListProposalsParams struct {
		ent.ListParams
		Statuses []kp.Status
		// EntityIDs keeps proposals naming a linking value one of the entities holds now.
		EntityIDs []uuid.UUID
	}

	// UnconfirmedMatch is a pending proposal naming one of an entity's linking values, seen from that entity:
	// the other value may belong to the same entity.
	UnconfirmedMatch struct {
		ProposalID uuid.UUID
		Rationale  string
		// Value is the proposal's other value, and Entity the entity that holds it now.
		Value  string
		Entity *ent.KnowledgeEntity
	}

	// KnowledgeLinkingValue is one value of a linking attribute, such as service.name checkout.
	KnowledgeLinkingValue struct {
		Attribute string
		Value     string
	}

	// KnowledgeValueHolders are the entities holding linking values now. A value no entity holds is absent.
	KnowledgeValueHolders map[KnowledgeLinkingValue]*ent.KnowledgeEntity

	// KnowledgeProposalValue is one of a proposal's values, with the entity holding it now.
	KnowledgeProposalValue struct {
		Value string
		// Holder is the entity holding the value, with its aliases; nil when no entity holds it.
		Holder *KnowledgeEntityWithAliases
	}

	// KnowledgeProposalWithValues is a proposal with the entities holding its values.
	KnowledgeProposalWithValues struct {
		Proposal       *ent.KnowledgeProposal
		ValueA, ValueB KnowledgeProposalValue
		// EvidenceIntegrationDisplayName is the display name of the integration named by the cited evidence's
		// alias, empty when none is registered by that name.
		EvidenceIntegrationDisplayName string
	}

	// KnowledgeProposalService reads the identity matches Rezible proposes, and records people's decisions on them.
	KnowledgeProposalService interface {
		// ListProposals lists proposals, newest first unless OrderAsc, each with the evidence it cites and that
		// evidence's alias, and its decisions with their users, oldest first.
		ListProposals(context.Context, ListProposalsParams) (*ent.ListResult[ent.KnowledgeProposal], error)
		// ListProposalsWithValues lists proposals as ListProposals does, with the entities holding their values.
		ListProposalsWithValues(context.Context, ListProposalsParams) (*ent.ListResult[KnowledgeProposalWithValues], error)
		// GetProposalWithValues gets a proposal in any status, loaded as ListProposals loads them, with the entities holding its values.
		GetProposalWithValues(context.Context, uuid.UUID) (*KnowledgeProposalWithValues, error)
		// ListImpliedProposals lists the other proposals that confirming the proposal would also confirm, because
		// the merge would put both their values on one entity. It is empty when confirming would merge nothing.
		ListImpliedProposals(context.Context, *ent.KnowledgeProposal) ([]*ent.KnowledgeProposal, error)
		// ConfirmProposal confirms a pending or rejected proposal as the user in the context, merging the two
		// entities holding its values. Confirming is final. It returns the proposal as GetProposalWithValues does.
		ConfirmProposal(context.Context, uuid.UUID) (*KnowledgeProposalWithValues, error)
		// RejectProposal rejects a pending proposal as the user in the context. It returns the proposal as
		// GetProposalWithValues does.
		RejectProposal(context.Context, uuid.UUID) (*KnowledgeProposalWithValues, error)
		// ListUnconfirmedMatches lists, by entity ID, the pending proposals naming each entity, oldest first.
		// A proposal whose other value no entity holds, or the same entity holds, is left out.
		ListUnconfirmedMatches(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]UnconfirmedMatch, error)
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

// Of returns the entities holding the proposal's two values, nil for a value no entity holds.
func (h KnowledgeValueHolders) Of(p *ent.KnowledgeProposal) (a, b *ent.KnowledgeEntity) {
	a = h[KnowledgeLinkingValue{Attribute: p.Attribute, Value: p.ValueA}]
	b = h[KnowledgeLinkingValue{Attribute: p.Attribute, Value: p.ValueB}]
	return a, b
}

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
		Kinds      []sae.Kind
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
		IngestMany(context.Context, []ProviderEvent) error
		SyncEvents(context.Context, ProviderEventQuerier, ProviderEventSourceCursors) ProviderEventSyncResult
	}

	ProviderEventIngester interface {
		ProcessProviderEvent(context.Context, ProviderEvent) ([]uuid.UUID, error)
		ProjectNormalizedEvent(ctx context.Context, eventID uuid.UUID) error
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
		ValidateUserSettings([]byte) error
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
		// GetIntegrationDisplayName returns the display name of the named registered integration, or empty
		// when none is registered by that name.
		GetIntegrationDisplayName(ctx context.Context, name string) string

		InstallNew(context.Context, string, []byte) (InstalledIntegration, error)
		ListUserInstallationTargets(ctx context.Context) ([]IntegrationInstallationTarget, error)
		InstallFromTarget(context.Context, IntegrationInstallationTarget) (InstalledIntegration, error)

		UpdateInstallation(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (InstalledIntegration, error)
		DeleteInstalled(ctx context.Context, id uuid.UUID) error
		// IssueWebhookToken replaces the installation's webhook token and returns its full webhook URL. The URL
		// is a secret and is only available here.
		IssueWebhookToken(ctx context.Context, id uuid.UUID) (string, error)

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
		IsIncidentManagementEnabled(context.Context) (bool, error)
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
		User ent.User
		Org  ent.Organization
		Role user.Role
		// RefreshToken refreshes the session with the provider. Empty when the provider issued none.
		RefreshToken string
		ExpiresAt    time.Time
	}

	// AuthSessionRefreshExchange sends a refresh token to the provider and returns the refreshed provider session.
	// It returns errs.ErrAuthSessionInvalid when the provider refuses the token.
	AuthSessionRefreshExchange func(ctx context.Context, refreshToken string) (*UserAuthProviderSession, error)

	AuthSessionService interface {
		CreateFromUserAuthResponse(context.Context, *UserAuthProviderSession) (*ent.UserAuthSession, error)
		CreateForToken(context.Context, string) (*ent.UserAuthSession, error)
		Refresh(context.Context, uuid.UUID, AuthSessionRefreshExchange) (*ent.UserAuthSession, error)
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
		// ContinueFromState resumes committed messages without sending Input again.
		// Input remains available to invocation middleware as the turn's assignment.
		ContinueFromState bool
		OnChunk           func(AiAgentTurnChunk)
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

	StartAgentProviderReadParams struct {
		AgentTurnID     uuid.UUID
		ToolName        string
		ServiceEntityID uuid.UUID
		// IntegrationID is the installation that runs the read.
		IntegrationID uuid.UUID
		// Input is the tool's input as the agent sent it.
		Input json.RawMessage
	}

	FinishAgentProviderReadParams struct {
		ID uuid.UUID
		// Read describes the read that ran, when the provider got as far as building it.
		Read json.RawMessage
		// Error is why the read failed; empty when it succeeded.
		Error string
	}

	// AgentProviderReadService records the provider reads agents make and holds each turn to its budget.
	AgentProviderReadService interface {
		// StartProviderRead records a provider read as running, unless the turn's reads from all its attempts
		// already use its budget, which is refused. It returns the read and how many the turn has used,
		// including this one.
		StartProviderRead(context.Context, StartAgentProviderReadParams) (*ent.AgentProviderRead, int, error)
		// FinishProviderRead records the outcome of a started read.
		FinishProviderRead(context.Context, FinishAgentProviderReadParams) error
		// CountTurnProviderReads counts the turn's provider reads from all its attempts, including any left
		// running.
		CountTurnProviderReads(ctx context.Context, turnID uuid.UUID) (int, error)
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

	AlertService interface {
		SituationSignalSource

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
	ListServiceChangesParams struct {
		// ServiceEntityID is a container/service entity.
		ServiceEntityID uuid.UUID
		Start, End      time.Time
		// Environment is raw and normalized by the service. Empty lists every environment.
		Environment string
	}

	ServiceChanges struct {
		// Deployments are the service's deployments in the window, most recent first.
		Deployments []ServiceDeployment
		// Limits describes, in plain words, each limit that cut the result.
		Limits []string
	}

	// ServiceDeployment is a deployment with the merged changes linked to it.
	ServiceDeployment struct {
		// Deployment is the deployment's event entity.
		Deployment  *ent.KnowledgeEntity
		DeployedAt  time.Time
		Environment string
		Status      string
		Sha         string
		Version     string
		URL         string
		// RepositoryEntityID is the repository the deployment touches; nil when it names none.
		RepositoryEntityID *uuid.UUID
		// RepositoryName is the repository's display name, empty when the deployment names none.
		RepositoryName string
		// RepositorySeenByCodeForge is whether a code forge has reported the repository. Merged changes are
		// only linked when one has.
		RepositorySeenByCodeForge bool
		// PreviousDeployedAt is when the previous successful deployment of the same service, environment and
		// repository was deployed, if there is one.
		PreviousDeployedAt *time.Time
		MergedChanges      []LinkedMergedChange
		// MergedChangesCut is set when the merged change limit cut MergedChanges.
		MergedChangesCut bool
	}

	// LinkedMergedChange is a merged change linked to a deployment.
	LinkedMergedChange struct {
		// Change is the code change's event entity; its display name is the change's title.
		Change         *ent.KnowledgeEntity
		MergedAt       time.Time
		MergeCommitSha string
		BaseRef        string
		// IntoDefaultBranch is whether the change was merged into the repository's default branch.
		IntoDefaultBranch bool
		Number            int
		URL               string
		Author            string
		Link              MergedChangeLink
	}

	// MergedChangeLink is how a merged change was linked to a deployment.
	MergedChangeLink string

	ChangesService interface {
		// ListServiceChanges lists a service's deployments in a window with the merged changes linked to each.
		ListServiceChanges(ctx context.Context, params ListServiceChangesParams) (*ServiceChanges, error)
	}
)

const (
	// MergedChangeLinkMergeCommit links a change whose merge commit is the deployed commit.
	MergedChangeLinkMergeCommit MergedChangeLink = "merge_commit"
	// MergedChangeLinkMergedSincePrevious links a change merged into the default branch after the previous
	// deployment.
	MergedChangeLinkMergedSincePrevious MergedChangeLink = "merged_since_previous_deployment"
	// MergedChangeLinkMergedInLookback links a change merged into the default branch within the lookback
	// before a deployment with no previous deployment.
	MergedChangeLinkMergedInLookback MergedChangeLink = "merged_in_lookback"
)

const (
	// MaxBriefServices is how many of a situation's services its brief shows.
	MaxBriefServices = 10
	// BriefChangesLeadIn is how long before a situation opened its brief lists deployments.
	BriefChangesLeadIn = 24 * time.Hour
)

type (
	// SituationBrief is what code states about a situation: its services, their changes and what Rezible
	// cannot see.
	SituationBrief struct {
		// Services are the situation's services, matching ones first, then by name, at most MaxBriefServices.
		Services []BriefService
		Changes  SituationBriefChanges
		Unseen   []SituationBriefUnseen
	}

	// BriefService is one of a situation's services in its brief.
	BriefService struct {
		Entity             *ent.KnowledgeEntity
		UnconfirmedMatches []UnconfirmedMatch
	}

	SituationBriefChanges struct {
		// Available is whether any installation reports deployments; without one, nothing is looked up and
		// UnavailableReason says why.
		Available         bool
		UnavailableReason string
		// Start is BriefChangesLeadIn before the situation opened, and End when it closed, or now.
		Start, OpenedAt, End time.Time
		// Services holds the changes of each of Services, in the same order.
		Services []SituationBriefServiceChanges
	}

	SituationBriefServiceChanges struct {
		ServiceEntityID uuid.UUID
		// Deployments are the deployments after the situation opened, then those before, each newest first.
		Deployments []SituationBriefDeployment
		// Limits describes, in plain words, each limit that cut the lookups.
		Limits []string
	}

	SituationBriefDeployment struct {
		ServiceDeployment
		AfterOpened bool
		// LatestEvidenceID is the deployment's latest knowledge evidence, if any.
		LatestEvidenceID *uuid.UUID
		// MergedChangesUnlinked says why merged changes cannot be linked; empty when they can.
		MergedChangesUnlinked SituationBriefUnlinkedReason
	}

	SituationBriefUnlinkedReason string

	SituationBriefUnseen struct {
		Kind SituationBriefUnseenKind
		// Text is written for people.
		Text string
	}

	SituationBriefUnseenKind string

	// SituationLatestDeployment is the newest deployment of one of a situation's services before it opened.
	SituationLatestDeployment struct {
		Service     *ent.KnowledgeEntity
		DeployedAt  time.Time
		Environment string
		Status      string
		Version     string
		Sha         string
	}

	SituationBriefService interface {
		GetSituationBrief(ctx context.Context, situationID uuid.UUID) (*SituationBrief, error)
		// ListBriefServices returns the services the situation's brief lists, in its order, without their changes.
		ListBriefServices(ctx context.Context, situationID uuid.UUID) ([]*ent.KnowledgeEntity, error)
		// LatestDeploymentBefore returns the newest deployment of the brief's services from
		// BriefChangesLeadIn before the situation opened until it opened, of any status, or nil.
		LatestDeploymentBefore(ctx context.Context, situationID uuid.UUID) (*SituationLatestDeployment, error)
	}
)

const (
	SituationBriefUnlinkedNoRepository       SituationBriefUnlinkedReason = "no_repository"
	SituationBriefUnlinkedRepositoryNotForge SituationBriefUnlinkedReason = "repository_not_from_code_forge"

	SituationBriefUnseenSourceUnavailable SituationBriefUnseenKind = "source_unavailable"
	SituationBriefUnseenServicesCut       SituationBriefUnseenKind = "services_cut"
	SituationBriefUnseenNotRead           SituationBriefUnseenKind = "not_read"
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
		// CitedItems are items of cited provider reads, by read ID, to merge into each read's highlight.
		CitedItems map[uuid.UUID][]schematypes.ProviderReadItem
		// NextChecks are at most rezai.MaxNextChecks, most useful first.
		NextChecks []schematypes.InvestigationNextCheck
	}

	PublishInvestigationFindingParams struct {
		Key         string
		Title       string
		Body        string
		EvidenceIDs []uuid.UUID
		// CitedItems are items of cited provider reads, by read ID, to merge into each read's highlight.
		CitedItems        map[uuid.UUID][]schematypes.ProviderReadItem
		FindingReferences []InvestigationFindingVersionReference
	}

	PublishInvestigationAnswerParams struct {
		Title       string
		Body        string
		EvidenceIDs []uuid.UUID
		// CitedItems are items of cited provider reads, by read ID, to merge into each read's highlight.
		CitedItems        map[uuid.UUID][]schematypes.ProviderReadItem
		FindingReferences []InvestigationFindingVersionReference
	}

	PublishInvestigationHypothesisParams struct {
		Key           string
		Title         string
		Justification string
		Status        ihv.Status
		EvidenceIDs   []uuid.UUID
		// CitedItems are items of cited provider reads, by read ID, to merge into each read's highlight.
		CitedItems map[uuid.UUID][]schematypes.ProviderReadItem
	}

	GetInvestigationReportParams struct {
		InvestigationID uuid.UUID
		Selection       InvestigationReportSelection
	}

	// InvestigationOutputVersion holds one output version: exactly one field is set.
	InvestigationOutputVersion struct {
		Report     *ent.InvestigationReport
		Finding    *ent.InvestigationFindingVersion
		Hypothesis *ent.InvestigationHypothesisVersion
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

		GetInvestigationReport(context.Context, GetInvestigationReportParams) (*ent.InvestigationReport, error)

		ListInvestigationFindings(context.Context, ListInvestigationFindingsParams) (*ent.ListResult[ent.InvestigationFindingVersion], error)
		GetInvestigationFindingVersion(ctx context.Context, investigationID uuid.UUID, versionID uuid.UUID) (*ent.InvestigationFindingVersion, error)

		ListInvestigationHypotheses(context.Context, ListInvestigationHypothesesParams) (*ent.ListResult[ent.InvestigationHypothesisVersion], error)
		GetInvestigationHypothesisVersion(ctx context.Context, investigationID uuid.UUID, versionID uuid.UUID) (*ent.InvestigationHypothesisVersion, error)

		// GetInvestigationOutputVersion returns the investigation's report, finding version or hypothesis version with the ID, whichever it is.
		GetInvestigationOutputVersion(ctx context.Context, investigationID uuid.UUID, versionID uuid.UUID) (*InvestigationOutputVersion, error)
	}
)

type (
	SituationSignalKind string

	SituationSignal struct {
		EntityID             uuid.UUID
		Kind                 SituationSignalKind
		Title                string
		SourceEntityID       *uuid.UUID
		SignalAttention      situationsignalattention.Level
		SignalAttentionSetAt *time.Time
		Severity             schematypes.SignalSeverity
		StartedAt            time.Time
		FinishedAt           *time.Time
		Active               bool
		LastActivityAt       time.Time
		EventIDs             []uuid.UUID
		Revision             int
		Alert                *schematypes.SituationAlertFacts
	}

	LoadSituationSignalsOptions struct {
		History bool
	}

	SituationSignalSource interface {
		SituationSignalKind() SituationSignalKind
		LoadSignals(ctx context.Context, entityIDs []uuid.UUID, opts LoadSituationSignalsOptions) ([]SituationSignal, error)
	}

	// SituationStage is derived from stored times: closed if closed, raised if raised, otherwise a candidate.
	SituationStage string

	ListSituationsParams struct {
		ent.ListParams
		Stages      []SituationStage
		Muted       *bool
		OpenedAfter *time.Time
		// InvestigationID lists the situation investigated by that investigation.
		InvestigationID uuid.UUID
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
		GetSituation(context.Context, predicate.Situation) (*ent.Situation, error)
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

		// LoadSignals loads signals of any kind from their sources, in the order of entityIDs. An unknown entity
		// or a kind without a source is invalid input.
		LoadSignals(ctx context.Context, entityIDs []uuid.UUID, opts LoadSituationSignalsOptions) ([]SituationSignal, error)
	}

	SituationSignalService interface {
		// NotifySignalChanged tells situations that a signal's stored state changed
		NotifySignalChanged(ctx context.Context, signalEntityID uuid.UUID) error
	}
)

// SituationNoticeService posts a notice of each raised situation to the chat channel of every installation
// that sends them, and edits it as the situation changes.
type SituationNoticeService interface {
	// RetrySituationNotices marks the situation's failed notices pending and asks for them to be delivered
	// again.
	RetrySituationNotices(ctx context.Context, situationID uuid.UUID) error
}

// EventOnSituationUpdated says that a situation, or what is shown about it, may have changed. It carries
// only the ID: it is published inside the writing transaction and may be delivered before that commits, so
// handlers read nothing from it and schedule work that reads the situation later.
type EventOnSituationUpdated struct {
	SituationID uuid.UUID
}

func (EventOnSituationUpdated) MessageName() string { return "situation.updated.v1" }

const SignalKindAlertEpisode SituationSignalKind = "alert_episode"

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

	ListIncidentsParams struct {
		ent.ListParams
		// Resolved filters by whether the incident has a resolution milestone.
		Resolved     *bool
		SeverityId   uuid.UUID
		UserId       uuid.UUID
		OpenedAfter  time.Time
		OpenedBefore time.Time
	}

	// UpdateIncidentParams leaves a nil value unchanged and replaces a set one.
	UpdateIncidentParams struct {
		Title          *string
		Summary        *string
		SeverityID     *uuid.UUID
		ClearSeverity  bool
		TypeID         *uuid.UUID
		ClearType      bool
		FieldOptionIDs *[]uuid.UUID
		TagIDs         *[]uuid.UUID
	}

	// RecordObservedIncidentParams describe an incident observed from a provider, identified by its knowledge entity.
	RecordObservedIncidentParams struct {
		KnowledgeEntityID uuid.UUID
		Title             string
		Summary           string
		SeverityID        uuid.UUID
		TypeID            uuid.UUID
		OpenedAt          time.Time
	}

	// IncidentDetailsService reads incidents and writes what does not depend on incident management being on.
	IncidentDetailsService interface {
		GetIncident(context.Context, predicate.Incident) (*ent.Incident, error)
		ListIncidents(context.Context, ListIncidentsParams) (*ent.ListResult[ent.Incident], error)

		GetIncidentMetadata(context.Context) (*IncidentMetadata, error)
		ListIncidentTypes(context.Context) (ent.IncidentTypes, error)
		ListIncidentTags(context.Context) (ent.IncidentTags, error)
		ListIncidentSeverities(context.Context) (ent.IncidentSeverities, error)
		ListIncidentRoles(context.Context) (ent.IncidentRoles, error)

		ListMilestones(ctx context.Context, incidentID uuid.UUID) (ent.IncidentMilestones, error)
		GetMilestone(context.Context, uuid.UUID) (*ent.IncidentMilestone, error)
		SetMilestoneMetadata(ctx context.Context, id uuid.UUID, metadata map[string]string) error

		UpdateIncident(context.Context, uuid.UUID, UpdateIncidentParams) (*ent.Incident, error)
		LinkSituation(ctx context.Context, incidentID uuid.UUID, situationID uuid.UUID) error
		UnlinkSituation(ctx context.Context, incidentID uuid.UUID, situationID uuid.UUID) error
		ArchiveIncident(context.Context, uuid.UUID) error
		SetChatChannel(ctx context.Context, id uuid.UUID, channelID string) error
		SetChatDetailsMessage(ctx context.Context, id uuid.UUID, messageID string) error
		RecordObservedIncident(context.Context, RecordObservedIncidentParams) (uuid.UUID, error)
	}

	DeclareIncidentParams struct {
		Title          string
		Summary        string
		SeverityID     *uuid.UUID
		TypeID         *uuid.UUID
		FieldOptionIDs []uuid.UUID
		TagIDs         []uuid.UUID
		SituationIDs   []uuid.UUID
		DeclaredByID   uuid.UUID
	}

	AddIncidentMilestoneParams struct {
		IncidentID uuid.UUID
		Kind       incidentmilestone.Kind
		// Timestamp defaults to now.
		Timestamp   time.Time
		Description string
		// UserID is nil for a milestone recorded by Rezible.
		UserID *uuid.UUID
		// Source is where the milestone was recorded, such as web.
		Source string
	}

	// UpdateIncidentMilestoneParams leaves a nil value unchanged. A milestone's kind never changes.
	UpdateIncidentMilestoneParams struct {
		Timestamp   *time.Time
		Description *string
	}

	// IncidentManagementService runs incident response. Every write needs incident management to be on.
	IncidentManagementService interface {
		Declare(context.Context, DeclareIncidentParams) (*ent.Incident, error)
		AddMilestone(context.Context, AddIncidentMilestoneParams) (*ent.IncidentMilestone, error)
		UpdateMilestone(context.Context, uuid.UUID, UpdateIncidentMilestoneParams) (*ent.IncidentMilestone, error)
		SetRoleAssignment(ctx context.Context, incidentID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) (*ent.IncidentRoleAssignment, error)
		RemoveRoleAssignment(ctx context.Context, assignmentID uuid.UUID) error
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
		// TargetKind, if set, lists threads with that kind of target; TargetID, which needs it, one target.
		TargetKind DiscussionThreadTargetKind
		TargetID   uuid.UUID
		Resolved   *bool
	}

	// DiscussionThreadTargetKind is the kind of thing a discussion thread is about, within its owner.
	DiscussionThreadTargetKind string

	ListDiscussionCommentsParams struct {
		ent.ListParams
		ThreadID uuid.UUID
		ParentID uuid.UUID
	}

	CreateDiscussionThreadParams struct {
		AnalysisID      *uuid.UUID
		RetrospectiveID *uuid.UUID
		TargetKind      *DiscussionThreadTargetKind
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
	}
)

const (
	DiscussionThreadTargetFinding               DiscussionThreadTargetKind = "finding"
	DiscussionThreadTargetEntry                 DiscussionThreadTargetKind = "entry"
	DiscussionThreadTargetKnowledgeEntity       DiscussionThreadTargetKind = "knowledge_entity"
	DiscussionThreadTargetKnowledgeRelationship DiscussionThreadTargetKind = "knowledge_relationship"
	DiscussionThreadTargetNormalizedEvent       DiscussionThreadTargetKind = "normalized_event"
)

var DiscussionThreadTargetKinds = []DiscussionThreadTargetKind{
	DiscussionThreadTargetFinding,
	DiscussionThreadTargetEntry,
	DiscussionThreadTargetKnowledgeEntity,
	DiscussionThreadTargetKnowledgeRelationship,
	DiscussionThreadTargetNormalizedEvent,
}

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

	SetReviewDecisionParams struct {
		ReviewerID uuid.UUID
		Decision   review.State
		Feedback   string
	}

	RetrospectiveService interface {
		GetRetrospective(context.Context, predicate.Retrospective) (*ent.Retrospective, error)
		SetOwner(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error)
		RequestReview(ctx context.Context, id uuid.UUID, reviewerIDs []uuid.UUID) (*ent.Retrospective, error)
		AddReviewer(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error)
		RemoveReviewer(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error)
		SetReviewDecision(ctx context.Context, id uuid.UUID, params SetReviewDecisionParams) (*ent.Retrospective, error)
		CloseReview(ctx context.Context, id uuid.UUID) (*ent.Retrospective, error)
		SkipReview(ctx context.Context, id uuid.UUID, reason string) (*ent.Retrospective, error)
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
