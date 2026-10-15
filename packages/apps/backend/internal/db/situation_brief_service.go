package db

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	sit "github.com/rezible/rezible/ent/situation"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
)

// briefChangesAfterOpenedStart is how long after a situation opened the lookup of later deployments starts.
// The changes lookup's start is inclusive, so a deployment at the opening is listed once, before it opened.
const briefChangesAfterOpenedStart = time.Millisecond

type SituationBriefService struct {
	db           rez.Database
	clock        rez.Clock
	graph        rez.KnowledgeGraphQueryService
	changes      rez.ChangesService
	proposals    rez.KnowledgeProposalService
	integrations rez.InstalledIntegrationGetter
}

func NewSituationBriefService(
	db rez.Database,
	clock rez.Clock,
	graph rez.KnowledgeGraphQueryService,
	changes rez.ChangesService,
	proposals rez.KnowledgeProposalService,
	integrations rez.InstalledIntegrationGetter,
) (*SituationBriefService, error) {
	s := &SituationBriefService{
		db:           db,
		clock:        clock,
		graph:        graph,
		changes:      changes,
		proposals:    proposals,
		integrations: integrations,
	}
	return s, nil
}

// briefSituation is a situation with its services, as the brief shows them.
type briefSituation struct {
	situation *ent.Situation
	// services are at most rez.MaxBriefServices; total counts all of them.
	services []*ent.KnowledgeEntity
	total    int
}

func (s *SituationBriefService) loadSituation(ctx context.Context, situationID uuid.UUID) (*briefSituation, error) {
	querySituation := s.db.Client(ctx).Situation.Query().
		Where(sit.ID(situationID)).
		WithEntities(func(q *ent.SituationEntityQuery) {
			q.WithKnowledgeEntity()
		})
	situation, queryErr := querySituation.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query situation: %w", queryErr)
	}

	var members []*ent.SituationEntity
	for _, member := range situation.Edges.Entities {
		entity := member.Edges.KnowledgeEntity
		if entity != nil && entity.Category == kne.CategoryContainer && entity.Kind == projections.EntityKindService {
			members = append(members, member)
		}
	}
	slices.SortFunc(members, func(a, b *ent.SituationEntity) int {
		if a.Matching != b.Matching {
			if a.Matching {
				return -1
			}
			return 1
		}
		nameA, nameB := a.Edges.KnowledgeEntity.State.DisplayName, b.Edges.KnowledgeEntity.State.DisplayName
		return cmp.Or(strings.Compare(nameA, nameB), strings.Compare(a.KnowledgeEntityID.String(), b.KnowledgeEntityID.String()))
	})

	loaded := &briefSituation{situation: situation, total: len(members)}
	for _, member := range members[:min(len(members), rez.MaxBriefServices)] {
		loaded.services = append(loaded.services, member.Edges.KnowledgeEntity)
	}
	return loaded, nil
}

func (s *SituationBriefService) GetSituationBrief(ctx context.Context, situationID uuid.UUID) (*rez.SituationBrief, error) {
	loaded, loadErr := s.loadSituation(ctx, situationID)
	if loadErr != nil {
		return nil, loadErr
	}
	sources := integrations.CheckSources(ctx, s.integrations)

	serviceIDs := make([]uuid.UUID, len(loaded.services))
	for index, service := range loaded.services {
		serviceIDs[index] = service.ID
	}
	matches, matchesErr := s.proposals.ListUnconfirmedMatches(ctx, serviceIDs)
	if matchesErr != nil {
		return nil, fmt.Errorf("list unconfirmed matches: %w", matchesErr)
	}

	brief := &rez.SituationBrief{
		Services: make([]rez.BriefService, len(loaded.services)),
		Unseen:   s.unseen(sources, loaded),
	}
	for index, service := range loaded.services {
		brief.Services[index] = rez.BriefService{Entity: service, UnconfirmedMatches: matches[service.ID]}
	}

	changes, changesErr := s.listChanges(ctx, loaded, sources.Changes)
	if changesErr != nil {
		return nil, changesErr
	}
	brief.Changes = *changes
	return brief, nil
}

// changesWindow is the brief's two changes lookups: before the situation opened, and after it.
type changesWindow struct {
	start, openedAt, afterStart, end time.Time
	// afterStartMoved is set when the later lookup's start was moved to stay within the changes lookup's limit.
	afterStartMoved bool
}

func (s *SituationBriefService) changesWindow(situation *ent.Situation) changesWindow {
	window := changesWindow{
		start:      situation.OpenedAt.Add(-rez.BriefChangesLeadIn),
		openedAt:   situation.OpenedAt,
		afterStart: situation.OpenedAt.Add(briefChangesAfterOpenedStart),
		end:        s.clock.Now(),
	}
	if situation.ClosedAt != nil {
		window.end = *situation.ClosedAt
	}
	if earliest := window.end.Add(-changesMaxWindow); window.afterStart.Before(earliest) {
		window.afterStart = earliest
		window.afterStartMoved = true
	}
	return window
}

// hasAfter is whether the lookup after opening has a window at all.
func (w changesWindow) hasAfter() bool {
	return w.end.After(w.afterStart)
}

func (s *SituationBriefService) listChanges(ctx context.Context, loaded *briefSituation, deployments rezai.SourceAvailability) (*rez.SituationBriefChanges, error) {
	window := s.changesWindow(loaded.situation)
	changes := &rez.SituationBriefChanges{
		Available:         deployments.Available,
		UnavailableReason: deployments.Reason,
		Start:             window.start,
		OpenedAt:          window.openedAt,
		End:               window.end,
		Services:          make([]rez.SituationBriefServiceChanges, 0, len(loaded.services)),
	}
	if !deployments.Available {
		return changes, nil
	}
	for _, service := range loaded.services {
		serviceChanges, listErr := s.listServiceChanges(ctx, service, window)
		if listErr != nil {
			return nil, fmt.Errorf("service %s: %w", service.ID, listErr)
		}
		changes.Services = append(changes.Services, *serviceChanges)
	}
	return changes, nil
}

// listServiceChanges looks up the service's deployments before and after the situation opened separately, so
// later deployments cannot push out the ones before it.
func (s *SituationBriefService) listServiceChanges(ctx context.Context, service *ent.KnowledgeEntity, window changesWindow) (*rez.SituationBriefServiceChanges, error) {
	before := rez.ListServiceChangesParams{
		ServiceEntityID: service.ID,
		Start:           window.start,
		End:             window.openedAt,
	}
	listedBefore, beforeErr := s.changes.ListServiceChanges(ctx, before)
	if beforeErr != nil {
		return nil, fmt.Errorf("list changes before opening: %w", beforeErr)
	}

	listedAfter := &rez.ServiceChanges{}
	if window.hasAfter() {
		after := rez.ListServiceChangesParams{
			ServiceEntityID: service.ID,
			Start:           window.afterStart,
			End:             window.end,
		}
		var afterErr error
		listedAfter, afterErr = s.changes.ListServiceChanges(ctx, after)
		if afterErr != nil {
			return nil, fmt.Errorf("list changes after opening: %w", afterErr)
		}
	}

	serviceChanges := &rez.SituationBriefServiceChanges{
		ServiceEntityID: service.ID,
		Deployments:     make([]rez.SituationBriefDeployment, 0, len(listedAfter.Deployments)+len(listedBefore.Deployments)),
		Limits:          slices.Concat(listedAfter.Limits, listedBefore.Limits),
	}
	if window.afterStartMoved && window.hasAfter() {
		limit := fmt.Sprintf("deployments after the situation opened are listed from %s, 7 days before its end", window.afterStart.UTC().Format(time.RFC3339))
		serviceChanges.Limits = append(serviceChanges.Limits, limit)
	}
	for _, deployment := range listedAfter.Deployments {
		serviceChanges.Deployments = append(serviceChanges.Deployments, s.briefDeployment(deployment, true))
	}
	for _, deployment := range listedBefore.Deployments {
		serviceChanges.Deployments = append(serviceChanges.Deployments, s.briefDeployment(deployment, false))
	}

	if evidenceErr := s.setLatestEvidence(ctx, serviceChanges.Deployments); evidenceErr != nil {
		return nil, evidenceErr
	}
	return serviceChanges, nil
}

func (s *SituationBriefService) briefDeployment(deployment rez.ServiceDeployment, afterOpened bool) rez.SituationBriefDeployment {
	briefed := rez.SituationBriefDeployment{ServiceDeployment: deployment, AfterOpened: afterOpened}
	if deployment.RepositoryEntityID == nil {
		briefed.MergedChangesUnlinked = rez.SituationBriefUnlinkedNoRepository
	} else if !deployment.RepositorySeenByCodeForge {
		briefed.MergedChangesUnlinked = rez.SituationBriefUnlinkedRepositoryNotForge
	}
	return briefed
}

// setLatestEvidence sets each deployment's latest evidence, for a deployment with no link to open.
func (s *SituationBriefService) setLatestEvidence(ctx context.Context, deployments []rez.SituationBriefDeployment) error {
	if len(deployments) == 0 {
		return nil
	}
	entityIDs := make([]uuid.UUID, len(deployments))
	for index, deployment := range deployments {
		entityIDs[index] = deployment.Deployment.ID
	}
	queryEvidence := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.HasSubjectAliasWith(ksa.EntityIDIn(entityIDs...))).
		WithSubjectAlias().
		Order(ke.ByEffectiveAt(sql.OrderDesc()), ke.ByID(sql.OrderDesc()))
	evidence, queryErr := queryEvidence.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query deployment evidence: %w", queryErr)
	}
	latest := make(map[uuid.UUID]uuid.UUID, len(deployments))
	for _, item := range evidence {
		entityID := item.Edges.SubjectAlias.EntityID
		if entityID == nil {
			continue
		}
		if _, seen := latest[*entityID]; !seen {
			latest[*entityID] = item.ID
		}
	}
	for index := range deployments {
		if evidenceID, found := latest[deployments[index].Deployment.ID]; found {
			deployments[index].LatestEvidenceID = &evidenceID
		}
	}
	return nil
}

func (s *SituationBriefService) unseen(sources integrations.Sources, loaded *briefSituation) []rez.SituationBriefUnseen {
	var unseen []rez.SituationBriefUnseen
	if !sources.Logs.Available {
		unseen = append(unseen, rez.SituationBriefUnseen{
			Kind: rez.SituationBriefUnseenSourceUnavailable,
			Text: fmt.Sprintf("Logs cannot be read: %s.", sources.Logs.Reason),
		})
	}
	if !sources.Changes.Available {
		unseen = append(unseen, rez.SituationBriefUnseen{
			Kind: rez.SituationBriefUnseenSourceUnavailable,
			Text: fmt.Sprintf("Deployments cannot be read: %s.", sources.Changes.Reason),
		})
	}
	if !sources.CodeChanges.Available {
		unseen = append(unseen, rez.SituationBriefUnseen{
			Kind: rez.SituationBriefUnseenSourceUnavailable,
			Text: "No code forge is connected, so merged changes are not linked.",
		})
	}
	if loaded.total > len(loaded.services) {
		unseen = append(unseen, rez.SituationBriefUnseen{
			Kind: rez.SituationBriefUnseenServicesCut,
			Text: fmt.Sprintf("This situation involves %d services; the brief covers the first %d.", loaded.total, len(loaded.services)),
		})
	}
	unseen = append(unseen, rez.SituationBriefUnseen{
		Kind: rez.SituationBriefUnseenNotRead,
		Text: "Rezible does not read metrics or traces yet.",
	})
	return unseen
}

func (s *SituationBriefService) LatestDeploymentBefore(ctx context.Context, situationID uuid.UUID) (*rez.SituationLatestDeployment, error) {
	loaded, loadErr := s.loadSituation(ctx, situationID)
	if loadErr != nil {
		return nil, loadErr
	}
	window := s.changesWindow(loaded.situation)

	var latest *rez.SituationLatestDeployment
	for _, service := range loaded.services {
		params := rez.ListRelatedEventsParams{
			EntityID:      service.ID,
			Predicate:     knr.PredicateImpacts,
			Kind:          changesDeploymentEventKind,
			From:          window.start,
			FromInclusive: true,
			To:            window.openedAt,
			ToInclusive:   true,
			Limit:         1,
		}
		deployments, listErr := s.graph.ListRelatedEvents(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list deployments of service %s: %w", service.ID, listErr)
		}
		if len(deployments.Events) == 0 {
			continue
		}
		deployment := deployments.Events[0]
		if latest != nil && !deployment.StateEffectiveAt.After(latest.DeployedAt) {
			continue
		}
		properties := changesEventProperties(deployment.State.Properties)
		latest = &rez.SituationLatestDeployment{
			Service:     service,
			DeployedAt:  *deployment.StateEffectiveAt,
			Environment: properties.get(projections.DeploymentPropertyEnvironment),
			Status:      properties.get(projections.DeploymentPropertyStatus),
			Version:     properties.get(projections.DeploymentPropertyVersion),
			Sha:         properties.get(projections.DeploymentPropertySha),
		}
	}
	return latest, nil
}
