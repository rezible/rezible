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

		// Clearing the parent flags around the two entities writes is_parent, so it holds the tenant's structure
		// lock, as placement does (placeParents). The merge takes it after locking its two entities and before
		// flush locks the rest, so it can wait for another flush's row locks while holding it; a deadlock aborts one
		// transaction and its request or job retries.
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeStructureLockNamespace, "tenant"); lockErr != nil {
			return nil, fmt.Errorf("lock knowledge structure: %w", lockErr)
		}
		m := &knowledgeEntityMerge{
			tx:         tx,
			proposalID: proposal.ID,
			survivor:   survivor,
			other:      other,
			refresh:    newSubjectRefresh(),
		}
		if mergeErr := m.run(ctx); mergeErr != nil {
			return nil, mergeErr
		}
		if refreshErr := s.flush(ctx, m.refresh); refreshErr != nil {
			return nil, fmt.Errorf("refresh merged subjects: %w", refreshErr)
		}
		return &rez.MergeEntitiesResult{SurvivorID: survivor.ID, Merged: true}, nil
	})
}

// lockMergingEntities locks the entities holding the proposal's names FOR UPDATE, in ID order, which blocks every
// writer's key-share lock until the merge commits. It then resolves the names again, since an overlapping merge may
// have moved one while it waited. It returns one entity when both names are already on it.
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

// knowledgeEntityMerge moves everything of other onto survivor and leaves other as a redirect to it. Both rows are
// locked, so every move is a set-based update of what refers to them now. It touches the survivor and the
// relationships it moved or folded in refresh, and marks the survivor and the targets of the moved and folded
// claims for placement, for the caller to flush: a move neither creates a claim nor flips its withdrawal.
type knowledgeEntityMerge struct {
	tx         *ent.Client
	proposalID uuid.UUID
	survivor   *ent.KnowledgeEntity
	other      *ent.KnowledgeEntity
	refresh    *subjectRefresh
}

func (m *knowledgeEntityMerge) run(ctx context.Context) error {
	if clearErr := m.clearParentClaims(ctx); clearErr != nil {
		return fmt.Errorf("clear parent claims: %w", clearErr)
	}

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

	for _, reference := range ent.KnowledgeEntityReferences {
		if repointErr := m.repoint(ctx, reference, m.other.ID, m.survivor.ID); repointErr != nil {
			return fmt.Errorf("re-point %s.%s: %w", reference.Table, reference.Column, repointErr)
		}
	}

	// Redirects to other move to the survivor too, so redirects stay flat.
	redirect := m.tx.KnowledgeEntity.Update().
		Where(kne.Or(kne.ID(m.other.ID), kne.MergedIntoID(m.other.ID))).
		SetMergedIntoID(m.survivor.ID)
	if redirectErr := redirect.Exec(ctx); redirectErr != nil {
		return fmt.Errorf("redirect merged entity: %w", redirectErr)
	}
	// The survivor gained aliases. Every entity whose parent was the other is now the target of a moved claim,
	// and the moved claims may change the survivor's own parent.
	m.refresh.entities.Add(m.survivor.ID)
	m.refresh.placements.Add(m.survivor.ID)
	return nil
}

// clearParentClaims clears is_parent on every relationship with either entity as an endpoint, before any
// endpoint moves: a moved flagged claim could give the survivor a second flagged parent claim, or turn a flagged
// claim between the two into a flagged one from the survivor to itself. It touches the cleared relationships and
// marks their targets for placement, so flush flags each child's claim again. The caller holds the structure lock.
func (m *knowledgeEntityMerge) clearParentClaims(ctx context.Context) error {
	endpoints := []uuid.UUID{m.survivor.ID, m.other.ID}
	queryFlagged := m.tx.KnowledgeRelationship.Query().
		Where(
			knr.IsParent(true),
			knr.Or(knr.SourceEntityIDIn(endpoints...), knr.TargetEntityIDIn(endpoints...)),
		)
	flagged, flaggedErr := queryFlagged.All(ctx)
	if flaggedErr != nil {
		return fmt.Errorf("query flagged claims: %w", flaggedErr)
	}
	if len(flagged) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(flagged))
	for i, relationship := range flagged {
		ids[i] = relationship.ID
		m.refresh.relationships.Add(relationship.ID)
		m.refresh.placements.Add(relationship.TargetEntityID)
	}
	clearFlags := m.tx.KnowledgeRelationship.Update().
		Where(knr.IDIn(ids...)).
		SetIsParent(false)
	return clearFlags.Exec(ctx)
}

// knowledgeRelationshipKey is a relationship's identity under its unique index.
type knowledgeRelationshipKey struct {
	predicate      knr.Predicate
	sourceEntityID uuid.UUID
	targetEntityID uuid.UUID
}

// moveRelationships re-points the other entity's relationships to the survivor. One that would duplicate an
// existing relationship is folded into it: its aliases and references move over, and it stays as a redirect with
// its old endpoints. The relationships it reads cannot change meanwhile, because both endpoints' entities are
// locked.
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
		if key.predicate == knr.PredicateContains {
			m.refresh.placements.Add(key.targetEntityID)
		}
		if !isDuplicate {
			keptByKey[key] = relationship.ID
			m.refresh.relationships.Add(relationship.ID)
			continue
		}
		if _, seen := foldedInto[kept]; !seen {
			keptOrder = append(keptOrder, kept)
			m.refresh.relationships.Add(kept)
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

	// Queries leave redirects out, updates do not, so the moves below exclude them themselves. A relationship from
	// the other to itself moves both endpoints at once: moving one would leave it from the survivor to the other, or
	// back, which is the key of a relationship folded into it.
	moveSelfLoops := m.tx.KnowledgeRelationship.Update().
		Where(knr.SourceEntityID(m.other.ID), knr.TargetEntityID(m.other.ID), knr.MergedIntoIDIsNil()).
		SetSourceEntityID(m.survivor.ID).
		SetTargetEntityID(m.survivor.ID)
	if moveErr := moveSelfLoops.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move relationships to the other itself: %w", moveErr)
	}
	moveSources := m.tx.KnowledgeRelationship.Update().
		Where(knr.SourceEntityID(m.other.ID), knr.MergedIntoIDIsNil()).
		SetSourceEntityID(m.survivor.ID)
	if moveErr := moveSources.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move relationship sources: %w", moveErr)
	}
	moveTargets := m.tx.KnowledgeRelationship.Update().
		Where(knr.TargetEntityID(m.other.ID), knr.MergedIntoIDIsNil()).
		SetTargetEntityID(m.survivor.ID)
	if moveErr := moveTargets.Exec(ctx); moveErr != nil {
		return fmt.Errorf("move relationship targets: %w", moveErr)
	}
	return nil
}

// lockFoldedRelationships locks the relationships to be folded and those they fold into FOR UPDATE, in ID order,
// after the entity locks and before anything moves, and holds them until the merge commits. A writer naming one of
// them then either commits before its reference is re-pointed, or waits and stores the relationship it folds into.
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
	// Redirects to the duplicates move to kept too, so redirects stay flat.
	redirect := m.tx.KnowledgeRelationship.Update().
		Where(knr.Or(knr.IDIn(duplicates...), knr.MergedIntoIDIn(duplicates...))).
		SetMergedIntoID(kept)
	if redirectErr := redirect.Exec(ctx); redirectErr != nil {
		return fmt.Errorf("redirect duplicate relationships: %w", redirectErr)
	}
	return nil
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
