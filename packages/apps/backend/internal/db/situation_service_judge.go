package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/situations"
)

// judgeByModel asks the model judge to decide the captured candidate, outside any transaction, within
// JudgeCallTimeout of real time and on the facts cut to its limits. A malformed or invalid answer becomes a
// rejected hold, and a failed or timed-out call an unavailable hold; only cancellation of the evaluation
// itself is an error. Without a configured model the hold is final for these facts. The model's outage never
// falls back to the rules.
func (s *SituationService) judgeByModel(ctx context.Context, snapshot *candidateSnapshot) (situations.Judgment, error) {
	facts, truncated := situations.JudgeFacts(snapshot.facts)
	input := rezai.SituationJudgeInput{Facts: facts, Reasons: snapshot.reasons, Truncated: truncated}
	callCtx, cancelCall := context.WithTimeout(ctx, situations.JudgeCallTimeout)
	answer, runErr := s.judge.Run(callCtx, input)
	cancelCall()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return situations.Judgment{}, fmt.Errorf("judge situation candidate: %w", ctxErr)
	}
	switch {
	case errors.Is(runErr, rezai.ErrNoModel):
		slog.WarnContext(ctx, "situation judge has no model", "situationId", snapshot.situationID)
		return situations.NoModelJudgment(), nil
	case errors.Is(runErr, rezai.ErrWorkflowInvalidOutput):
		slog.WarnContext(ctx, "situation judge answer malformed", "situationId", snapshot.situationID, "error", runErr)
		return situations.RejectedJudgment(situations.ErrMalformedAnswer), nil
	case runErr != nil:
		slog.WarnContext(ctx, "situation judge unavailable", "situationId", snapshot.situationID, "error", runErr)
		return situations.UnavailableJudgment(), nil
	}
	modelAnswer := situations.Judgment{
		Decision:    sitjudg.Decision(answer.Decision),
		Cited:       answer.Reasons,
		Explanation: answer.Explanation,
	}
	return snapshot.reasons.JudgeByModel(modelAnswer), nil
}

// applyModelJudgment records the model judge's judgment under the row lock, at a new processing time, only
// if the candidate is still an unmuted candidate that is not due to close and its facts did not change while
// the model ran. Otherwise it discards the judgment and requests an evaluation now. When another evaluation
// already judged the same facts, there is nothing left to do.
func (s *SituationService) applyModelJudgment(ctx context.Context, captured *candidateSnapshot, judgment situations.Judgment) error {
	id := captured.situationID
	locked, lockErr := s.lockSituations(ctx, id)
	if lockErr != nil {
		return lockErr
	}
	situation := locked[id]
	if situation.ClosedAt != nil || situation.RaisedAt != nil || situation.MutedAt != nil {
		return s.requestEvaluation(ctx, id)
	}
	now := s.clock.Now()
	input, _, inputErr := s.loadEvaluationInput(ctx, situation, now)
	if inputErr != nil {
		return inputErr
	}
	if closeableAt := input.CloseableAt(); closeableAt != nil && !now.Before(*closeableAt) {
		return s.requestEvaluation(ctx, id)
	}
	current, snapshotErr := s.snapshotCandidate(ctx, situation, input)
	if snapshotErr != nil {
		return snapshotErr
	}
	if current.fingerprint != captured.fingerprint {
		slog.InfoContext(ctx, "situation judgment discarded: facts changed while judging", "situationId", id)
		return s.requestEvaluation(ctx, id)
	}
	if input.JudgmentStands(current.fingerprint) {
		return nil
	}
	if recordErr := s.recordJudgment(ctx, situation, captured, judgment, now); recordErr != nil {
		return recordErr
	}
	input.Record(judgment, captured.fingerprint)
	return s.scheduleEvaluation(ctx, id, input.NextDeadline())
}

// loadJudgeContext loads what only the model judge reads: the situation's entities with their names and
// properties, each member observation's stored runtime entities, the depends-on and adjacent relationships
// among the entities.
func (s *SituationService) loadJudgeContext(ctx context.Context, situationID uuid.UUID, input situations.EvaluationInput) (*situations.JudgeContext, error) {
	judgeContext := &situations.JudgeContext{
		ObservationEntityIDs: make(map[rez.SituationObservationRef][]uuid.UUID, len(input.Members)),
	}
	queryMembers := s.db.Client(ctx).SituationObservation.Query().
		Where(sitobs.SituationID(situationID)).
		WithScopes()
	members, membersErr := queryMembers.All(ctx)
	if membersErr != nil {
		return nil, fmt.Errorf("query situation observations: %w", membersErr)
	}
	entities := situations.RuntimeEntities(members)
	entityIDs := make([]uuid.UUID, 0, len(entities))
	for _, runtimeEntity := range entities {
		entity := runtimeEntity.Entity
		entityFacts := schematypes.SituationEntityFacts{
			ID:          entity.ID,
			Category:    string(entity.Category),
			Kind:        entity.Kind,
			DisplayName: entity.State.DisplayName,
			Properties:  entity.State.Properties,
			Matching:    runtimeEntity.Matching,
		}
		judgeContext.Entities = append(judgeContext.Entities, entityFacts)
		entityIDs = append(entityIDs, entity.ID)
	}

	for _, member := range members {
		scope := member.Scope()
		if scope == nil {
			continue
		}
		observationEntityIDs := make([]uuid.UUID, len(scope.Edges.Entities))
		for i, row := range scope.Edges.Entities {
			observationEntityIDs[i] = row.KnowledgeEntityID
		}
		judgeContext.ObservationEntityIDs[member.Ref()] = observationEntityIDs
	}

	if len(entityIDs) > 0 {
		classPredicates := knowledgegraph.PredicatesWithDependency(knowledgegraph.DependencySourceOnTarget, knowledgegraph.DependencyMutual)
		listParams := rez.ListKnowledgeRelationshipsParams{
			PageSize: observationGraphPageSize,
			Predicates: []predicate.KnowledgeRelationship{
				knr.PredicateIn(classPredicates...),
				knr.SourceEntityIDIn(entityIDs...),
				knr.TargetEntityIDIn(entityIDs...),
			},
		}
		relationships, relationshipsErr := s.graph.ListRelationships(ctx, listParams)
		if relationshipsErr != nil {
			return nil, fmt.Errorf("list situation entity relationships: %w", relationshipsErr)
		}
		for _, relationship := range relationships.Data {
			relationshipFacts := schematypes.SituationRelationshipFacts{
				SourceID:  relationship.SourceEntityID,
				Predicate: string(relationship.Predicate),
				TargetID:  relationship.TargetEntityID,
			}
			judgeContext.Relationships = append(judgeContext.Relationships, relationshipFacts)
		}
	}

	return judgeContext, nil
}
