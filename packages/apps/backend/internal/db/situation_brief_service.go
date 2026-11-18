package db

import (
	"cmp"
	"context"
	"fmt"
	"github.com/rezible/rezible/pkg/situations"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sit "github.com/rezible/rezible/ent/situation"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
)

// briefChangesAfterOpenedStart is how long after a situation opened the lookup of later deployments starts.
// The changes lookup's start is inclusive, so a deployment at the opening is listed once, before it opened.
const briefChangesAfterOpenedStart = time.Millisecond

type SituationBriefService struct {
	db           rez.Database
	clock        rez.Clock
	changePoints rez.ChangePointService
	proposals    rez.KnowledgeProposalService
	integrations rez.InstalledIntegrationGetter
}

func NewSituationBriefService(
	db rez.Database,
	clock rez.Clock,
	changePoints rez.ChangePointService,
	proposals rez.KnowledgeProposalService,
	integrations rez.InstalledIntegrationGetter,
) (*SituationBriefService, error) {
	s := &SituationBriefService{
		db:           db,
		clock:        clock,
		changePoints: changePoints,
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
		WithObservations(func(q *ent.SituationObservationQuery) {
			q.WithScopes()
		})
	situation, queryErr := querySituation.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query situation: %w", queryErr)
	}

	var members []situations.RuntimeEntity
	for _, member := range situations.RuntimeEntities(situation.Edges.Observations) {
		entity := member.Entity
		if entity != nil && knowledgegraph.IsService(entity.Category, entity.Kind) {
			members = append(members, member)
		}
	}
	slices.SortFunc(members, func(a, b situations.RuntimeEntity) int {
		if a.Matching != b.Matching {
			if a.Matching {
				return -1
			}
			return 1
		}
		nameA, nameB := a.Entity.State.DisplayName, b.Entity.State.DisplayName
		return cmp.Or(strings.Compare(nameA, nameB), strings.Compare(a.Entity.ID.String(), b.Entity.ID.String()))
	})

	loaded := &briefSituation{situation: situation, total: len(members)}
	for _, member := range members[:min(len(members), rez.MaxBriefServices)] {
		loaded.services = append(loaded.services, member.Entity)
	}
	return loaded, nil
}

func (s *SituationBriefService) ListBriefServices(ctx context.Context, situationID uuid.UUID) ([]*ent.KnowledgeEntity, error) {
	loaded, loadErr := s.loadSituation(ctx, situationID)
	if loadErr != nil {
		return nil, loadErr
	}
	return loaded.services, nil
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

	brief.Changes = *s.listChanges(ctx, loaded, sources.Changes)
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
	if earliest := window.end.Add(-rez.MaxServiceChangesWindow); window.afterStart.Before(earliest) {
		window.afterStart = earliest
		window.afterStartMoved = true
	}
	return window
}

// hasAfter is whether the lookup after opening has a window at all.
func (w changesWindow) hasAfter() bool {
	return w.end.After(w.afterStart)
}

// listChanges looks up each service's changes. A service whose lookup fails says so, and the others are kept.
func (s *SituationBriefService) listChanges(ctx context.Context, loaded *briefSituation, deployments rezai.SourceAvailability) *rez.SituationBriefChanges {
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
		return changes
	}
	for _, service := range loaded.services {
		serviceChanges, listErr := s.listServiceChanges(ctx, service, window)
		if listErr != nil {
			slog.ErrorContext(ctx, "list service changes for situation brief", "service_entity_id", service.ID, "error", listErr)
			serviceChanges = &rez.SituationBriefServiceChanges{
				ServiceEntityID:   service.ID,
				UnavailableReason: "This service's deployments could not be loaded.",
			}
		}
		changes.Services = append(changes.Services, *serviceChanges)
	}
	return changes
}

// listServiceChanges looks up the service's deployments before and after the situation opened separately, so
// later deployments cannot push out the ones before it.
func (s *SituationBriefService) listServiceChanges(ctx context.Context, service *ent.KnowledgeEntity, window changesWindow) (*rez.SituationBriefServiceChanges, error) {
	before := rez.ListServiceChangesParams{
		ServiceEntityID: service.ID,
		Start:           window.start,
		End:             window.openedAt,
	}
	listedBefore, beforeErr := s.changePoints.ListServiceChanges(ctx, before)
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
		listedAfter, afterErr = s.changePoints.ListServiceChanges(ctx, after)
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
		params := rez.ListChangePointsParams{
			EntityID:      service.ID,
			Roles:         []string{projections.AssertionDeploymentService},
			Kind:          projections.KindDeployment,
			From:          window.start,
			FromInclusive: true,
			To:            window.openedAt,
			ToInclusive:   true,
			Limit:         1,
		}
		deployments, listErr := s.changePoints.ListChangePoints(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list deployments of service %s: %w", service.ID, listErr)
		}
		if len(deployments.ChangePoints) == 0 {
			continue
		}
		deployment := deployments.ChangePoints[0]
		if latest != nil && !deployment.ChangePoint.OccurredAt.After(latest.DeployedAt) {
			continue
		}
		decoded, decodeErr := projections.DecodeDeploymentEvent(deployment.ChangePoint.Edges.LatestEvent)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode deployment: %w", decodeErr)
		}
		attributes := decoded.Attributes
		latest = &rez.SituationLatestDeployment{
			Service:     service,
			DeployedAt:  deployment.ChangePoint.OccurredAt,
			Environment: attributes.Environment.Name,
			Status:      attributes.Status,
			Version:     attributes.Version,
			Sha:         attributes.Sha,
		}
	}
	return latest, nil
}
