package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	sit "github.com/rezible/rezible/ent/situation"
	sitent "github.com/rezible/rezible/ent/situationentity"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	sitog "github.com/rezible/rezible/ent/situationobservationgroup"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/situations"
)

// observationGraphPageSize is large enough to read an observation's entities or relationships in one page.
const observationGraphPageSize = 10_000

// AttachSituationObservationsParams attaches observations to a situation, in its group with GroupID when set.
type AttachSituationObservationsParams struct {
	SituationID  uuid.UUID
	GroupID      *uuid.UUID
	Observations []rez.SituationObservationRef
	Explanation  string
}

// AttachSituationObservations attaches the observations the situation does not already have. Other
// situations, closed or not, may have the same observations.
func (s *SituationService) AttachSituationObservations(ctx context.Context, params AttachSituationObservationsParams) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		now := s.clock.Now()
		locked, lockErr := s.lockSituations(ctx, params.SituationID)
		if lockErr != nil {
			return lockErr
		}
		situation := locked[params.SituationID]
		if situation.ClosedAt != nil {
			return fmt.Errorf("%w: situation is closed", errs.ErrConflict)
		}
		if params.GroupID != nil {
			queryGroup := tx.SituationObservationGroup.Query().
				Where(sitog.ID(*params.GroupID), sitog.SituationID(situation.ID))
			if exists, groupErr := queryGroup.Exist(ctx); groupErr != nil {
				return fmt.Errorf("query observation group: %w", groupErr)
			} else if !exists {
				return fmt.Errorf("%w: observation group is not the situation's", errs.ErrInvalidInput)
			}
		}
		observed, observedErr := s.observedRefs(ctx, situation.ID)
		if observedErr != nil {
			return observedErr
		}
		newRefs := mapset.NewSet(params.Observations...).Difference(mapset.NewSet(observed...))
		newRefs.Remove(rez.SituationObservationRef{})
		if newRefs.IsEmpty() {
			return nil
		}
		observations, loadErr := s.LoadObservations(ctx, newRefs.ToSlice(), rez.LoadSituationObservationsOptions{})
		if loadErr != nil {
			return loadErr
		}
		explanation := strings.TrimSpace(params.Explanation)
		if observeErr := s.createObservations(ctx, situation.ID, params.GroupID, newRefs.ToSlice(), explanation, now); observeErr != nil {
			return observeErr
		}
		if lowerErr := s.lowerOpenedAt(ctx, situation, s.earliestStart(observations)); lowerErr != nil {
			return lowerErr
		}
		if entitiesErr := s.recomputeSituationEntities(ctx, situation.ID); entitiesErr != nil {
			return entitiesErr
		}
		return s.requestEvaluation(ctx, situation.ID)
	})
}

// createObservations attaches the observations to the situation, in its group when groupID is set.
func (s *SituationService) createObservations(ctx context.Context, situationID uuid.UUID, groupID *uuid.UUID, refs []rez.SituationObservationRef, explanation string, now time.Time) error {
	createObservations := s.db.Client(ctx).SituationObservation.MapCreateBulk(refs, func(c *ent.SituationObservationCreate, i int) {
		s.setObservationRef(c, refs[i]).
			SetSituationID(situationID).
			SetNillableObservationGroupID(groupID).
			SetAttachedAt(now).
			SetMatchExplanation(explanation)
	})
	if createErr := createObservations.Exec(ctx); createErr != nil {
		return fmt.Errorf("create situation observations: %w", createErr)
	}
	return nil
}

// lowerOpenedAt moves the situation's opening back to an earlier observation start; it never moves it forward.
func (s *SituationService) lowerOpenedAt(ctx context.Context, situation *ent.Situation, startedAt time.Time) error {
	if !startedAt.Before(situation.OpenedAt) {
		return nil
	}
	updateOpenedAt := s.db.Client(ctx).Situation.UpdateOneID(situation.ID).
		SetOpenedAt(startedAt)
	if updateErr := updateOpenedAt.Exec(ctx); updateErr != nil {
		return fmt.Errorf("lower situation opened time: %w", updateErr)
	}
	situation.OpenedAt = startedAt
	return nil
}

// recomputeSituationEntities replaces the situation's entity rows with the union of its observations'
// runtime entities, resolving each again. An entity is matching when an observation that is not broad
// touches it.
func (s *SituationService) recomputeSituationEntities(ctx context.Context, situationID uuid.UUID) error {
	refs, refsErr := s.observedRefs(ctx, situationID)
	if refsErr != nil {
		return refsErr
	}
	matching := make(map[uuid.UUID]bool)
	if len(refs) > 0 {
		observations, loadErr := s.LoadObservations(ctx, refs, rez.LoadSituationObservationsOptions{})
		if loadErr != nil {
			return loadErr
		}
		for _, observation := range observations {
			entityIDs, entitiesErr := s.observationRuntimeEntities(ctx, observation)
			if entitiesErr != nil {
				return entitiesErr
			}
			broad := situations.IsBroad(entityIDs.Cardinality())
			for entityID := range entityIDs.Iter() {
				matching[entityID] = matching[entityID] || !broad
			}
		}
	}

	client := s.db.Client(ctx)
	deleteEntities := client.SituationEntity.Delete().
		Where(sitent.SituationID(situationID))
	if _, deleteErr := deleteEntities.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete situation entities: %w", deleteErr)
	}
	creates := make([]*ent.SituationEntityCreate, 0, len(matching))
	for entityID, isMatching := range matching {
		createEntity := client.SituationEntity.Create().
			SetSituationID(situationID).
			SetKnowledgeEntityID(entityID).
			SetMatching(isMatching)
		creates = append(creates, createEntity)
	}
	if createErr := client.SituationEntity.CreateBulk(creates...).Exec(ctx); createErr != nil {
		return fmt.Errorf("create situation entities: %w", createErr)
	}
	return nil
}

// observationRuntimeEntities returns the runtime entities an observation is about: the entities its events
// observed, and the endpoints of relationships they observed, resolved to the runtime structure level.
// Evidence that an entity was deleted does not make an observation about it.
func (s *SituationService) observationRuntimeEntities(ctx context.Context, observation rez.SituationObservation) (mapset.Set[uuid.UUID], error) {
	runtime := mapset.NewSet[uuid.UUID]()
	if len(observation.EventIDs) == 0 {
		return runtime, nil
	}
	observed := ksa.HasEvidenceWith(knev.EventIDIn(observation.EventIDs...), knev.KindEQ(knev.KindObserved))
	pageParams := ent.ListParams{PageSize: observationGraphPageSize}

	listEntitiesParams := rez.ListKnowledgeEntitiesParams{
		ListParams: pageParams,
		Predicates: []predicate.KnowledgeEntity{kne.HasAliasesWith(observed)},
	}
	entities, entitiesErr := s.graph.ListEntities(ctx, listEntitiesParams)
	if entitiesErr != nil {
		return nil, fmt.Errorf("list observed entities: %w", entitiesErr)
	}
	listRelationshipsParams := rez.ListKnowledgeRelationshipsParams{
		ListParams: pageParams,
		Predicates: []predicate.KnowledgeRelationship{knr.HasAliasesWith(observed)},
	}
	relationships, relationshipsErr := s.graph.ListRelationships(ctx, listRelationshipsParams)
	if relationshipsErr != nil {
		return nil, fmt.Errorf("list observed relationships: %w", relationshipsErr)
	}

	named := mapset.NewSet[uuid.UUID]()
	for _, entity := range entities.Data {
		named.Add(entity.ID)
	}
	for _, relationship := range relationships.Data {
		named.Append(relationship.SourceEntityID, relationship.TargetEntityID)
	}
	if named.IsEmpty() {
		return runtime, nil
	}
	representatives, resolveErr := s.graph.ResolveRepresentatives(ctx, named.ToSlice(), situations.RepresentativeCategories, situations.EntityResolutionDepth)
	if resolveErr != nil {
		return nil, fmt.Errorf("resolve observation entities: %w", resolveErr)
	}
	for _, representative := range representatives {
		runtime.Add(representative)
	}
	return runtime, nil
}

// MergeSituations moves the source's observations into the target and closes the source as merged.
func (s *SituationService) MergeSituations(ctx context.Context, params rez.MergeSituationsParams) (*ent.Situation, error) {
	if params.SourceID == uuid.Nil || params.TargetID == uuid.Nil {
		return nil, fmt.Errorf("%w: source and target situations are required", errs.ErrInvalidInput)
	}
	if params.SourceID == params.TargetID {
		return nil, fmt.Errorf("%w: a situation cannot be merged into itself", errs.ErrInvalidInput)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Situation, error) {
		now := s.clock.Now()
		locked, lockErr := s.lockSituations(ctx, params.SourceID, params.TargetID)
		if lockErr != nil {
			return nil, lockErr
		}
		source := locked[params.SourceID]
		target := locked[params.TargetID]
		if source.ClosedAt != nil || target.ClosedAt != nil {
			return nil, fmt.Errorf("%w: closed situations cannot be merged", errs.ErrConflict)
		}
		explanation := strings.TrimSpace(params.Explanation)
		if moveErr := s.moveObservations(ctx, source.ID, target.ID, explanation, now); moveErr != nil {
			return nil, moveErr
		}
		if closeErr := s.closeLocked(ctx, source, sit.CloseReasonMerged, explanation, now, now); closeErr != nil {
			return nil, closeErr
		}
		if lowerErr := s.lowerOpenedAt(ctx, target, source.OpenedAt); lowerErr != nil {
			return nil, lowerErr
		}
		for _, situationID := range []uuid.UUID{source.ID, target.ID} {
			if entitiesErr := s.recomputeSituationEntities(ctx, situationID); entitiesErr != nil {
				return nil, entitiesErr
			}
		}
		if evaluateErr := s.requestEvaluation(ctx, target.ID); evaluateErr != nil {
			return nil, evaluateErr
		}
		return s.GetSituation(ctx, sit.ID(target.ID))
	})
}

// moveObservations moves the source's observations to the target as attachments made now; the reset
// attachment time counts the merge as new evidence for the target, and the reset observed revision brings
// them into the target's investigation. The source's observations the target already has are dropped. A
// source group joins the target's group of the same title; other groups move.
func (s *SituationService) moveObservations(ctx context.Context, sourceID, targetID uuid.UUID, explanation string, now time.Time) error {
	client := s.db.Client(ctx)
	targetRefs, targetRefsErr := s.observedRefs(ctx, targetID)
	if targetRefsErr != nil {
		return targetRefsErr
	}
	deleteDuplicates := client.SituationObservation.Delete().
		Where(sitobs.SituationID(sourceID), s.observesAny(targetRefs...))
	if _, deleteErr := deleteDuplicates.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete duplicate observations: %w", deleteErr)
	}

	queryTargetGroups := client.SituationObservationGroup.Query().
		Where(sitog.SituationID(targetID))
	targetGroups, targetErr := queryTargetGroups.All(ctx)
	if targetErr != nil {
		return fmt.Errorf("load target observation groups: %w", targetErr)
	}
	targetGroupByTitle := make(map[string]uuid.UUID, len(targetGroups))
	for _, group := range targetGroups {
		if _, exists := targetGroupByTitle[group.Title]; !exists {
			targetGroupByTitle[group.Title] = group.ID
		}
	}
	querySourceGroups := client.SituationObservationGroup.Query().
		Where(sitog.SituationID(sourceID))
	sourceGroups, sourceErr := querySourceGroups.All(ctx)
	if sourceErr != nil {
		return fmt.Errorf("load source observation groups: %w", sourceErr)
	}
	for _, group := range sourceGroups {
		destinationID, joins := targetGroupByTitle[group.Title]
		if !joins {
			moveGroup := client.SituationObservationGroup.UpdateOneID(group.ID).
				SetSituationID(targetID)
			if moveErr := moveGroup.Exec(ctx); moveErr != nil {
				return fmt.Errorf("move observation group: %w", moveErr)
			}
			continue
		}
		joinGroup := client.SituationObservation.Update().
			Where(sitobs.ObservationGroupID(group.ID)).
			SetObservationGroupID(destinationID)
		if joinErr := joinGroup.Exec(ctx); joinErr != nil {
			return fmt.Errorf("join observation group: %w", joinErr)
		}
		if deleteErr := client.SituationObservationGroup.DeleteOneID(group.ID).Exec(ctx); deleteErr != nil {
			return fmt.Errorf("delete joined observation group: %w", deleteErr)
		}
	}

	moveObservations := client.SituationObservation.Update().
		Where(sitobs.SituationID(sourceID)).
		SetSituationID(targetID).
		SetMatchExplanation(explanation).
		SetAttachedAt(now).
		SetObservedRevision(0)
	if moveErr := moveObservations.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move situation observations: %w", moveErr)
	}
	return nil
}
