package db

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	sit "github.com/rezible/rezible/ent/situation"
	sitent "github.com/rezible/rezible/ent/situationentity"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/situations"
)

// situationCorrelationLock names the tenant correlation lock.
const situationCorrelationLock = "situation_correlation"

func NewProcessSituationSignalWorker(service *SituationService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(func(ctx context.Context, args jobs.ProcessSituationSignal) error {
		return service.ProcessSignal(ctx, args.SignalEntityID)
	})
}

// ProcessSignal brings situations up to date with a signal's stored state. It refreshes the situation
// holding the signal, or places an eligible signal that no situation holds. Unknown signals and kinds
// without a source are ignored.
func (s *SituationService) ProcessSignal(ctx context.Context, signalEntityID uuid.UUID) error {
	signals, loadErr := s.loadSignals(ctx, []uuid.UUID{signalEntityID}, situations.LoadSignalsOptions{})
	if loadErr != nil {
		if errors.Is(loadErr, rez.ErrInvalidInput) {
			return nil
		}
		return loadErr
	}
	signal := signals[signalEntityID]
	held, heldErr := s.refreshHoldingSituation(ctx, signalEntityID)
	if heldErr != nil || held {
		return heldErr
	}
	if !signal.EligibleAt(s.clock.Now()) {
		return nil
	}
	return s.correlateSignal(ctx, signal)
}

// refreshHoldingSituation recomputes the entity rows of the situation holding the signal and requests its
// evaluation, unless it has closed. It reports whether any situation holds the signal: a signal is never
// placed twice.
func (s *SituationService) refreshHoldingSituation(ctx context.Context, signalEntityID uuid.UUID) (bool, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (bool, error) {
		queryMembership := tx.SituationSignal.Query().
			Where(sitsig.KnowledgeEntityID(signalEntityID))
		membership, queryErr := queryMembership.Only(ctx)
		if queryErr != nil {
			if ent.IsNotFound(queryErr) {
				return false, nil
			}
			return false, fmt.Errorf("query signal membership: %w", queryErr)
		}
		locked, lockErr := s.lockSituations(ctx, membership.SituationID)
		if lockErr != nil {
			return false, lockErr
		}
		if locked[membership.SituationID].ClosedAt != nil {
			return true, nil
		}
		if entitiesErr := s.recomputeSituationEntities(ctx, membership.SituationID); entitiesErr != nil {
			return true, entitiesErr
		}
		return true, s.requestEvaluation(ctx, membership.SituationID)
	})
}

// correlateSignal places the signal into a situation about the same or a related runtime entity, or into
// a new candidate.
func (s *SituationService) correlateSignal(ctx context.Context, signal situations.Signal) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// Without this lock, two signals about the same entities correlated at once could each find no
		// match and each start a candidate. No row represents the tenant's situations about some entities,
		// so no row lock can prevent it. AcquireTxLocks scopes the key to the tenant, making it one lock per
		// tenant. It is taken before any situation row lock.
		if lockErr := s.db.AcquireTxLocks(ctx, situationCorrelationLock, "tenant"); lockErr != nil {
			return fmt.Errorf("acquire situation correlation lock: %w", lockErr)
		}
		// Two runs for one signal can both find it unplaced before either takes the lock.
		held, heldErr := s.refreshHoldingSituation(ctx, signal.EntityID)
		if heldErr != nil || held {
			return heldErr
		}

		entityIDs, entitiesErr := s.signalRuntimeEntities(ctx, signal)
		if entitiesErr != nil {
			return entitiesErr
		}
		relationships, relationshipsErr := s.correlationRelationships(ctx, entityIDs)
		if relationshipsErr != nil {
			return relationshipsErr
		}
		reach := entityIDs.Clone()
		for _, relationship := range relationships {
			reach.Append(relationship.SourceEntityID, relationship.TargetEntityID)
		}
		candidates, candidatesErr := s.correlationSituations(ctx, reach)
		if candidatesErr != nil {
			return candidatesErr
		}
		input := situations.CorrelationInput{
			Now:           s.clock.Now(),
			Signal:        signal,
			EntityIDs:     entityIDs,
			Relationships: relationships,
			Situations:    candidates,
		}
		decision := situations.DecideCorrelation(input)
		switch {
		case !decision.Placed():
			return nil
		case decision.MatchKind == sitsig.MatchKindSeed:
			return s.startCandidate(ctx, signal, input.Now)
		}
		attachParams := AttachSituationSignalsParams{
			SituationID:       decision.SituationID,
			GroupTitle:        signal.Title,
			SignalEntityIDs:   []uuid.UUID{signal.EntityID},
			MatchKind:         decision.MatchKind,
			ViaRelationshipID: decision.ViaRelationshipID,
		}
		return s.AttachSituationSignals(ctx, attachParams)
	})
}

// correlationRelationships lists the depends-on and adjacent relationships touching the entities.
func (s *SituationService) correlationRelationships(ctx context.Context, entityIDs mapset.Set[uuid.UUID]) ([]situations.CorrelationRelationship, error) {
	if entityIDs.IsEmpty() {
		return nil, nil
	}
	ids := entityIDs.ToSlice()
	classPredicates := slices.Concat(
		knowledgegraph.RelationshipClassDependsOn.Predicates(),
		knowledgegraph.RelationshipClassAdjacent.Predicates(),
	)
	listParams := rez.ListKnowledgeRelationshipsParams{
		PageSize: signalGraphPageSize,
		Predicates: []predicate.KnowledgeRelationship{
			knr.PredicateIn(classPredicates...),
			knr.Or(knr.SourceEntityIDIn(ids...), knr.TargetEntityIDIn(ids...)),
		},
	}
	listed, listErr := s.graph.ListRelationships(ctx, listParams)
	if listErr != nil {
		return nil, fmt.Errorf("list correlation relationships: %w", listErr)
	}
	relationships := make([]situations.CorrelationRelationship, len(listed.Data))
	for i, relationship := range listed.Data {
		relationships[i] = situations.CorrelationRelationship{
			ID:             relationship.ID,
			Predicate:      relationship.Predicate,
			SourceEntityID: relationship.SourceEntityID,
			TargetEntityID: relationship.TargetEntityID,
		}
	}
	return relationships, nil
}

// correlationSituations locks the unclosed situations with a matching entity in the reach and describes
// them from their signals. A situation that closed before its lock was taken is left out.
func (s *SituationService) correlationSituations(ctx context.Context, reach mapset.Set[uuid.UUID]) ([]situations.CorrelationSituation, error) {
	if reach.IsEmpty() {
		return nil, nil
	}
	inReach := []predicate.SituationEntity{
		sitent.Matching(true),
		sitent.KnowledgeEntityIDIn(reach.ToSlice()...),
	}
	client := s.db.Client(ctx)
	querySituations := client.Situation.Query().
		Where(sit.ClosedAtIsNil(), sit.HasEntitiesWith(inReach...))
	ids, queryErr := querySituations.IDs(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query correlation situations: %w", queryErr)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	locked, lockErr := s.lockSituations(ctx, ids...)
	if lockErr != nil {
		return nil, lockErr
	}
	byID := make(map[uuid.UUID]*situations.CorrelationSituation, len(locked))
	for id, situation := range locked {
		if situation.ClosedAt == nil {
			byID[id] = &situations.CorrelationSituation{
				ID:                id,
				Raised:            situation.RaisedAt != nil,
				Muted:             situation.MutedAt != nil,
				MatchingEntityIDs: mapset.NewSet[uuid.UUID](),
			}
		}
	}

	queryEntities := client.SituationEntity.Query().
		Where(sitent.SituationIDIn(ids...)).
		Where(inReach...)
	entities, entitiesErr := queryEntities.All(ctx)
	if entitiesErr != nil {
		return nil, fmt.Errorf("query correlation situation entities: %w", entitiesErr)
	}
	for _, entity := range entities {
		if candidate := byID[entity.SituationID]; candidate != nil {
			candidate.MatchingEntityIDs.Add(entity.KnowledgeEntityID)
		}
	}

	queryMembers := client.SituationSignal.Query().
		Where(sitsig.SituationIDIn(ids...))
	members, membersErr := queryMembers.All(ctx)
	if membersErr != nil {
		return nil, fmt.Errorf("query correlation situation signals: %w", membersErr)
	}
	memberIDs := make([]uuid.UUID, len(members))
	for i, member := range members {
		memberIDs[i] = member.KnowledgeEntityID
	}

	signals, loadErr := s.loadSignals(ctx, memberIDs, situations.LoadSignalsOptions{})
	if loadErr != nil {
		return nil, loadErr
	}

	for _, member := range members {
		candidate := byID[member.SituationID]
		if candidate == nil {
			continue
		}
		signal := signals[member.KnowledgeEntityID]
		candidate.Active = candidate.Active || signal.Active
		for _, at := range []time.Time{member.AttachedAt, signal.LastActivityAt} {
			if at.After(candidate.LastActivityAt) {
				candidate.LastActivityAt = at
			}
		}
	}

	candidates := make([]situations.CorrelationSituation, 0, len(byID))
	for _, candidate := range byID {
		candidates = append(candidates, *candidate)
	}
	return candidates, nil
}

// startCandidate creates the signal's own candidate. It links the candidate as a recurrence of the most
// recent situation, other than a merged one, that closed within RecurrenceLookback and held a signal of
// the same source; equal closing times go to the lower ID.
func (s *SituationService) startCandidate(ctx context.Context, signal situations.Signal, now time.Time) error {
	createParams := rez.CreateSituationParams{
		Title:        signal.Title,
		SeedEntityID: signal.EntityID,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:           signal.Title,
			SignalEntityIDs: []uuid.UUID{signal.EntityID},
		}},
	}
	created, createErr := s.CreateSituation(ctx, createParams)
	if createErr != nil {
		return fmt.Errorf("start candidate: %w", createErr)
	}
	if signal.SourceEntityID == nil {
		return nil
	}
	client := s.db.Client(ctx)
	queryRecurrence := client.Situation.Query().
		Where(
			sit.ClosedAtGTE(now.Add(-situations.RecurrenceLookback)),
			sit.CloseReasonNEQ(sit.CloseReasonMerged),
			sit.HasSignalsWith(sitsig.SourceEntityID(*signal.SourceEntityID)),
		).
		Order(sit.ByClosedAt(sql.OrderDesc()), sit.ByID())
	earlier, queryErr := queryRecurrence.First(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil
		}
		return fmt.Errorf("query recurrence situation: %w", queryErr)
	}
	createLink := client.SituationLink.Create().
		SetSituationID(created.ID).
		SetLinkedSituationID(earlier.ID).
		SetKind(sitlink.KindRecurrenceOf).
		SetCreatedAt(now)
	if linkErr := createLink.Exec(ctx); linkErr != nil {
		return fmt.Errorf("link recurrence situation: %w", linkErr)
	}
	return nil
}
