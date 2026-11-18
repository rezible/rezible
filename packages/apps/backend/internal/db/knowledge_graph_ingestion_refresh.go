package db

import (
	"context"
	"fmt"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
)

// subjectRefresh collects the subjects a write touched, so that flush refreshes each once, at the end of the
// write's transaction.
type subjectRefresh struct {
	entities, relationships mapset.Set[uuid.UUID]
	// placements are entities whose parent the write may have changed without creating a claim or flipping a
	// withdrawal: a merge's survivor and the targets of the claims it moved or folded. Each is a touched entity
	// or the target of a touched contains relationship, so flush locks it.
	placements mapset.Set[uuid.UUID]
	// reranked are touched relationships whose parent ranking inputs the write changed (aliasRanksClaim). A
	// contains relationship among them places its target.
	reranked mapset.Set[uuid.UUID]
	// locked are the entities the transaction has locked FOR NO KEY UPDATE. Placement adds the ones it locks.
	locked mapset.Set[uuid.UUID]
}

func newSubjectRefresh() *subjectRefresh {
	return &subjectRefresh{
		entities:      mapset.NewSet[uuid.UUID](),
		relationships: mapset.NewSet[uuid.UUID](),
		placements:    mapset.NewSet[uuid.UUID](),
		reranked:      mapset.NewSet[uuid.UUID](),
		locked:        mapset.NewSet[uuid.UUID](),
	}
}

// rerank touches a relationship whose parent ranking inputs the write changed.
func (r *subjectRefresh) rerank(id uuid.UUID) {
	r.relationships.Add(id)
	r.reranked.Add(id)
}

const knowledgeStructureLockNamespace = "knowledge_structure"

// flush re-merges every touched subject from its aliases and places the parents that may have changed.
//
// It locks every row it may write before writing any, entities then relationships, each in ID order. A merge
// leaves one subject with two names, which the identity locks do not serialize, so the row locks protect the
// read-modify-write.
func (s *KnowledgeGraphIngestionService) flush(ctx context.Context, r *subjectRefresh) error {
	if r.entities.IsEmpty() && r.relationships.IsEmpty() {
		return nil
	}
	client := s.db.Client(ctx)

	// The lock set is the touched entities, the entities they contain, and the targets of the touched contains
	// relationships, which are every entity whose parent this flush can place.
	queryContains := client.KnowledgeRelationship.Query().
		Where(
			knr.PredicateEQ(knr.PredicateContains),
			knr.Or(knr.SourceEntityIDIn(r.entities.ToSlice()...), knr.IDIn(r.relationships.ToSlice()...)),
		)
	contains, containsErr := queryContains.All(ctx)
	if containsErr != nil {
		return fmt.Errorf("query contains relationships: %w", containsErr)
	}
	lockEntityIDs := r.entities.Clone()
	for _, relationship := range contains {
		lockEntityIDs.Add(relationship.TargetEntityID)
	}
	entities, lockErr := s.lockEntities(ctx, lockEntityIDs)
	if lockErr != nil {
		return lockErr
	}
	// A contained entity that a merge redirected since containment was read has nothing left to place.
	locked := make(map[uuid.UUID]*ent.KnowledgeEntity, len(entities))
	for _, entity := range entities {
		locked[entity.ID] = entity
		r.locked.Add(entity.ID)
	}
	// A relationship a merge folded is frozen as it was, like an entity it redirected (lockEntities), and the query
	// leaves it out.
	lockRelationships := client.KnowledgeRelationship.Query().
		Where(knr.IDIn(r.relationships.ToSlice()...)).
		Order(knr.ByID()).
		ForUpdate(ent.KeyUpdateLock)
	relationships, lockErr := lockRelationships.All(ctx)
	if lockErr != nil {
		return fmt.Errorf("lock relationships: %w", lockErr)
	}

	flipped := mapset.NewSet[uuid.UUID]()
	// Touched entities are held FOR KEY SHARE, or are a merge's survivor, so no merge should have redirected them.
	// One missing anyway is retried as a merged subject.
	for _, entityID := range sortedUUIDs(r.entities) {
		entity, isLocked := locked[entityID]
		if !isLocked {
			return fmt.Errorf("refresh entity %s: %w", entityID, errs.ErrKnowledgeSubjectMerged)
		}
		stored := entitySubjectState(entity)
		update := client.KnowledgeEntity.UpdateOneID(entityID)
		merged, written, remergeErr := s.remergeSubject(ctx, stored, ksa.EntityID(entityID), update.Mutation(), update.Exec)
		if remergeErr != nil {
			return fmt.Errorf("refresh entity: %w", remergeErr)
		}
		if written && (stored.WithdrawnAt == nil) != (merged.WithdrawnAt == nil) {
			flipped.Add(entityID)
		}
	}
	// A contains relationship that was created or whose withdrawal flipped adds or removes a claim on its target.
	// One whose ranking inputs changed can reorder the target's claims without either.
	isClaim := func(state knowledgegraph.SubjectState) bool {
		return state.LastObservedAt != nil && state.WithdrawnAt == nil
	}
	claimTargets := mapset.NewSet[uuid.UUID]()
	for _, relationship := range relationships {
		if relationship.Predicate == knr.PredicateContains && r.reranked.Contains(relationship.ID) {
			claimTargets.Add(relationship.TargetEntityID)
		}
		stored := relationshipSubjectState(relationship)
		update := client.KnowledgeRelationship.UpdateOneID(relationship.ID)
		merged, written, remergeErr := s.remergeSubject(ctx, stored, ksa.RelationshipID(relationship.ID), update.Mutation(), update.Exec)
		if remergeErr != nil {
			return fmt.Errorf("refresh relationship: %w", remergeErr)
		}
		if !written {
			continue
		}
		// A row never merged before has neither time set, so a new claim reads as not one before.
		if relationship.Predicate == knr.PredicateContains && isClaim(stored) != isClaim(merged) {
			claimTargets.Add(relationship.TargetEntityID)
		}
	}

	// This is the one place that decides when a parent can change. An entity becoming withdrawn or current again
	// changes which claims are candidates, for its own parent and for the entities it contains.
	placements := r.placements.Union(claimTargets).Union(flipped)
	if placements.IsEmpty() {
		return nil
	}
	return s.placeParents(ctx, r, placements, flipped)
}

// remergeSubject merges a subject's state from its aliases and, when it differs from stored, writes it through
// the subject's update. It reports whether it wrote.
func (s *KnowledgeGraphIngestionService) remergeSubject(ctx context.Context, stored knowledgegraph.SubjectState, aliasesOf predicate.KnowledgeSubjectAlias, update fieldMutation, exec func(context.Context) error) (knowledgegraph.SubjectState, bool, error) {
	aliases, aliasesErr := s.subjectAliases(ctx, aliasesOf)
	if aliasesErr != nil {
		return knowledgegraph.SubjectState{}, false, aliasesErr
	}
	merged := knowledgegraph.MergeSubjectState(aliases)
	if merged.Equal(stored) {
		return merged, false, nil
	}
	if mutationErr := setSubjectState(update, merged); mutationErr != nil {
		return knowledgegraph.SubjectState{}, false, mutationErr
	}
	if execErr := exec(ctx); execErr != nil {
		return knowledgegraph.SubjectState{}, false, fmt.Errorf("update state: %w", execErr)
	}
	return merged, true, nil
}

func entitySubjectState(entity *ent.KnowledgeEntity) knowledgegraph.SubjectState {
	return knowledgegraph.SubjectState{State: entity.State, Sources: entity.StateSources,
		StateEffectiveAt: entity.StateEffectiveAt, LastObservedAt: entity.LastObservedAt,
		WithdrawnAt: entity.WithdrawnAt, ExpiresAt: entity.ExpiresAt}
}

func relationshipSubjectState(relationship *ent.KnowledgeRelationship) knowledgegraph.SubjectState {
	return knowledgegraph.SubjectState{State: relationship.State, Sources: relationship.StateSources,
		StateEffectiveAt: relationship.StateEffectiveAt, LastObservedAt: relationship.LastObservedAt,
		WithdrawnAt: relationship.WithdrawnAt, ExpiresAt: relationship.ExpiresAt}
}

// lockEntities locks the entities FOR NO KEY UPDATE in one query ordered by ID, and returns those that a merge has
// not redirected, which the query leaves out. A redirected entity is frozen as the merge left it, so neither its
// state nor its parent is written again.
func (s *KnowledgeGraphIngestionService) lockEntities(ctx context.Context, ids mapset.Set[uuid.UUID]) ([]*ent.KnowledgeEntity, error) {
	lockEntities := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.IDIn(ids.ToSlice()...)).
		Order(kne.ByID()).
		ForUpdate(ent.KeyUpdateLock)
	entities, lockErr := lockEntities.All(ctx)
	if lockErr != nil {
		return nil, fmt.Errorf("lock entities: %w", lockErr)
	}
	return entities, nil
}

// placeParents places each entity in placements and the children of flipped entities. While any parent moves,
// it also re-places the tenant's fallback entities, since a move can free a claim skipped for a cycle.
//
// The tenant's structure lock serializes every is_parent write, so the cycle check sees a stable parent graph and
// no cycle is committed. It is taken after flush's row locks; a deadlock with another holder aborts one
// transaction, which retries.
func (s *KnowledgeGraphIngestionService) placeParents(ctx context.Context, r *subjectRefresh, placements, flipped mapset.Set[uuid.UUID]) error {
	if lockErr := s.db.AcquireTxLocks(ctx, knowledgeStructureLockNamespace, "tenant"); lockErr != nil {
		return fmt.Errorf("lock knowledge structure: %w", lockErr)
	}
	client := s.db.Client(ctx)
	lockUnlocked := func(ids mapset.Set[uuid.UUID]) error {
		entities, lockErr := s.lockEntities(ctx, ids.Difference(r.locked))
		if lockErr != nil {
			return lockErr
		}
		for _, entity := range entities {
			r.locked.Add(entity.ID)
		}
		return nil
	}
	toPlace := placements

	if !flipped.IsEmpty() {
		queryChildren := client.KnowledgeRelationship.Query().
			Where(knr.PredicateEQ(knr.PredicateContains), knr.SourceEntityIDIn(flipped.ToSlice()...))
		contains, childrenErr := queryChildren.All(ctx)
		if childrenErr != nil {
			return fmt.Errorf("query contained entities: %w", childrenErr)
		}
		children := mapset.NewSet[uuid.UUID]()
		for _, relationship := range contains {
			children.Add(relationship.TargetEntityID)
		}
		if lockErr := lockUnlocked(children); lockErr != nil {
			return lockErr
		}
		toPlace = toPlace.Union(children)
	}

	anyMoved := false
	for _, childID := range sortedUUIDs(toPlace) {
		if !r.locked.Contains(childID) {
			continue
		}
		moved, placeErr := s.placeParent(ctx, childID)
		if placeErr != nil {
			return fmt.Errorf("place parent: %w", placeErr)
		}
		anyMoved = anyMoved || moved
	}

	// A fallback that moves is itself a changed parent and can free another fallback's skipped claim, so the
	// passes repeat until one moves nothing. They end: every committed parent is a current claim that closes no
	// cycle, and moves keep the graph acyclic, so an entity's current claim never becomes blocked. Re-placing an
	// entity therefore only moves it to a better-ranked claim, and each entity has finitely many.
	for anyMoved {
		anyMoved = false
		queryFallbacks := client.KnowledgeEntity.Query().
			Where(kne.ParentIsFallback(true))
		fallbackIDs, fallbacksErr := queryFallbacks.IDs(ctx)
		if fallbacksErr != nil {
			return fmt.Errorf("query fallback placements: %w", fallbacksErr)
		}
		fallbacks := mapset.NewSet(fallbackIDs...)
		if lockErr := lockUnlocked(fallbacks); lockErr != nil {
			return lockErr
		}
		for _, childID := range sortedUUIDs(fallbacks) {
			if !r.locked.Contains(childID) {
				continue
			}
			moved, placeErr := s.placeParent(ctx, childID)
			if placeErr != nil {
				return fmt.Errorf("re-place fallback parent: %w", placeErr)
			}
			anyMoved = anyMoved || moved
		}
	}
	return nil
}

// placeParent flags the best-ranked current contains claim on the entity that closes no cycle, or none, and
// marks the entity as a fallback when it skipped a better claim for a cycle. It writes only what changed and
// reports whether the parent moved. The caller holds the structure lock and the entity's row lock.
func (s *KnowledgeGraphIngestionService) placeParent(ctx context.Context, childID uuid.UUID) (bool, error) {
	client := s.db.Client(ctx)
	child, childErr := client.KnowledgeEntity.Get(ctx, childID)
	if childErr != nil {
		return false, fmt.Errorf("load entity: %w", childErr)
	}
	queryFlagged := client.KnowledgeRelationship.Query().
		Where(knr.TargetEntityID(childID), knr.IsParent(true))
	flaggedIDs, flaggedErr := queryFlagged.IDs(ctx)
	if flaggedErr != nil {
		return false, fmt.Errorf("query parent claim: %w", flaggedErr)
	}
	// The unique index allows at most one.
	currentID := uuid.Nil
	if len(flaggedIDs) > 0 {
		currentID = flaggedIDs[0]
	}
	queryClaims := client.KnowledgeRelationship.Query().
		Where(
			knr.PredicateEQ(knr.PredicateContains),
			knr.TargetEntityID(childID),
			knr.SourceEntityIDNEQ(childID),
			currentKnowledgeRelationship(),
		).
		WithAliases(liveParentClaimAliases)
	claims, claimsErr := queryClaims.All(ctx)
	if claimsErr != nil {
		return false, fmt.Errorf("query parent claims: %w", claimsErr)
	}
	slices.SortFunc(claims, func(a, b *ent.KnowledgeRelationship) int {
		return knowledgegraph.RankParentClaims(parentClaimOf(a), parentClaimOf(b))
	})
	chosenID := uuid.Nil
	isFallback := false
	for _, claim := range claims {
		cycles, walkErr := s.hasAncestor(ctx, claim.SourceEntityID, childID)
		if walkErr != nil {
			return false, walkErr
		}
		if !cycles {
			chosenID = claim.ID
			break
		}
		isFallback = true
	}

	moved := chosenID != currentID
	if moved {
		// The flag is cleared before it is set, so the child never has two flagged claims for the unique index.
		clearFlag := client.KnowledgeRelationship.Update().
			Where(knr.TargetEntityID(childID), knr.IsParent(true)).
			SetIsParent(false)
		if clearErr := clearFlag.Exec(ctx); clearErr != nil {
			return false, fmt.Errorf("clear parent claim: %w", clearErr)
		}
		if chosenID != uuid.Nil {
			if setErr := client.KnowledgeRelationship.UpdateOneID(chosenID).SetIsParent(true).Exec(ctx); setErr != nil {
				return false, fmt.Errorf("set parent claim: %w", setErr)
			}
		}
	}
	if child.ParentIsFallback != isFallback {
		if updateErr := client.KnowledgeEntity.UpdateOneID(childID).SetParentIsFallback(isFallback).Exec(ctx); updateErr != nil {
			return false, fmt.Errorf("mark fallback placement: %w", updateErr)
		}
	}
	return moved, nil
}

// liveClaimAliases loads a claim's aliases that are not withdrawn: the providers claiming it, and what
// parentClaimOf ranks it by.
func liveClaimAliases(query *ent.KnowledgeSubjectAliasQuery) {
	query.Where(ksa.WithdrawnAtIsNil())
}

// liveParentClaimAliases loads only what parentClaimOf reads of a claim's live aliases.
func liveParentClaimAliases(query *ent.KnowledgeSubjectAliasQuery) {
	liveClaimAliases(query)
	query.Select(ksa.FieldRelationshipID, ksa.FieldAuthority, ksa.FieldFirstObservedAt)
}

// parentClaimOf is what knowledgegraph.RankParentClaims ranks of a contains relationship loaded with
// liveClaimAliases or liveParentClaimAliases.
func parentClaimOf(relationship *ent.KnowledgeRelationship) knowledgegraph.ParentClaim {
	claim := knowledgegraph.ParentClaim{RelationshipID: relationship.ID}
	for i, alias := range relationship.Edges.Aliases {
		if i == 0 || alias.Authority.Compare(claim.Authority) > 0 {
			claim.Authority = alias.Authority
		}
		if observed := alias.FirstObservedAt; observed != nil && (claim.FirstObservedAt.IsZero() || observed.Before(claim.FirstObservedAt)) {
			claim.FirstObservedAt = *observed
		}
	}
	return claim
}

// aliasRanksClaim reports whether an alias's change from before to after, with its authority raised or not,
// changes what parentClaimOf reads of the relationship it is on: the alias became live or withdrawn, or, while it
// is live, its authority rose or its first observation moved earlier, as when the alias is new. A withdrawn alias
// does not rank, so its authority and first observation do not matter until it is live again.
func aliasRanksClaim(before, after knowledgegraph.AliasRecord, raised bool) bool {
	live := after.WithdrawnAt == nil
	movedEarlier := after.FirstObservedAt != nil &&
		(before.FirstObservedAt == nil || after.FirstObservedAt.Before(*before.FirstObservedAt))
	return (before.WithdrawnAt == nil) != live || (live && (raised || movedEarlier))
}

// hasAncestor reports whether ancestorID is the entity or one of its ancestors, along the flagged parent claims,
// walking the whole chain in one query.
func (s *KnowledgeGraphIngestionService) hasAncestor(ctx context.Context, entityID, ancestorID uuid.UUID) (bool, error) {
	if entityID == ancestorID {
		return true, nil
	}
	ancestors, ancestryErr := ancestryQuery(ctx, s.db.Client(ctx), []uuid.UUID{entityID}, 0)
	if ancestryErr != nil {
		return false, ancestryErr
	}
	if cyclic := cyclicAncestry(ancestors); !cyclic.IsEmpty() {
		return false, fmt.Errorf("parent claims form a cycle above entities %v", cyclic.ToSlice())
	}
	return slices.ContainsFunc(ancestors, func(ancestor knowledgeAncestor) bool {
		return ancestor.AncestorID == ancestorID
	}), nil
}
