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
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/integrations"
)

type KnowledgeProposalService struct {
	db        rez.Database
	ingestion rez.KnowledgeGraphIngestionService
	aliases   rez.KnowledgeEntityAliasService
	names     integrationNames
}

func NewKnowledgeProposalService(db rez.Database, ingestion rez.KnowledgeGraphIngestionService, aliases rez.KnowledgeEntityAliasService, registry *integrations.Registry) (*KnowledgeProposalService, error) {
	return &KnowledgeProposalService{db: db, ingestion: ingestion, aliases: aliases, names: integrationNames{registry: registry}}, nil
}

// withProposalDetails loads a proposal's cited evidence with its alias, and its decisions with their users.
func (s *KnowledgeProposalService) withProposalDetails(query *ent.KnowledgeProposalQuery) *ent.KnowledgeProposalQuery {
	return query.
		WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.WithSubjectAlias()
		}).
		WithDecisions(func(dq *ent.KnowledgeProposalDecisionQuery) {
			dq.WithUser().Order(kpd.ByCreatedAt(), kpd.ByID())
		})
}

func (s *KnowledgeProposalService) ListProposals(ctx context.Context, params rez.ListProposalsParams) (*ent.ListResult[ent.KnowledgeProposal], error) {
	query := s.withProposalDetails(s.db.Client(ctx).KnowledgeProposal.Query()).
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

func (s *KnowledgeProposalService) ListProposalsWithValues(ctx context.Context, params rez.ListProposalsParams) (*ent.ListResult[rez.KnowledgeProposalWithValues], error) {
	proposals, listErr := s.ListProposals(ctx, params)
	if listErr != nil {
		return nil, listErr
	}
	withValues, valuesErr := s.withValues(ctx, proposals.Data)
	if valuesErr != nil {
		return nil, valuesErr
	}
	return &ent.ListResult[rez.KnowledgeProposalWithValues]{
		Data:     withValues,
		Page:     proposals.Page,
		PageSize: proposals.PageSize,
		Total:    proposals.Total,
	}, nil
}

func (s *KnowledgeProposalService) getProposal(ctx context.Context, id uuid.UUID) (*ent.KnowledgeProposal, error) {
	query := s.withProposalDetails(s.db.Client(ctx).KnowledgeProposal.Query()).
		Where(kp.ID(id))
	return query.Only(ctx)
}

func (s *KnowledgeProposalService) GetProposalWithValues(ctx context.Context, id uuid.UUID) (*rez.KnowledgeProposalWithValues, error) {
	proposal, getErr := s.getProposal(ctx, id)
	if getErr != nil {
		return nil, getErr
	}
	withValues, valuesErr := s.withValues(ctx, []*ent.KnowledgeProposal{proposal})
	if valuesErr != nil {
		return nil, valuesErr
	}
	return withValues[0], nil
}

// withValues reads the entities holding the proposals' values now, with their aliases, in one query for the
// holders and one for the aliases.
func (s *KnowledgeProposalService) withValues(ctx context.Context, proposals []*ent.KnowledgeProposal) ([]*rez.KnowledgeProposalWithValues, error) {
	holders, holdersErr := s.resolveProposalValues(ctx, proposals)
	if holdersErr != nil {
		return nil, fmt.Errorf("resolve proposal values: %w", holdersErr)
	}
	entityIDs := mapset.NewSet[uuid.UUID]()
	for _, holder := range holders {
		entityIDs.Add(holder.ID)
	}
	aliases, aliasesErr := s.aliases.ListEntityAliases(ctx, entityIDs.ToSlice())
	if aliasesErr != nil {
		return nil, fmt.Errorf("list entity aliases: %w", aliasesErr)
	}

	valueOf := func(value string, holder *ent.KnowledgeEntity) rez.KnowledgeProposalValue {
		if holder == nil {
			return rez.KnowledgeProposalValue{Value: value}
		}
		withAliases := &rez.KnowledgeEntityWithAliases{Entity: holder, Aliases: aliases[holder.ID]}
		return rez.KnowledgeProposalValue{Value: value, Holder: withAliases}
	}
	withValues := make([]*rez.KnowledgeProposalWithValues, len(proposals))
	for i, proposal := range proposals {
		holderA, holderB := holders.Of(proposal)
		withValues[i] = &rez.KnowledgeProposalWithValues{
			Proposal: proposal,
			ValueA:   valueOf(proposal.ValueA, holderA),
			ValueB:   valueOf(proposal.ValueB, holderB),
		}
		if evidence := proposal.Edges.Evidence; evidence != nil && evidence.Edges.SubjectAlias != nil {
			withValues[i].EvidenceIntegrationDisplayName = s.names.displayName(evidence.Edges.SubjectAlias.Provider)
		}
	}
	return withValues, nil
}

// resolveProposalValues finds the entity holding each of the proposals' values now.
func (s *KnowledgeProposalService) resolveProposalValues(ctx context.Context, proposals []*ent.KnowledgeProposal) (rez.KnowledgeValueHolders, error) {
	holders := make(rez.KnowledgeValueHolders, 2*len(proposals))
	if len(proposals) == 0 {
		return holders, nil
	}
	valuePreds := make([]predicate.KnowledgeEntityLinkingAttribute, 0, len(proposals))
	for _, proposal := range proposals {
		valuePreds = append(valuePreds, kela.And(
			kela.Attribute(proposal.Attribute),
			kela.ValueIn(proposal.ValueA, proposal.ValueB),
		))
	}
	queryHolders := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(kela.Or(valuePreds...)).
		WithEntity()
	values, holdersErr := queryHolders.All(ctx)
	if holdersErr != nil {
		return nil, fmt.Errorf("query proposal value holders: %w", holdersErr)
	}
	for _, value := range values {
		holders[rez.KnowledgeLinkingValue{Attribute: value.Attribute, Value: value.Value}] = value.Edges.Entity
	}
	return holders, nil
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

	holders, holdersErr := s.resolveProposalValues(ctx, proposals.Data)
	if holdersErr != nil {
		return nil, holdersErr
	}

	requested := mapset.NewSet(entityIDs...)
	for _, proposal := range proposals.Data {
		entityA, entityB := holders.Of(proposal)
		if entityA == nil || entityB == nil || entityA.ID == entityB.ID {
			continue
		}
		if requested.Contains(entityA.ID) {
			matches[entityA.ID] = append(matches[entityA.ID], rez.UnconfirmedMatch{
				ProposalID: proposal.ID,
				Rationale:  proposal.Rationale,
				Value:      proposal.ValueB,
				Entity:     entityB,
			})
		}
		if requested.Contains(entityB.ID) {
			matches[entityB.ID] = append(matches[entityB.ID], rez.UnconfirmedMatch{
				ProposalID: proposal.ID,
				Rationale:  proposal.Rationale,
				Value:      proposal.ValueA,
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
			kp.Or(kp.ValueAIn(attributeValues...), kp.ValueBIn(attributeValues...)),
		))
	}
	if len(preds) == 0 {
		// The entities hold no linking values, so no proposal names them.
		return kp.IDIn(), nil
	}
	return kp.Or(preds...), nil
}
