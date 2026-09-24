package demoprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	"github.com/rezible/rezible/pkg/execution"
)

func SeedDemoData(
	ctx context.Context,
	knowledge rez.KnowledgeGraphQueryService,
	incidents rez.IncidentService,
	retrospectives rez.RetrospectiveService,
	systemAnalysis rez.SystemAnalysisService,
) error {
	s := &dataSeeder{
		knowledge,
		incidents,
		retrospectives,
		systemAnalysis,
	}
	return s.seedData(ctx)
}

type dataSeeder struct {
	knowledge      rez.KnowledgeGraphQueryService
	incidents      rez.IncidentService
	retrospectives rez.RetrospectiveService
	systemAnalysis rez.SystemAnalysisService
}

func (d *dataSeeder) seedData(ctx context.Context) error {
	inc, lookupErr := d.lookupIncident(ctx)
	if lookupErr != nil {
		return lookupErr
	}

	if milestonesErr := d.createMilestones(ctx, inc); milestonesErr != nil {
		return milestonesErr
	}

	if retrospectiveErr := d.createRetrospective(ctx, inc); retrospectiveErr != nil {
		return retrospectiveErr
	}

	return nil
}

func (d *dataSeeder) lookupAlias(ctx context.Context, resourceRef string) (*ent.KnowledgeSubjectAlias, error) {
	params := rez.ListKnowledgeSubjectAliasesParams{
		Predicates: []predicate.KnowledgeSubjectAlias{
			ksa.Provider(providerName),
			ksa.ProviderNamespace(providerName),
			ksa.ProviderResourceRef(resourceRef),
		},
	}
	aliases, queryErr := d.knowledge.ListSubjectAliases(ctx, params)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("lookup demo knowledge alias %q: %w", resourceRef, queryErr)
	} else if len(aliases.Data) == 0 {
		return nil, nil
	} else if len(aliases.Data) > 1 {
		return nil, fmt.Errorf("multiple demo knowledge aliases found for %q", resourceRef)
	}
	return aliases.Data[0], nil
}

func (d *dataSeeder) lookupIncident(ctx context.Context) (*ent.Incident, error) {
	alias, aliasErr := d.lookupAlias(ctx, "demo:incident:checkout-search-timeouts")
	if aliasErr != nil {
		return nil, aliasErr
	} else if alias == nil {
		return nil, nil
	} else if alias.SubjectKind != ksa.SubjectKindEntity || alias.EntityID == nil {
		return nil, fmt.Errorf("demo incident alias does not reference an entity")
	}
	return d.incidents.Get(ctx, incident.KnowledgeEntityID(*alias.EntityID))
}

func (d *dataSeeder) createMilestones(ctx context.Context, inc *ent.Incident) error {
	userID, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return fmt.Errorf("execution user ID not found in context")
	}
	setOpenedMilestone := func(m *ent.IncidentMilestoneMutation) {
		m.SetIncidentID(inc.ID)
		m.SetUserID(userID)
		m.SetKind("opened")
		m.SetTimestamp(inc.OpenedAt)
	}
	if _, milestoneErr := d.incidents.SetIncidentMilestone(ctx, uuid.Nil, setOpenedMilestone); milestoneErr != nil {
		return fmt.Errorf("create demo incident opened milestone: %w", milestoneErr)
	}

	setResolutionMilestone := func(m *ent.IncidentMilestoneMutation) {
		m.SetIncidentID(inc.ID)
		m.SetUserID(userID)
		m.SetKind("resolution")
		m.SetTimestamp(inc.OpenedAt.Add(30 * time.Minute))
	}
	if _, milestoneErr := d.incidents.SetIncidentMilestone(ctx, uuid.Nil, setResolutionMilestone); milestoneErr != nil {
		return fmt.Errorf("create demo incident resolution milestone: %w", milestoneErr)
	}

	return nil
}

func (d *dataSeeder) createRetrospective(ctx context.Context, inc *ent.Incident) error {
	checkoutAlias, checkoutAliasErr := d.lookupAlias(ctx, "demo:component:checkout_service")
	if checkoutAliasErr != nil {
		return checkoutAliasErr
	}
	if checkoutAlias == nil || checkoutAlias.SubjectKind != ksa.SubjectKindEntity || checkoutAlias.EntityID == nil {
		return fmt.Errorf("demo component alias %q was not found", "demo:component:checkout_service")
	}
	searchAlias, searchAliasErr := d.lookupAlias(ctx, "demo:component:search_api")
	if searchAliasErr != nil {
		return searchAliasErr
	}
	if searchAlias == nil || searchAlias.SubjectKind != ksa.SubjectKindEntity || searchAlias.EntityID == nil {
		return fmt.Errorf("demo component alias %q was not found", "demo:component:search_api")
	}
	callsAlias, callsAliasErr := d.lookupAlias(ctx, "demo:relationship:checkout_service:calls:search_api")
	if callsAliasErr != nil {
		return callsAliasErr
	}
	if callsAlias == nil || callsAlias.SubjectKind != ksa.SubjectKindRelationship || callsAlias.RelationshipID == nil {
		return fmt.Errorf("demo relationship alias %q was not found", "demo:relationship:checkout_service:calls:search_api")
	}

	retro, createRetroErr := d.retrospectives.CreateForIncident(ctx, inc.ID)
	if createRetroErr != nil {
		return fmt.Errorf("create demo retrospective: %w", createRetroErr)
	}

	includeSubjectsParams := rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      retro.SystemAnalysisID,
		EntityIds:       []uuid.UUID{*checkoutAlias.EntityID, *searchAlias.EntityID},
		RelationshipIds: []uuid.UUID{*callsAlias.RelationshipID},
	}
	if includeErr := d.systemAnalysis.IncludeSystemAnalysisSubjects(ctx, includeSubjectsParams); includeErr != nil {
		return fmt.Errorf("include demo retrospective graph subjects: %w", includeErr)
	}

	listEntitiesParams := rez.ListSystemAnalysisEntitiesParams{
		Predicates: []predicate.SystemAnalysisEntity{
			sae.AnalysisID(retro.SystemAnalysisID),
		},
	}
	analysisEntities, listEntitiesErr := d.systemAnalysis.ListSystemAnalysisEntities(ctx, listEntitiesParams)
	if listEntitiesErr != nil {
		return fmt.Errorf("list demo retrospective graph entities: %w", listEntitiesErr)
	}
	if positionErr := d.positionEntity(ctx, analysisEntities.Data, *checkoutAlias.EntityID, 0, 0); positionErr != nil {
		return positionErr
	}
	if positionErr := d.positionEntity(ctx, analysisEntities.Data, *searchAlias.EntityID, 350, 0); positionErr != nil {
		return positionErr
	}

	setObservation := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(retro.SystemAnalysisID)
		m.SetKind("observation")
		m.SetTitle("Search enrichment requests timing out")
		m.SetBody("Checkout requests are waiting for optional search enrichment.")
		m.SetOccurredAt(inc.OpenedAt.Add(5 * time.Minute))
	}
	setObservationSubject := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetKnowledgeRelationshipID(*callsAlias.RelationshipID)
		m.SetRole("subject")
	}
	if _, observationErr := d.systemAnalysis.SetSystemAnalysisEntry(ctx, uuid.Nil, setObservation, setObservationSubject); observationErr != nil {
		return fmt.Errorf("create demo retrospective observation: %w", observationErr)
	}
	setAction := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(retro.SystemAnalysisID)
		m.SetKind("action")
		m.SetTitle("Disabled optional search enrichment")
		m.SetBody("Checkout requests recovered after disabling optional search enrichment.")
		m.SetOccurredAt(inc.OpenedAt.Add(25 * time.Minute))
		m.SetSequence(1)
	}
	setActionSubject := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetKnowledgeEntityID(*checkoutAlias.EntityID)
		m.SetRole("subject")
	}
	if _, actionErr := d.systemAnalysis.SetSystemAnalysisEntry(ctx, uuid.Nil, setAction, setActionSubject); actionErr != nil {
		return fmt.Errorf("create demo retrospective action: %w", actionErr)
	}

	return nil
}

func (d *dataSeeder) positionEntity(ctx context.Context, entities []*ent.SystemAnalysisEntity, entityID uuid.UUID, posX, posY float64) error {
	var analysisEntity *ent.SystemAnalysisEntity
	for _, candidate := range entities {
		if candidate.KnowledgeEntityID == entityID {
			analysisEntity = candidate
			break
		}
	}
	if analysisEntity == nil {
		return fmt.Errorf("demo retrospective graph entity %s was not included", entityID)
	}
	setPosition := func(m *ent.SystemAnalysisEntityMutation) {
		m.SetPosX(posX)
		m.SetPosY(posY)
	}
	if _, setEntityErr := d.systemAnalysis.SetSystemAnalysisEntity(ctx, analysisEntity.ID, setPosition); setEntityErr != nil {
		return fmt.Errorf("position demo retrospective graph entity: %w", setEntityErr)
	}
	return nil
}
