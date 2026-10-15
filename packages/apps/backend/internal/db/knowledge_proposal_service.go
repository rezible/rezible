package db

import (
	"context"
	"fmt"

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
