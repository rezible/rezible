package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

func (s *KnowledgeProposalService) ConfirmProposal(ctx context.Context, id uuid.UUID) (*rez.KnowledgeProposalWithValues, error) {
	userID, userErr := s.decidingUser(ctx)
	if userErr != nil {
		return nil, userErr
	}

	confirmErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		proposal, lockErr := s.lockProposal(ctx, id)
		if lockErr != nil {
			return lockErr
		}
		if proposal.Status == kp.StatusConfirmed {
			return errs.Wrap(errs.ErrConflict, errs.CodeConflict, "This proposal is already confirmed.")
		}

		entityA, entityB, valuesErr := s.resolveValues(ctx, proposal)
		if valuesErr != nil {
			return valuesErr
		}
		var merged *rez.MergeEntitiesResult
		if entityA.ID != entityB.ID {
			mergeParams := rez.MergeEntitiesParams{
				ProposalID: proposal.ID,
				EntityIDs:  [2]uuid.UUID{entityA.ID, entityB.ID},
			}
			result, mergeErr := s.ingestion.MergeEntities(ctx, mergeParams)
			if mergeErr != nil {
				return fmt.Errorf("merge entities: %w", mergeErr)
			}
			merged = result
		}

		if decideErr := s.decide(ctx, proposal.ID, kp.StatusConfirmed, userID); decideErr != nil {
			return decideErr
		}
		if merged == nil || !merged.Merged {
			return nil
		}
		return s.confirmImplied(ctx, proposal, merged.SurvivorID, userID)
	})
	if confirmErr != nil {
		// A merge can deadlock with a projection or another merge naming the same entities; Postgres aborts
		// one side, and nothing was changed.
		if s.db.IsTransientError(confirmErr) {
			return nil, errs.Wrap(fmt.Errorf("%w: %w", errs.ErrConflict, confirmErr), errs.CodeConflict,
				"Something else changed these services at the same time. Try again.")
		}
		return nil, confirmErr
	}
	return s.GetProposalWithValues(ctx, id)
}

func (s *KnowledgeProposalService) RejectProposal(ctx context.Context, id uuid.UUID) (*rez.KnowledgeProposalWithValues, error) {
	userID, userErr := s.decidingUser(ctx)
	if userErr != nil {
		return nil, userErr
	}

	rejectErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		proposal, lockErr := s.lockProposal(ctx, id)
		if lockErr != nil {
			return lockErr
		}
		if proposal.Status != kp.StatusPending {
			return errs.Wrap(errs.ErrConflict, errs.CodeConflict,
				fmt.Sprintf("Only a pending proposal can be rejected; this one is %s.", proposal.Status))
		}
		return s.decide(ctx, proposal.ID, kp.StatusRejected, userID)
	})
	if rejectErr != nil {
		return nil, rejectErr
	}
	return s.GetProposalWithValues(ctx, id)
}

func (s *KnowledgeProposalService) ListImpliedProposals(ctx context.Context, proposal *ent.KnowledgeProposal) ([]*ent.KnowledgeProposal, error) {
	if proposal.Status == kp.StatusConfirmed {
		return nil, nil
	}
	holders, holdersErr := s.resolveProposalValues(ctx, []*ent.KnowledgeProposal{proposal})
	if holdersErr != nil {
		return nil, holdersErr
	}
	entityA, entityB := holders.Of(proposal)
	if entityA == nil || entityB == nil || entityA.ID == entityB.ID {
		return nil, nil
	}
	return s.queryImplied(ctx, proposal, entityA.ID, entityB.ID)
}

// decidingUser returns the person deciding on a proposal. Only people decide: system and agent identities are
// refused (K0 decision 7).
func (s *KnowledgeProposalService) decidingUser(ctx context.Context) (uuid.UUID, error) {
	exec := execution.GetContext(ctx)
	userID, hasUser := exec.UserID()
	if !exec.IsUser() || !hasUser || userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: only a person can decide on a proposal", errs.ErrForbidden)
	}
	return userID, nil
}

// lockProposal locks the proposal's row, so two decisions on it are made one after the other and each sees the
// status the other left.
func (s *KnowledgeProposalService) lockProposal(ctx context.Context, id uuid.UUID) (*ent.KnowledgeProposal, error) {
	lockProposal := s.db.Client(ctx).KnowledgeProposal.Query().
		Where(kp.ID(id)).
		ForUpdate()
	proposal, lockErr := lockProposal.Only(ctx)
	if lockErr != nil {
		return nil, fmt.Errorf("lock proposal: %w", lockErr)
	}
	return proposal, nil
}

// resolveValues returns the entities holding the proposal's two values, which may be one entity.
func (s *KnowledgeProposalService) resolveValues(ctx context.Context, proposal *ent.KnowledgeProposal) (*ent.KnowledgeEntity, *ent.KnowledgeEntity, error) {
	holders, holdersErr := s.resolveProposalValues(ctx, []*ent.KnowledgeProposal{proposal})
	if holdersErr != nil {
		return nil, nil, holdersErr
	}
	entityA, entityB := holders.Of(proposal)
	if entityA == nil || entityB == nil {
		return nil, nil, errs.Wrap(errs.ErrConflict, errs.CodeConflict, fmt.Sprintf(
			"This proposal is %s, but one of its names no longer belongs to anything Rezible knows.", proposal.Status))
	}
	return entityA, entityB, nil
}

// decide records a person's decision and sets the proposal's status to it.
func (s *KnowledgeProposalService) decide(ctx context.Context, proposalID uuid.UUID, status kp.Status, userID uuid.UUID) error {
	client := s.db.Client(ctx)
	recordDecision := client.KnowledgeProposalDecision.Create().
		SetProposalID(proposalID).
		SetStatus(kpd.Status(status)).
		SetUserID(userID)
	if recordErr := recordDecision.Exec(ctx); recordErr != nil {
		return fmt.Errorf("record decision: %w", recordErr)
	}
	setStatus := client.KnowledgeProposal.UpdateOneID(proposalID).
		SetStatus(status)
	if statusErr := setStatus.Exec(ctx); statusErr != nil {
		return fmt.Errorf("set proposal status: %w", statusErr)
	}
	return nil
}

// confirmImplied confirms the proposals whose values the merge put on the survivor, each as a decision by the
// person who confirmed the merge, implied by the merged proposal. Each is confirmed with an update conditioned on
// its status, so one confirmed meanwhile is skipped, and a rejection committed meanwhile is either seen by the
// query or deliberately confirmed.
func (s *KnowledgeProposalService) confirmImplied(ctx context.Context, merged *ent.KnowledgeProposal, survivorID uuid.UUID, userID uuid.UUID) error {
	implied, impliedErr := s.queryImplied(ctx, merged, survivorID)
	if impliedErr != nil {
		return impliedErr
	}
	client := s.db.Client(ctx)
	for _, proposal := range implied {
		confirm := client.KnowledgeProposal.UpdateOneID(proposal.ID).
			Where(kp.StatusNEQ(kp.StatusConfirmed)).
			SetStatus(kp.StatusConfirmed)
		if confirmErr := confirm.Exec(ctx); confirmErr != nil {
			if ent.IsNotFound(confirmErr) {
				continue
			}
			return fmt.Errorf("confirm implied proposal: %w", confirmErr)
		}
		recordDecision := client.KnowledgeProposalDecision.Create().
			SetProposalID(proposal.ID).
			SetStatus(kpd.StatusConfirmed).
			SetUserID(userID).
			SetImpliedByProposalID(merged.ID)
		if recordErr := recordDecision.Exec(ctx); recordErr != nil {
			return fmt.Errorf("record implied decision: %w", recordErr)
		}
	}
	return nil
}

// queryImplied lists the proposals with the proposal's attribute, other than it and not confirmed, whose two
// values are both linking values of the entities, with their decisions.
func (s *KnowledgeProposalService) queryImplied(ctx context.Context, proposal *ent.KnowledgeProposal, entityIDs ...uuid.UUID) ([]*ent.KnowledgeProposal, error) {
	client := s.db.Client(ctx)
	queryValues := client.KnowledgeEntityLinkingAttribute.Query().
		Where(kela.EntityIDIn(entityIDs...), kela.Attribute(proposal.Attribute)).
		Select(kela.FieldValue)
	values, valuesErr := queryValues.Strings(ctx)
	if valuesErr != nil {
		return nil, fmt.Errorf("query entity values: %w", valuesErr)
	}
	queryImplied := s.withProposalDetails(client.KnowledgeProposal.Query()).
		Where(
			kp.Attribute(proposal.Attribute),
			kp.IDNEQ(proposal.ID),
			kp.StatusNEQ(kp.StatusConfirmed),
			kp.ValueAIn(values...),
			kp.ValueBIn(values...),
		).
		Order(kp.ByCreatedAt(), kp.ByID())
	implied, impliedErr := queryImplied.All(ctx)
	if impliedErr != nil {
		return nil, fmt.Errorf("query implied proposals: %w", impliedErr)
	}
	return implied, nil
}
