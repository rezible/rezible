package demoprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertepisode"
	"github.com/rezible/rezible/ent/alertinstance"
	"github.com/rezible/rezible/ent/alertinstanceevent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentimpact"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/incidentrole"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/situationsignal"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	"github.com/rezible/rezible/pkg/execution"
)

// SignalProcessor places a recorded signal into situations, as the process-situation-signal job does.
// Processing a signal again is safe.
type SignalProcessor interface {
	ProcessSignal(ctx context.Context, signalEntityID uuid.UUID) error
}

func SeedDemoData(
	ctx context.Context,
	db rez.Database,
	knowledge rez.KnowledgeGraphQueryService,
	incidents rez.IncidentService,
	retrospectives rez.RetrospectiveService,
	systemAnalysis rez.SystemAnalysisService,
	events rez.EventsService,
	situations rez.SituationService,
	signals SignalProcessor,
) error {
	s := &dataSeeder{
		db:             db,
		knowledge:      knowledge,
		incidents:      incidents,
		retrospectives: retrospectives,
		systemAnalysis: systemAnalysis,
		events:         events,
		situations:     situations,
		signals:        signals,
	}
	return s.seedData(ctx)
}

type dataSeeder struct {
	db             rez.Database
	knowledge      rez.KnowledgeGraphQueryService
	incidents      rez.IncidentService
	retrospectives rez.RetrospectiveService
	systemAnalysis rez.SystemAnalysisService
	events         rez.EventsService
	situations     rez.SituationService
	signals        SignalProcessor
}

func (d *dataSeeder) seedData(ctx context.Context) error {
	inc, lookupErr := d.lookupIncident(ctx)
	if lookupErr != nil {
		return lookupErr
	}
	if inc == nil {
		return fmt.Errorf("demo source incident has not been imported")
	}

	if assignmentsErr := d.createRoleAssignments(ctx, inc); assignmentsErr != nil {
		return assignmentsErr
	}
	if milestonesErr := d.createMilestones(ctx, inc); milestonesErr != nil {
		return milestonesErr
	}
	if impactsErr := d.createImpacts(ctx, inc); impactsErr != nil {
		return impactsErr
	}
	if situationErr := d.linkSituation(ctx, inc); situationErr != nil {
		return situationErr
	}

	if retrospectiveErr := d.createRetrospective(ctx, inc); retrospectiveErr != nil {
		return retrospectiveErr
	}

	return nil
}

func (d *dataSeeder) lookupAlias(ctx context.Context, resourceRef string) (*ent.KnowledgeSubjectAlias, error) {
	params := rez.ListKnowledgeSubjectAliasesParams{
		Predicates: []predicate.KnowledgeSubjectAlias{
			ksa.Provider(ProviderName),
			ksa.ProviderNamespace(ProviderName),
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

func (d *dataSeeder) ensureIncidentRole(ctx context.Context, name string) (*ent.IncidentRole, error) {
	roleQuery := d.db.Client(ctx).IncidentRole.Query().Where(incidentrole.Name(name))
	role, queryErr := roleQuery.Only(ctx)
	if queryErr == nil {
		return role, nil
	}
	if !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("lookup demo incident role %q: %w", name, queryErr)
	}
	created, createErr := d.db.Client(ctx).IncidentRole.Create().SetName(name).Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create demo incident role %q: %w", name, createErr)
	}
	return created, nil
}

func (d *dataSeeder) createRoleAssignments(ctx context.Context, inc *ent.Incident) error {
	userID, userOK := execution.GetContext(ctx).UserID()
	if !userOK || userID == uuid.Nil {
		return fmt.Errorf("demo incident role assignments require a user context")
	}
	commander, commanderErr := d.ensureIncidentRole(ctx, "Incident Commander")
	if commanderErr != nil {
		return commanderErr
	}
	userParams := rez.SetIncidentRoleAssignmentParams{IncidentID: inc.ID, RoleID: commander.ID, UserID: userID}
	if _, assignErr := d.incidents.SetIncidentRoleAssignment(ctx, uuid.Nil, userParams); assignErr != nil {
		return fmt.Errorf("assign demo incident commander: %w", assignErr)
	}
	return nil
}

func (d *dataSeeder) createMilestones(ctx context.Context, inc *ent.Incident) error {
	type milestoneItem struct {
		kind        incidentmilestone.Kind
		timestamp   time.Time
		description string
	}
	items := []milestoneItem{
		{kind: incidentmilestone.KindDetection, timestamp: inc.OpenedAt, description: "The first customer reports identified elevated checkout timeouts."},
		{kind: incidentmilestone.KindInvestigation, timestamp: inc.OpenedAt.Add(8 * time.Minute), description: "The team isolated failures to optional search enrichment."},
		{kind: incidentmilestone.KindMitigation, timestamp: inc.OpenedAt.Add(25 * time.Minute), description: "Optional search enrichment was disabled for checkout requests."},
	}
	if resolutionAt := inc.ResolvedAt; resolutionAt != nil {
		items = append(items, milestoneItem{kind: incidentmilestone.KindResolution, timestamp: *resolutionAt, description: "Checkout requests returned to normal after mitigation."})
	}
	for _, item := range items {
		present := false
		for _, existing := range inc.Edges.Milestones {
			if existing.Kind == item.kind && existing.Source == "demo" {
				present = true
				break
			}
		}
		if present {
			continue
		}
		setMilestone := func(m *ent.IncidentMilestoneMutation) {
			m.SetIncidentID(inc.ID)
			m.SetKind(item.kind)
			m.SetSource("demo")
			m.SetDescription(item.description)
			m.SetTimestamp(item.timestamp)
		}
		if _, milestoneErr := d.incidents.SetIncidentMilestone(ctx, uuid.Nil, setMilestone); milestoneErr != nil {
			return fmt.Errorf("create demo incident %s milestone: %w", item.kind, milestoneErr)
		}
	}
	return nil
}

func (d *dataSeeder) createImpacts(ctx context.Context, inc *ent.Incident) error {
	impacts := []struct {
		resourceRef string
		note        string
	}{
		{resourceRef: "demo:component:checkout_service", note: "Checkout requests that use search enrichment timed out for some customers."},
		{resourceRef: "demo:component:search_api", note: "Search API latency delayed checkout enrichment requests."},
	}
	for _, item := range impacts {
		alias, aliasErr := d.lookupAlias(ctx, item.resourceRef)
		if aliasErr != nil {
			return aliasErr
		}
		if alias == nil || alias.SubjectKind != ksa.SubjectKindEntity || alias.EntityID == nil {
			return fmt.Errorf("demo impact entity %q was not found", item.resourceRef)
		}
		existsQuery := d.db.Client(ctx).IncidentImpact.Query().
			Where(incidentimpact.IncidentID(inc.ID), incidentimpact.KnowledgeEntityID(*alias.EntityID))
		exists, queryErr := existsQuery.Exist(ctx)
		if queryErr != nil {
			return fmt.Errorf("check demo incident impact: %w", queryErr)
		}
		if exists {
			continue
		}
		createImpact := d.db.Client(ctx).IncidentImpact.Create().
			SetIncidentID(inc.ID).
			SetKnowledgeEntityID(*alias.EntityID).
			SetSource("demo").
			SetNote(item.note)
		if createErr := createImpact.Exec(ctx); createErr != nil {
			return fmt.Errorf("create demo incident impact: %w", createErr)
		}
	}
	return nil
}

// linkSituation places the demo incident's alerts into situations, as signal processing does, and links the
// incident to the situation holding its search latency alert. Background signal processing may already have
// placed them; processing again changes nothing.
func (d *dataSeeder) linkSituation(ctx context.Context, inc *ent.Incident) error {
	var latencyEntityID uuid.UUID
	for _, firing := range []alertObservedPayload{demoSearchLatencyFiring, demoElasticsearchCPUFiring} {
		entityID, episodeErr := d.alertEpisodeEntity(ctx, firing.eventRef())
		if episodeErr != nil {
			return episodeErr
		}
		if processErr := d.signals.ProcessSignal(ctx, entityID); processErr != nil {
			return fmt.Errorf("place demo alert %s: %w", firing.DefinitionRef, processErr)
		}
		if firing.DefinitionRef == demoSearchLatencyFiring.DefinitionRef {
			latencyEntityID = entityID
		}
	}

	queryMembership := d.db.Client(ctx).SituationSignal.Query().
		Where(situationsignal.KnowledgeEntityID(latencyEntityID))
	membership, membershipErr := queryMembership.Only(ctx)
	if membershipErr != nil {
		return fmt.Errorf("find demo incident situation: %w", membershipErr)
	}
	links := rez.IncidentSituationLinkChanges{Added: []uuid.UUID{membership.SituationID}}
	if linkErr := d.situations.SyncIncidentLinks(ctx, inc.ID, links); linkErr != nil {
		return fmt.Errorf("link demo incident to situation: %w", linkErr)
	}
	return nil
}

// alertEpisodeEntity returns the knowledge entity of the alert episode holding the notification.
func (d *dataSeeder) alertEpisodeEntity(ctx context.Context, eventRef string) (uuid.UUID, error) {
	eventParams := rez.ListEventsParams{
		PageSize:   1,
		Predicates: []predicate.NormalizedEvent{normalizedevent.ProviderEventRef(eventRef)},
	}
	alertEvents, listErr := d.events.ListEvents(ctx, eventParams)
	if listErr != nil {
		return uuid.Nil, fmt.Errorf("find demo alert %s: %w", eventRef, listErr)
	}
	if len(alertEvents.Data) == 0 {
		return uuid.Nil, fmt.Errorf("demo alert %s was not found", eventRef)
	}
	queryEpisode := d.db.Client(ctx).AlertEpisode.Query().
		Where(alertepisode.HasInstancesWith(alertinstance.HasEventsWith(alertinstanceevent.EventID(alertEvents.Data[0].ID))))
	episode, episodeErr := queryEpisode.Only(ctx)
	if episodeErr != nil {
		return uuid.Nil, fmt.Errorf("find demo alert episode %s: %w", eventRef, episodeErr)
	}
	return *episode.KnowledgeEntityID, nil
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

	//setObservation := func(m *ent.SystemAnalysisEntryMutation) {
	//	m.SetAnalysisID(retro.SystemAnalysisID)
	//	m.SetKind("observation")
	//	m.SetTitle("Search enrichment requests timing out")
	//	m.SetBody("Checkout requests are waiting for optional search enrichment.")
	//	m.SetOccurredAt(inc.OpenedAt.Add(5 * time.Minute))
	//}
	//setObservationSubject := func(m *ent.SystemAnalysisEntrySubjectMutation) {
	//	m.SetKnowledgeRelationshipID(*callsAlias.RelationshipID)
	//	m.SetRole("context")
	//}
	//if _, observationErr := d.systemAnalysis.SetSystemAnalysisEntry(ctx, uuid.Nil, setObservation, setObservationSubject); observationErr != nil {
	//	return fmt.Errorf("create demo retrospective observation: %w", observationErr)
	//}
	//setAction := func(m *ent.SystemAnalysisEntryMutation) {
	//	m.SetAnalysisID(retro.SystemAnalysisID)
	//	m.SetKind("action")
	//	m.SetTitle("Disabled optional search enrichment")
	//	m.SetBody("Checkout requests recovered after disabling optional search enrichment.")
	//	m.SetOccurredAt(inc.OpenedAt.Add(25 * time.Minute))
	//	m.SetSequence(1)
	//}
	//setActionSubject := func(m *ent.SystemAnalysisEntrySubjectMutation) {
	//	m.SetKnowledgeEntityID(*checkoutAlias.EntityID)
	//	m.SetRole("context")
	//}
	//if _, actionErr := d.systemAnalysis.SetSystemAnalysisEntry(ctx, uuid.Nil, setAction, setActionSubject); actionErr != nil {
	//	return fmt.Errorf("create demo retrospective action: %w", actionErr)
	//}

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
