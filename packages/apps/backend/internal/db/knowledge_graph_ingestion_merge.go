package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	sitent "github.com/rezible/rezible/ent/situationentity"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
)

// mergeEntitiesLockAttempts bounds how often a merge locks the entities again after an overlapping merge moved
// one of its names while it waited for the locks.
const mergeEntitiesLockAttempts = 3

func (s *KnowledgeGraphIngestionService) MergeEntities(ctx context.Context, params rez.MergeEntitiesParams) (*rez.MergeEntitiesResult, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.MergeEntitiesResult, error) {
		proposal, proposalErr := tx.KnowledgeProposal.Get(ctx, params.ProposalID)
		if proposalErr != nil {
			return nil, fmt.Errorf("get proposal: %w", proposalErr)
		}

		entities, lockErr := s.lockMergingEntities(ctx, proposal, params.EntityIDs)
		if lockErr != nil {
			return nil, lockErr
		}
		if len(entities) == 1 {
			return &rez.MergeEntitiesResult{SurvivorID: entities[0].ID}, nil
		}
		survivor, other := entities[0], entities[1]
		if other.CreatedAt.Before(survivor.CreatedAt) {
			survivor, other = other, survivor
		}
		if kindErr := checkMergeableKinds(survivor, other); kindErr != nil {
			return nil, kindErr
		}

		m := &knowledgeEntityMerge{
			tx:         tx,
			proposalID: proposal.ID,
			survivor:   survivor,
			other:      other,
		}
		if mergeErr := m.run(ctx); mergeErr != nil {
			return nil, mergeErr
		}
		return &rez.MergeEntitiesResult{SurvivorID: survivor.ID, Merged: true}, nil
	})
}

// lockMergingEntities locks the entities holding the proposal's names FOR UPDATE, in ID order, so that nothing
// can refer to them anew until the merge commits: every write that points at an entity takes a key-share lock on
// it, through its foreign key or lockResolvedEntity. After the locks are held it resolves the names again, since
// an overlapping merge (`a`-`b` and `b`-`c`) may have committed while it waited, and locks the entities it now
// finds. It returns one entity when the names are on one already.
func (s *KnowledgeGraphIngestionService) lockMergingEntities(ctx context.Context, proposal *ent.KnowledgeProposal, entityIDs [2]uuid.UUID) ([]*ent.KnowledgeEntity, error) {
	client := s.db.Client(ctx)
	lockIDs := entityIDs[:]
	for range mergeEntitiesLockAttempts {
		lockEntities := client.KnowledgeEntity.Query().
			Where(kne.IDIn(lockIDs...)).
			Order(kne.ByID()).
			ForUpdate()
		locked, lockErr := lockEntities.All(ctx)
		if lockErr != nil {
			return nil, fmt.Errorf("lock entities: %w", lockErr)
		}

		queryHolders := client.KnowledgeEntityLinkingAttribute.Query().
			Where(kela.Attribute(proposal.Attribute), kela.ValueIn(proposal.ValueA, proposal.ValueB))
		holders, holdersErr := queryHolders.All(ctx)
		if holdersErr != nil {
			return nil, fmt.Errorf("query name holders: %w", holdersErr)
		}
		if len(holders) != 2 {
			return nil, errs.Wrap(errs.ErrConflict, errs.CodeConflict, "One of the names no longer belongs to anything Rezible knows.")
		}
		heldBy := mapset.NewSet(holders[0].EntityID, holders[1].EntityID)
		lockedIDs := mapset.NewSet[uuid.UUID]()
		lockedByID := make(map[uuid.UUID]*ent.KnowledgeEntity, len(locked))
		for _, entity := range locked {
			lockedIDs.Add(entity.ID)
			lockedByID[entity.ID] = entity
		}
		if heldBy.IsSubset(lockedIDs) {
			matched := make([]*ent.KnowledgeEntity, 0, heldBy.Cardinality())
			for _, entity := range locked {
				if heldBy.Contains(entity.ID) {
					matched = append(matched, entity)
				}
			}
			return matched, nil
		}
		lockIDs = heldBy.ToSlice()
	}
	return nil, errs.Wrap(errs.ErrConflict, errs.CodeConflict, "The names kept moving while they were merged. Try again.")
}

// checkMergeableKinds refuses to merge entities of two kinds, or of a kind a confirmed match cannot merge.
func checkMergeableKinds(survivor, other *ent.KnowledgeEntity) error {
	if survivor.Category != other.Category || survivor.Kind != other.Kind {
		return errs.Wrap(errs.ErrConflict, errs.CodeConflict, fmt.Sprintf(
			"Only entities of one kind can be merged: one is a %s, the other a %s.", survivor.Kind, other.Kind))
	}
	if !knowledgegraph.IsMergeable(survivor.Category, survivor.Kind) {
		return errs.Wrap(errs.ErrConflict, errs.CodeConflict, fmt.Sprintf(
			"Only services and repositories can be merged, not a %s.", survivor.Kind))
	}
	return nil
}

// knowledgeEntityMerge moves everything of other onto survivor and deletes other. Both rows are locked, so every
// move is a set-based update of what refers to them now.
type knowledgeEntityMerge struct {
	tx         *ent.Client
	proposalID uuid.UUID
	survivor   *ent.KnowledgeEntity
	other      *ent.KnowledgeEntity
}

func (m *knowledgeEntityMerge) run(ctx context.Context) error {
	moveLinkingAttributes := m.tx.KnowledgeEntityLinkingAttribute.Update().
		Where(kela.EntityID(m.other.ID)).
		SetEntityID(m.survivor.ID)
	if moveErr := moveLinkingAttributes.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move linking attributes: %w", moveErr)
	}

	moveAliases := m.tx.KnowledgeSubjectAlias.Update().
		Where(ksa.EntityID(m.other.ID)).
		SetEntityID(m.survivor.ID).
		SetMergedByProposalID(m.proposalID)
	if moveErr := moveAliases.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move aliases: %w", moveErr)
	}

	if relationshipsErr := m.moveRelationships(ctx); relationshipsErr != nil {
		return fmt.Errorf("move relationships: %w", relationshipsErr)
	}

	if matchingErr := m.mergeSituationMatching(ctx); matchingErr != nil {
		return fmt.Errorf("merge situation matching: %w", matchingErr)
	}
	for _, reference := range ent.KnowledgeEntityReferences {
		if repointErr := m.repoint(ctx, reference, m.other.ID, m.survivor.ID); repointErr != nil {
			return fmt.Errorf("re-point %s.%s: %w", reference.Table, reference.Column, repointErr)
		}
	}

	if stateErr := m.foldState(ctx); stateErr != nil {
		return fmt.Errorf("fold state: %w", stateErr)
	}

	if deleteErr := m.tx.KnowledgeEntity.DeleteOneID(m.other.ID).Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete merged entity: %w", deleteErr)
	}
	return nil
}

// knowledgeRelationshipKey is a relationship's identity under its unique index.
type knowledgeRelationshipKey struct {
	predicate      knr.Predicate
	sourceEntityID uuid.UUID
	targetEntityID uuid.UUID
}

// moveRelationships re-points the other entity's relationships to the survivor. A relationship that would then
// repeat one the survivor already has, or another moved one, is folded into it: its aliases and references move
// onto that relationship, and it is deleted. A relationship between the two becomes one from the survivor to
// itself. The relationships it reads to find duplicates cannot change meanwhile: each has one of the two locked
// entities as an endpoint, and inserting or re-pointing one takes a key-share lock on its endpoints.
func (m *knowledgeEntityMerge) moveRelationships(ctx context.Context) error {
	queryMoving := m.tx.KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityID(m.other.ID), knr.TargetEntityID(m.other.ID))).
		Order(knr.ByCreatedAt(), knr.ByID())
	moving, movingErr := queryMoving.All(ctx)
	if movingErr != nil {
		return fmt.Errorf("query relationships: %w", movingErr)
	}
	if len(moving) == 0 {
		return nil
	}

	movedKey := func(relationship *ent.KnowledgeRelationship) knowledgeRelationshipKey {
		key := knowledgeRelationshipKey{
			predicate:      relationship.Predicate,
			sourceEntityID: relationship.SourceEntityID,
			targetEntityID: relationship.TargetEntityID,
		}
		if key.sourceEntityID == m.other.ID {
			key.sourceEntityID = m.survivor.ID
		}
		if key.targetEntityID == m.other.ID {
			key.targetEntityID = m.survivor.ID
		}
		return key
	}
	existingPreds := make([]predicate.KnowledgeRelationship, 0, len(moving))
	for _, relationship := range moving {
		key := movedKey(relationship)
		existingPreds = append(existingPreds, knr.And(
			knr.PredicateEQ(key.predicate),
			knr.SourceEntityID(key.sourceEntityID),
			knr.TargetEntityID(key.targetEntityID),
		))
	}
	queryExisting := m.tx.KnowledgeRelationship.Query().
		Where(knr.Or(existingPreds...))
	existing, existingErr := queryExisting.All(ctx)
	if existingErr != nil {
		return fmt.Errorf("query survivor relationships: %w", existingErr)
	}

	keptByKey := make(map[knowledgeRelationshipKey]uuid.UUID, len(existing)+len(moving))
	for _, relationship := range existing {
		keptByKey[movedKey(relationship)] = relationship.ID
	}
	var keptOrder []uuid.UUID
	foldedInto := make(map[uuid.UUID][]uuid.UUID)
	for _, relationship := range moving {
		key := movedKey(relationship)
		kept, isDuplicate := keptByKey[key]
		if !isDuplicate {
			keptByKey[key] = relationship.ID
			continue
		}
		if _, seen := foldedInto[kept]; !seen {
			keptOrder = append(keptOrder, kept)
		}
		foldedInto[kept] = append(foldedInto[kept], relationship.ID)
	}

	if lockErr := m.lockFoldedRelationships(ctx, foldedInto); lockErr != nil {
		return lockErr
	}
	for _, kept := range keptOrder {
		if foldErr := m.foldRelationships(ctx, kept, foldedInto[kept]); foldErr != nil {
			return foldErr
		}
	}

	moveSources := m.tx.KnowledgeRelationship.Update().
		Where(knr.SourceEntityID(m.other.ID)).
		SetSourceEntityID(m.survivor.ID)
	if moveErr := moveSources.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move relationship sources: %w", moveErr)
	}
	moveTargets := m.tx.KnowledgeRelationship.Update().
		Where(knr.TargetEntityID(m.other.ID)).
		SetTargetEntityID(m.survivor.ID)
	if moveErr := moveTargets.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move relationship targets: %w", moveErr)
	}
	return nil
}

// lockFoldedRelationships locks the relationships to be folded and those they fold into FOR UPDATE, in ID order,
// after the entity locks and before anything moves. A foreign key to a relationship takes a key-share lock on the
// relationship, not on its endpoints, so the entity locks do not stop a writer from inserting a reference to a
// duplicate between its references being re-pointed and its deletion, which would then fail the merge. With the
// lock, such a writer either commits first and its reference is re-pointed with the rest, or waits and fails on
// the foreign key once the duplicate is gone. The locks are held until the merge commits.
func (m *knowledgeEntityMerge) lockFoldedRelationships(ctx context.Context, foldedInto map[uuid.UUID][]uuid.UUID) error {
	if len(foldedInto) == 0 {
		return nil
	}
	lockIDs := make([]uuid.UUID, 0, 2*len(foldedInto))
	for kept, duplicates := range foldedInto {
		lockIDs = append(lockIDs, kept)
		lockIDs = append(lockIDs, duplicates...)
	}
	lockRelationships := m.tx.KnowledgeRelationship.Query().
		Where(knr.IDIn(lockIDs...)).
		Order(knr.ByID()).
		ForUpdate()
	if _, lockErr := lockRelationships.IDs(ctx); lockErr != nil {
		return fmt.Errorf("lock folded relationships: %w", lockErr)
	}
	return nil
}

func (m *knowledgeEntityMerge) foldRelationships(ctx context.Context, kept uuid.UUID, duplicates []uuid.UUID) error {
	moveAliases := m.tx.KnowledgeSubjectAlias.Update().
		Where(ksa.RelationshipIDIn(duplicates...)).
		SetRelationshipID(kept).
		SetMergedByProposalID(m.proposalID)
	if moveErr := moveAliases.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move duplicate relationship aliases: %w", moveErr)
	}
	for _, duplicate := range duplicates {
		for _, reference := range ent.KnowledgeRelationshipReferences {
			if repointErr := m.repoint(ctx, reference, duplicate, kept); repointErr != nil {
				return fmt.Errorf("re-point %s.%s: %w", reference.Table, reference.Column, repointErr)
			}
		}
	}
	deleteDuplicates := m.tx.KnowledgeRelationship.Delete().
		Where(knr.IDIn(duplicates...))
	if _, deleteErr := deleteDuplicates.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete duplicate relationships: %w", deleteErr)
	}
	return nil
}

// foldState sets the survivor's state to the two states folded, the newer applied last, with the later time.
func (m *knowledgeEntityMerge) foldState(ctx context.Context) error {
	older, newer := m.other, m.survivor
	if m.other.StateEffectiveAt != nil &&
		(m.survivor.StateEffectiveAt == nil || m.other.StateEffectiveAt.After(*m.survivor.StateEffectiveAt)) {
		older, newer = m.survivor, m.other
	}
	updateState := m.tx.KnowledgeEntity.UpdateOneID(m.survivor.ID).
		SetState(older.State.MergeFrom(newer.State)).
		SetNillableStateEffectiveAt(newer.StateEffectiveAt)
	return updateState.Exec(ctx)
}

// mergeSituationMatching keeps an open situation attracting signals through the survivor when either entity did:
// in every situation where the other's row is matching, the survivor's becomes matching. Re-pointing then deletes
// the other's row in situations that hold both.
func (m *knowledgeEntityMerge) mergeSituationMatching(ctx context.Context) error {
	queryMatching := m.tx.SituationEntity.Query().
		Where(sitent.KnowledgeEntityID(m.other.ID), sitent.Matching(true))
	matching, queryErr := queryMatching.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query matching situation entities: %w", queryErr)
	}
	if len(matching) == 0 {
		return nil
	}
	situationIDs := make([]uuid.UUID, len(matching))
	for i, row := range matching {
		situationIDs[i] = row.SituationID
	}
	setMatching := m.tx.SituationEntity.Update().
		Where(sitent.KnowledgeEntityID(m.survivor.ID), sitent.SituationIDIn(situationIDs...), sitent.Matching(false)).
		SetMatching(true)
	return setMatching.Exec(ctx)
}

// repoint makes the rows of reference referring to from refer to to. For a reference that deletes duplicates, it
// first deletes the rows that would then repeat one of to's under the unique key; for any other, such a repeat
// fails the merge.
func (m *knowledgeEntityMerge) repoint(ctx context.Context, reference ent.KnowledgeReference, from, to uuid.UUID) error {
	referringToFrom := sql.FieldEQ(reference.Column, from)
	if reference.DeleteDuplicates {
		// Ent's predicates cannot compare a row with another of its table, so this is an EXISTS subquery. It
		// matches the whole unique key, tenant_id included, so a row only counts as a duplicate of the survivor's
		// row in the same tenant and scope.
		sharesUniqueKey := func(s *sql.Selector) {
			row := sql.Table(reference.Table).As("merge_survivor_row")
			match := sql.Select(row.C(reference.Column)).
				From(row).
				Where(sql.EQ(row.C(reference.Column), to))
			for _, column := range reference.UniqueWith {
				match.Where(sql.ColumnsEQ(row.C(column), s.C(column)))
			}
			s.Where(sql.Exists(match))
		}
		duplicateOf := func(s *sql.Selector) {
			referringToFrom(s)
			sharesUniqueKey(s)
		}
		if _, deleteErr := reference.Delete(ctx, m.tx, duplicateOf); deleteErr != nil {
			return fmt.Errorf("delete duplicates: %w", deleteErr)
		}
	}
	return reference.Update(ctx, m.tx, referringToFrom, to)
}
