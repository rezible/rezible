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
	sitlink "github.com/rezible/rezible/ent/situationlink"
	sitog "github.com/rezible/rezible/ent/situationobservationgroup"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/situations"
)

// signalGraphPageSize is large enough to read a signal's observed entities or relationships in one page.
const signalGraphPageSize = 10_000

// AttachSituationSignalsParams attaches signals to a situation under the observation group with GroupTitle.
type AttachSituationSignalsParams struct {
	SituationID       uuid.UUID
	GroupTitle        string
	SignalEntityIDs   []uuid.UUID
	MatchKind         sitsig.MatchKind
	ViaRelationshipID *uuid.UUID
	Explanation       string
}

// situationMembership describes membership rows to create in one group.
type situationMembership struct {
	groupID           uuid.UUID
	matchKind         sitsig.MatchKind
	viaRelationshipID *uuid.UUID
	explanation       string
}

// AttachSituationSignals attaches the signals the situation does not already hold. A signal held by another
// situation, closed or not, is a conflict.
func (s *SituationService) AttachSituationSignals(ctx context.Context, params AttachSituationSignalsParams) error {
	title := strings.TrimSpace(params.GroupTitle)
	if title == "" {
		return fmt.Errorf("%w: observation group title is required", errs.ErrInvalidInput)
	}
	if matchErr := sitsig.MatchKindValidator(params.MatchKind); matchErr != nil {
		return fmt.Errorf("%w: %w", errs.ErrInvalidInput, matchErr)
	}
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
		memberIDs, membersErr := s.memberEntityIDs(ctx, situation.ID)
		if membersErr != nil {
			return membersErr
		}
		newIDs := mapset.NewSet(params.SignalEntityIDs...).Difference(mapset.NewSet(memberIDs...))
		newIDs.Remove(uuid.Nil)
		if newIDs.IsEmpty() {
			return nil
		}
		signals, loadErr := s.loadSignals(ctx, newIDs.ToSlice(), situations.LoadSignalsOptions{})
		if loadErr != nil {
			return loadErr
		}

		groupID, groupErr := s.observationGroupByTitle(ctx, situation.ID, title)
		if groupErr != nil {
			return groupErr
		}
		membership := situationMembership{
			groupID:           groupID,
			matchKind:         params.MatchKind,
			viaRelationshipID: params.ViaRelationshipID,
			explanation:       strings.TrimSpace(params.Explanation),
		}
		newSignals := make([]situations.Signal, 0, len(signals))
		for _, signal := range signals {
			newSignals = append(newSignals, signal)
		}
		if memberErr := s.createMemberships(ctx, situation.ID, membership, newSignals, now); memberErr != nil {
			return memberErr
		}
		if lowerErr := s.lowerOpenedAt(ctx, situation, s.earliestStart(signals)); lowerErr != nil {
			return lowerErr
		}
		if entitiesErr := s.recomputeSituationEntities(ctx, situation.ID); entitiesErr != nil {
			return entitiesErr
		}
		return s.requestEvaluation(ctx, situation.ID)
	})
}

// observationGroupByTitle returns the situation's group with exactly this title, creating it if needed.
func (s *SituationService) observationGroupByTitle(ctx context.Context, situationID uuid.UUID, title string) (uuid.UUID, error) {
	client := s.db.Client(ctx)
	queryGroup := client.SituationObservationGroup.Query().
		Where(sitog.SituationID(situationID), sitog.Title(title)).
		Order(sitog.ByCreatedAt(), sitog.ByID())
	group, queryErr := queryGroup.First(ctx)
	if queryErr == nil {
		return group.ID, nil
	}
	if !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("find observation group: %w", queryErr)
	}
	createGroup := client.SituationObservationGroup.Create().
		SetSituationID(situationID).
		SetTitle(title)
	created, createErr := createGroup.Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create observation group: %w", createErr)
	}
	return created.ID, nil
}

// createMemberships creates one membership per signal in the group. A signal belongs to at most one
// situation; the unique index turns a second placement into a conflict.
func (s *SituationService) createMemberships(ctx context.Context, situationID uuid.UUID, membership situationMembership, signals []situations.Signal, now time.Time) error {
	createMembers := s.db.Client(ctx).SituationSignal.MapCreateBulk(signals, func(c *ent.SituationSignalCreate, i int) {
		c.SetSituationID(situationID).
			SetObservationGroupID(membership.groupID).
			SetKnowledgeEntityID(signals[i].EntityID).
			SetKind(string(signals[i].Kind)).
			SetNillableSourceEntityID(signals[i].SourceEntityID).
			SetAttachedAt(now).
			SetMatchKind(membership.matchKind).
			SetNillableViaRelationshipID(membership.viaRelationshipID).
			SetMatchExplanation(membership.explanation)
	})
	if createErr := createMembers.Exec(ctx); createErr != nil {
		if ent.IsConstraintError(createErr) {
			return fmt.Errorf("%w: a signal already belongs to a situation", errs.ErrConflict)
		}
		return fmt.Errorf("create situation signals: %w", createErr)
	}
	return nil
}

// lowerOpenedAt moves the situation's opening back to an earlier signal start; it never moves it forward.
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

// recomputeSituationEntities replaces the situation's entity rows with the union of its signals' runtime
// entities, resolving every signal again. An entity is matching when a signal that is not broad touches it.
func (s *SituationService) recomputeSituationEntities(ctx context.Context, situationID uuid.UUID) error {
	memberIDs, membersErr := s.memberEntityIDs(ctx, situationID)
	if membersErr != nil {
		return membersErr
	}
	matching := make(map[uuid.UUID]bool)
	if len(memberIDs) > 0 {
		signals, loadErr := s.loadSignals(ctx, memberIDs, situations.LoadSignalsOptions{})
		if loadErr != nil {
			return loadErr
		}
		for _, signal := range signals {
			entityIDs, entitiesErr := s.signalRuntimeEntities(ctx, signal)
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

// signalRuntimeEntities returns the runtime entities a signal is about: the entities its events observed,
// and the endpoints of relationships they observed, resolved to the runtime structure level. Evidence that
// an entity was deleted does not make a signal about it.
func (s *SituationService) signalRuntimeEntities(ctx context.Context, signal situations.Signal) (mapset.Set[uuid.UUID], error) {
	runtime := mapset.NewSet[uuid.UUID]()
	if len(signal.EventIDs) == 0 {
		return runtime, nil
	}
	observed := ksa.HasEvidenceWith(knev.EventIDIn(signal.EventIDs...), knev.KindEQ(knev.KindObserved))
	pageParams := ent.ListParams{PageSize: signalGraphPageSize}

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
	resolveParams := rez.ResolveStructureParams{
		EntityIDs:        named.ToSlice(),
		TargetCategories: knowledgegraph.StructureLevelRuntime.Categories(),
		MaxDepth:         situations.EntityResolutionDepth,
	}
	resolved, resolveErr := s.graph.ResolveStructure(ctx, resolveParams)
	if resolveErr != nil {
		return nil, fmt.Errorf("resolve signal entities: %w", resolveErr)
	}
	for _, representatives := range resolved {
		runtime.Append(representatives...)
	}
	return runtime, nil
}

// MergeSituations moves the source's signals into the target and closes the source as merged.
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
		if moveErr := s.moveSignals(ctx, source.ID, target.ID, explanation, now); moveErr != nil {
			return nil, moveErr
		}
		if closeErr := s.closeLocked(ctx, source, sit.CloseReasonMerged, explanation, now, now); closeErr != nil {
			return nil, closeErr
		}
		createLink := tx.SituationLink.Create().
			SetSituationID(source.ID).
			SetLinkedSituationID(target.ID).
			SetKind(sitlink.KindMergedInto).
			SetCreatedAt(now)
		if linkErr := createLink.Exec(ctx); linkErr != nil {
			return nil, fmt.Errorf("link merged situation: %w", linkErr)
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
		return s.GetSituation(ctx, target.ID)
	})
}

// moveSignals moves the source's memberships to the target as manual attachments made now; the reset
// attachment time counts the merge as new evidence for the target, and the reset observed revision brings
// the signals into the target's investigation. A source group joins the target's group of the same title;
// other groups move with their signals.
func (s *SituationService) moveSignals(ctx context.Context, sourceID, targetID uuid.UUID, explanation string, now time.Time) error {
	client := s.db.Client(ctx)
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
			destinationID = group.ID
			moveGroup := client.SituationObservationGroup.UpdateOneID(group.ID).
				SetSituationID(targetID)
			if moveErr := moveGroup.Exec(ctx); moveErr != nil {
				return fmt.Errorf("move observation group: %w", moveErr)
			}
		}
		moveMembers := client.SituationSignal.Update().
			Where(sitsig.ObservationGroupID(group.ID)).
			SetSituationID(targetID).
			SetObservationGroupID(destinationID).
			SetMatchKind(sitsig.MatchKindManual).
			SetMatchExplanation(explanation).
			SetAttachedAt(now).
			SetObservedRevision(0)
		if moveErr := moveMembers.Exec(ctx); moveErr != nil {
			return fmt.Errorf("move situation signals: %w", moveErr)
		}
		if joins {
			if deleteErr := client.SituationObservationGroup.DeleteOneID(group.ID).Exec(ctx); deleteErr != nil {
				return fmt.Errorf("delete joined observation group: %w", deleteErr)
			}
		}
	}
	return nil
}
