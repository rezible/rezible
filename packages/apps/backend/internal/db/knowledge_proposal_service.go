package db

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeProposalService struct {
	db rez.Database
}

func NewKnowledgeProposalService(db rez.Database) (*KnowledgeProposalService, error) {
	return &KnowledgeProposalService{db: db}, nil
}

func (s *KnowledgeProposalService) ListProposals(ctx context.Context, params rez.ListProposalsParams) (*ent.ListResult[ent.KnowledgeProposal], error) {
	query := s.db.Client(ctx).KnowledgeProposal.Query().
		WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.WithSubjectAlias()
		}).
		Order(kp.ByCreatedAt(params.GetOrder()), kp.ByID(params.GetOrder()))

	if len(params.Statuses) > 0 {
		query.Where(kp.StatusIn(params.Statuses...))
	}
	if len(params.EntityIDs) > 0 {
		namedPred, namedErr := s.namingEntities(ctx, params)
		if namedErr != nil {
			return nil, namedErr
		}
		query.Where(namedPred)
	}

	return ent.DoListQuery[ent.KnowledgeProposal, *ent.KnowledgeProposalQuery](ctx, query, params.ListParams)
}

// unconfirmedMatchesTotalCapPerEntity sets the total cap on pending proposals read across all the entities
// asked about: ten times their number, shared, so one entity may have more than ten listed and another fewer.
const unconfirmedMatchesTotalCapPerEntity = 10

func (s *KnowledgeProposalService) ListUnconfirmedMatches(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]rez.UnconfirmedMatch, error) {
	matches := make(map[uuid.UUID][]rez.UnconfirmedMatch)
	if len(entityIDs) == 0 {
		return matches, nil
	}
	listParams := rez.ListProposalsParams{
		ListParams: ent.ListParams{PageSize: unconfirmedMatchesTotalCapPerEntity * len(entityIDs), OrderAsc: true},
		Statuses:   []kp.Status{kp.StatusPending},
		EntityIDs:  entityIDs,
	}
	proposals, listErr := s.ListProposals(ctx, listParams)
	if listErr != nil {
		return nil, fmt.Errorf("list pending proposals: %w", listErr)
	}
	if len(proposals.Data) == 0 {
		return matches, nil
	}

	// Each proposal names two values; find the entity holding each.
	var valuePreds []predicate.KnowledgeEntityLinkingAttribute
	for _, proposal := range proposals.Data {
		valuePreds = append(valuePreds, kela.And(
			kela.Attribute(proposal.Attribute),
			kela.ValueIn(proposal.NameA, proposal.NameB),
		))
	}
	queryHolders := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(kela.Or(valuePreds...)).
		WithEntity()
	holders, holdersErr := queryHolders.All(ctx)
	if holdersErr != nil {
		return nil, fmt.Errorf("query proposal value holders: %w", holdersErr)
	}
	holderOf := make(map[[2]string]*ent.KnowledgeEntity, len(holders))
	for _, holder := range holders {
		holderOf[[2]string{holder.Attribute, holder.Value}] = holder.Edges.Entity
	}

	requested := mapset.NewSet(entityIDs...)
	for _, proposal := range proposals.Data {
		entityA := holderOf[[2]string{proposal.Attribute, proposal.NameA}]
		entityB := holderOf[[2]string{proposal.Attribute, proposal.NameB}]
		if entityA == nil || entityB == nil || entityA.ID == entityB.ID {
			continue
		}
		if requested.Contains(entityA.ID) {
			matches[entityA.ID] = append(matches[entityA.ID], rez.UnconfirmedMatch{
				ProposalID: proposal.ID,
				Rationale:  proposal.Rationale,
				Name:       proposal.NameB,
				Entity:     entityB,
			})
		}
		if requested.Contains(entityB.ID) {
			matches[entityB.ID] = append(matches[entityB.ID], rez.UnconfirmedMatch{
				ProposalID: proposal.ID,
				Rationale:  proposal.Rationale,
				Name:       proposal.NameA,
				Entity:     entityA,
			})
		}
	}
	return matches, nil
}

// namingEntities matches proposals naming a linking value the entities hold now. Proposals name values, not
// entities, so a proposal follows its values when they move to another entity.
func (s *KnowledgeProposalService) namingEntities(ctx context.Context, params rez.ListProposalsParams) (predicate.KnowledgeProposal, error) {
	valuesQuery := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(kela.EntityIDIn(params.EntityIDs...))
	values, queryErr := valuesQuery.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query entity linking attributes: %w", queryErr)
	}

	valuesByAttribute := make(map[string][]string)
	for _, value := range values {
		valuesByAttribute[value.Attribute] = append(valuesByAttribute[value.Attribute], value.Value)
	}
	preds := make([]predicate.KnowledgeProposal, 0, len(valuesByAttribute))
	for attribute, attributeValues := range valuesByAttribute {
		preds = append(preds, kp.And(
			kp.Attribute(attribute),
			kp.Or(kp.NameAIn(attributeValues...), kp.NameBIn(attributeValues...)),
		))
	}
	if len(preds) == 0 {
		// The entities hold no linking values, so no proposal names them.
		return kp.IDIn(), nil
	}
	return kp.Or(preds...), nil
}
