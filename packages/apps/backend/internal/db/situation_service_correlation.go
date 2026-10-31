package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	sit "github.com/rezible/rezible/ent/situation"
	sitent "github.com/rezible/rezible/ent/situationentity"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/situations"
)

// situationCorrelationLock names the tenant correlation lock.
const situationCorrelationLock = "situation_correlation"

func NewProcessSituationObservationWorker(service *SituationService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(func(ctx context.Context, args jobs.ProcessSituationObservation) error {
		return service.observationJobError(service.ProcessObservation(ctx, args.Ref))
	})
}

// observationJobError marks processErr retryable when an entity merge deleted a knowledge subject that
// correlation read: correlation found it gone, or a write referring to it failed on its foreign key. The next
// attempt reads the merged graph.
func (s *SituationService) observationJobError(processErr error) error {
	if errors.Is(processErr, errs.ErrKnowledgeSubjectMerged) || s.db.IsForeignKeyError(processErr, ent.KnowledgeForeignKeyConstraints...) {
		return jobs.MarkRetryableError(processErr)
	}
	return processErr
}

// ProcessObservation brings situations up to date with an observation's stored state. It refreshes the
// situations that have the observation, or places an eligible observation that no situation has. Unknown
// observations are ignored.
func (s *SituationService) ProcessObservation(ctx context.Context, ref rez.SituationObservationRef) error {
	observations, loadErr := s.LoadObservations(ctx, []rez.SituationObservationRef{ref}, rez.LoadSituationObservationsOptions{})
	if loadErr != nil {
		if errors.Is(loadErr, errs.ErrInvalidInput) {
			return nil
		}
		return loadErr
	}
	observation := observations[ref]
	observed, observedErr := s.refreshObservingSituations(ctx, ref)
	if observedErr != nil || observed {
		return observedErr
	}
	if !situations.EligibleAt(observation, s.clock.Now()) {
		return nil
	}
	return s.correlateObservation(ctx, observation)
}

// refreshObservingSituations recomputes the entity rows of the unclosed situations with the observation and
// requests their evaluation. It reports whether any situation, closed or not, has the observation:
// correlation never places an observation again.
func (s *SituationService) refreshObservingSituations(ctx context.Context, ref rez.SituationObservationRef) (bool, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (bool, error) {
		queryObserving := tx.Situation.Query().
			Where(sit.HasObservationsWith(s.observesAny(ref)))
		ids, queryErr := queryObserving.IDs(ctx)
		if queryErr != nil {
			return false, fmt.Errorf("query observing situations: %w", queryErr)
		}
		if len(ids) == 0 {
			return false, nil
		}
		locked, lockErr := s.lockSituations(ctx, ids...)
		if lockErr != nil {
			return false, lockErr
		}
		for _, id := range ids {
			if locked[id].ClosedAt != nil {
				continue
			}
			if entitiesErr := s.recomputeSituationEntities(ctx, id); entitiesErr != nil {
				return true, entitiesErr
			}
			if evaluateErr := s.requestEvaluation(ctx, id); evaluateErr != nil {
				return true, evaluateErr
			}
		}
		return true, nil
	})
}

// correlateObservation places the observation into a situation about the same or a related runtime entity, or
// into a new candidate.
func (s *SituationService) correlateObservation(ctx context.Context, observation rez.SituationObservation) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// Without this lock, two observations about the same entities correlated at once could each find no
		// match and each start a candidate. No row represents the tenant's situations about some entities,
		// so no row lock can prevent it. AcquireTxLocks scopes the key to the tenant, making it one lock per
		// tenant. It is taken before any situation row lock.
		if lockErr := s.db.AcquireTxLocks(ctx, situationCorrelationLock, "tenant"); lockErr != nil {
			return fmt.Errorf("acquire situation correlation lock: %w", lockErr)
		}
		// Two runs for one observation can both find it unobserved before either takes the lock.
		observed, observedErr := s.refreshObservingSituations(ctx, observation.Ref)
		if observedErr != nil || observed {
			return observedErr
		}

		entityIDs, entitiesErr := s.observationRuntimeEntities(ctx, observation)
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
			Observation:   observation,
			EntityIDs:     entityIDs,
			Relationships: relationships,
			Situations:    candidates,
		}
		decision := situations.DecideCorrelation(input)
		switch {
		case !decision.Placed():
			return nil
		case decision.MatchKind == situations.MatchKindSeed:
			createParams := rez.CreateSituationParams{
				Title:        observation.Title,
				Observations: []rez.SituationObservationRef{observation.Ref},
			}
			if _, createErr := s.CreateSituation(ctx, createParams); createErr != nil {
				return fmt.Errorf("start candidate: %w", createErr)
			}
			return nil
		}
		attachParams := AttachSituationObservationsParams{
			SituationID:  decision.SituationID,
			Observations: []rez.SituationObservationRef{observation.Ref},
		}
		return s.AttachSituationObservations(ctx, attachParams)
	})
}

// correlationRelationships lists the depends-on and adjacent relationships touching the entities.
func (s *SituationService) correlationRelationships(ctx context.Context, entityIDs mapset.Set[uuid.UUID]) ([]situations.CorrelationRelationship, error) {
	if entityIDs.IsEmpty() {
		return nil, nil
	}
	ids := entityIDs.ToSlice()
	classPredicates := knowledgegraph.PredicatesWithDependency(knowledgegraph.DependencySourceOnTarget, knowledgegraph.DependencyMutual)
	listParams := rez.ListKnowledgeRelationshipsParams{
		PageSize: observationGraphPageSize,
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
// them from their observations. A situation that closed before its lock was taken is left out.
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

	queryMembers := client.SituationObservation.Query().
		Where(sitobs.SituationIDIn(ids...))
	members, membersErr := queryMembers.All(ctx)
	if membersErr != nil {
		return nil, fmt.Errorf("query correlation situation observations: %w", membersErr)
	}
	memberRefs := make([]rez.SituationObservationRef, len(members))
	for i, member := range members {
		memberRefs[i] = member.Ref()
	}

	observations, loadErr := s.LoadObservations(ctx, memberRefs, rez.LoadSituationObservationsOptions{})
	if loadErr != nil {
		return nil, loadErr
	}

	for _, member := range members {
		candidate := byID[member.SituationID]
		if candidate == nil {
			continue
		}
		observation := observations[member.Ref()]
		candidate.Active = candidate.Active || observation.Active
		for _, at := range []time.Time{member.AttachedAt, observation.LastActivityAt} {
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
