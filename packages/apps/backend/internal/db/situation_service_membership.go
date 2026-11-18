package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	osc "github.com/rezible/rezible/ent/observationscope"
	osce "github.com/rezible/rezible/ent/observationscopeentity"
	"github.com/rezible/rezible/ent/predicate"
	sit "github.com/rezible/rezible/ent/situation"
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

// storeObservationScope computes the observation's scope and stores it. A stored scope from the same or a newer
// revision is kept. An empty scope is computed again, since the graph may have learned the observation's
// entities since.
func (s *SituationService) storeObservationScope(ctx context.Context, observation rez.SituationObservation) error {
	queryCurrent := s.db.Client(ctx).ObservationScope.Query().
		Where(s.scopeOf(observation.Ref), osc.RevisionGTE(observation.Revision), osc.HasEntities())
	if current, currentErr := queryCurrent.Exist(ctx); currentErr != nil {
		return fmt.Errorf("query observation scope: %w", currentErr)
	} else if current {
		return nil
	}
	_, computeErr := s.computeObservationScope(ctx, observation, false)
	return computeErr
}

// computeObservationScope computes the observation's scope and stores it where the stored one may be replaced,
// reporting whether the scope's entities changed. A stored scope of a newer revision is always kept. One of the
// same revision is replaced when restructured is set, since the structure its entities were resolved through
// changed; otherwise only when it is empty.
func (s *SituationService) computeObservationScope(ctx context.Context, observation rez.SituationObservation, restructured bool) (bool, error) {
	entityIDs, entitiesErr := s.observationRuntimeEntities(ctx, observation)
	if entitiesErr != nil {
		return false, entitiesErr
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (bool, error) {
		create, refColumn := s.setScopeRef(tx.ObservationScope.Create(), observation.Ref)
		create.SetRevision(observation.Revision).
			SetBroad(situations.IsBroad(entityIDs.Cardinality()))
		// The conflicting row stays locked until commit whether or not it is updated, so concurrent jobs for one
		// observation replace its entities one at a time, each against the revision the previous one stored.
		// Entities found for the stored revision replace only an empty scope stored for it. Ent's predicates cannot
		// read the conflicting row's entities, so this is a NOT EXISTS subquery on it.
		builder := sql.Dialect(dialect.Postgres)
		stored, entities := builder.Table(osc.Table), builder.Table(osce.Table)
		replaceable := sql.LT(stored.C(osc.FieldRevision), observation.Revision)
		if restructured {
			replaceable = sql.LTE(stored.C(osc.FieldRevision), observation.Revision)
		} else if !entityIDs.IsEmpty() {
			storedEntities := builder.Select(entities.C(osce.FieldID)).
				From(entities).
				Where(sql.ColumnsEQ(entities.C(osce.FieldScopeID), stored.C(osc.FieldID)))
			replaceable = sql.Or(
				replaceable,
				sql.And(sql.EQ(stored.C(osc.FieldRevision), observation.Revision), sql.NotExists(storedEntities)),
			)
		}
		upsert := create.OnConflict(sql.ConflictColumns(refColumn), sql.UpdateWhere(replaceable)).
			UpdateRevision().
			UpdateBroad()
		scopeID, upsertErr := upsert.ID(ctx)
		if upsertErr != nil {
			if errors.Is(upsertErr, stdsql.ErrNoRows) {
				return false, nil
			}
			return false, fmt.Errorf("upsert observation scope: %w", upsertErr)
		}
		queryStored := tx.ObservationScopeEntity.Query().
			Where(osce.ScopeID(scopeID))
		storedEntities, storedErr := queryStored.All(ctx)
		if storedErr != nil {
			return false, fmt.Errorf("query observation scope entities: %w", storedErr)
		}
		storedIDs := mapset.NewSet[uuid.UUID]()
		for _, entity := range storedEntities {
			storedIDs.Add(entity.KnowledgeEntityID)
		}
		if storedIDs.Equal(entityIDs) {
			return false, nil
		}
		deleteEntities := tx.ObservationScopeEntity.Delete().
			Where(osce.ScopeID(scopeID))
		if _, deleteErr := deleteEntities.Exec(ctx); deleteErr != nil {
			return false, fmt.Errorf("delete observation scope entities: %w", deleteErr)
		}
		ids := entityIDs.ToSlice()
		// Two of the entities can be merged into one survivor after they were resolved, and the canonical reference
		// hook stores both as the survivor, so a repeat is skipped.
		createEntities := tx.ObservationScopeEntity.MapCreateBulk(ids, func(c *ent.ObservationScopeEntityCreate, i int) {
			c.SetScopeID(scopeID).
				SetKnowledgeEntityID(ids[i])
		}).
			OnConflictColumns(osce.FieldTenantID, osce.FieldScopeID, osce.FieldKnowledgeEntityID).
			DoNothing()
		if createErr := createEntities.Exec(ctx); createErr != nil {
			return false, fmt.Errorf("create observation scope entities: %w", createErr)
		}
		return true, nil
	})
}

// RefreshObservationScopes computes again the scopes of the observations of unclosed situations, and requests
// the evaluation of the situations whose entities changed. It runs after the knowledge structure changed:
// a scope is otherwise computed only when its observation's revision rises. Repeating it changes nothing.
func (s *SituationService) RefreshObservationScopes(ctx context.Context) error {
	queryObserved := s.db.Client(ctx).SituationObservation.Query().
		Where(sitobs.HasSituationWith(sit.ClosedAtIsNil()))
	observed, queryErr := queryObserved.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query observations of unclosed situations: %w", queryErr)
	}
	refs := mapset.NewSet[rez.SituationObservationRef]()
	for _, observation := range observed {
		refs.Add(observation.Ref())
	}
	if refs.IsEmpty() {
		return nil
	}
	observations, loadErr := s.LoadObservations(ctx, refs.ToSlice(), rez.LoadSituationObservationsOptions{})
	if loadErr != nil {
		return loadErr
	}
	for ref, observation := range observations {
		changed, computeErr := s.computeObservationScope(ctx, observation, true)
		if computeErr != nil {
			return computeErr
		}
		if !changed {
			continue
		}
		if _, refreshErr := s.refreshObservingSituations(ctx, ref); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}

// storedScopeEntities reads the entities of the observation's stored scope.
func (s *SituationService) storedScopeEntities(ctx context.Context, ref rez.SituationObservationRef) (mapset.Set[uuid.UUID], error) {
	queryEntities := s.db.Client(ctx).ObservationScopeEntity.Query().
		Where(osce.HasScopeWith(s.scopeOf(ref)))
	entities, queryErr := queryEntities.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query observation scope entities: %w", queryErr)
	}
	entityIDs := mapset.NewSet[uuid.UUID]()
	for _, entity := range entities {
		entityIDs.Add(entity.KnowledgeEntityID)
	}
	return entityIDs, nil
}

// MergeSituations moves the source's observations into the target and closes the source as merged into it.
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
		if closeErr := s.closeLocked(ctx, source, sit.CloseReasonMerged, &target.ID, explanation, now, now); closeErr != nil {
			return nil, closeErr
		}
		// Situations already merged into the source follow it, so a merge target is never itself merged.
		queryRedirected := tx.Situation.Query().Where(sit.MergedIntoID(source.ID))
		redirectedIDs, redirectedErr := queryRedirected.IDs(ctx)
		if redirectedErr != nil {
			return nil, fmt.Errorf("query merged situations: %w", redirectedErr)
		}
		redirectSources := tx.Situation.Update().
			Where(sit.IDIn(redirectedIDs...)).
			SetMergedIntoID(target.ID)
		if redirectErr := redirectSources.Exec(ctx); redirectErr != nil {
			return nil, fmt.Errorf("redirect merged situations: %w", redirectErr)
		}
		for _, redirectedID := range redirectedIDs {
			if publishErr := s.msgs.Publish(ctx, rez.EventOnSituationUpdated{SituationID: redirectedID}); publishErr != nil {
				return nil, fmt.Errorf("publish situation updated: %w", publishErr)
			}
		}
		if lowerErr := s.lowerOpenedAt(ctx, target, source.OpenedAt); lowerErr != nil {
			return nil, lowerErr
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
