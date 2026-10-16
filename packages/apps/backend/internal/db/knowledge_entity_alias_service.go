package db

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/integrations"
)

// KnowledgeEntityAliasService reads entities' aliases with the name each provider uses, its integration's
// display name and when Rezible last recorded it. It holds the integration registry, which the
// knowledge graph query service cannot, since integrations' projections depend on that service.
type KnowledgeEntityAliasService struct {
	query *KnowledgeGraphQueryService
	names integrationNames
}

func NewKnowledgeEntityAliasService(db rez.Database, registry *integrations.Registry) (*KnowledgeEntityAliasService, error) {
	return &KnowledgeEntityAliasService{
		query: &KnowledgeGraphQueryService{db: db},
		names: integrationNames{registry: registry},
	}, nil
}

func (s *KnowledgeEntityAliasService) ListEntityAliases(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]rez.KnowledgeEntityAlias, error) {
	byEntity := make(map[uuid.UUID][]rez.KnowledgeEntityAlias, len(entityIDs))
	if len(entityIDs) == 0 {
		return byEntity, nil
	}
	queryAliases := s.query.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.EntityIDIn(entityIDs...)).
		WithMergedByProposal(func(pq *ent.KnowledgeProposalQuery) {
			pq.WithDecisions(func(dq *ent.KnowledgeProposalDecisionQuery) {
				dq.WithUser().Order(kpd.ByCreatedAt(), kpd.ByID())
			})
		}).
		Order(ksa.ByID())
	aliases, aliasesErr := queryAliases.All(ctx)
	if aliasesErr != nil {
		return nil, fmt.Errorf("query entity aliases: %w", aliasesErr)
	}
	if len(aliases) == 0 {
		return byEntity, nil
	}

	aliasIDs := make([]uuid.UUID, len(aliases))
	for index, alias := range aliases {
		aliasIDs[index] = alias.ID
	}
	grouping := knowledgeEvidenceGrouping{predicates: []predicate.KnowledgeEvidence{ke.SubjectAliasIDIn(aliasIDs...)}}
	summaries, summarizeErr := s.query.summarizeEvidenceGroups(ctx, grouping, 0, 0)
	if summarizeErr != nil {
		return nil, fmt.Errorf("summarize alias evidence: %w", summarizeErr)
	}
	summaryByAlias := make(map[uuid.UUID]*rez.KnowledgeEvidenceSummary, len(summaries))
	for _, summary := range summaries {
		summaryByAlias[summary.SubjectAliasID] = summary
	}

	for _, alias := range aliases {
		entityAlias := rez.KnowledgeEntityAlias{
			Alias:                  alias,
			IntegrationDisplayName: s.names.displayName(alias.Provider),
			Name:                   alias.ProviderResourceRef,
		}
		if summary, recorded := summaryByAlias[alias.ID]; recorded {
			entityAlias.LastRecordedAt = summary.LatestAt
			if summary.LatestState.DisplayName != "" {
				entityAlias.Name = summary.LatestState.DisplayName
			}
		}
		byEntity[*alias.EntityID] = append(byEntity[*alias.EntityID], entityAlias)
	}
	for _, entityAliases := range byEntity {
		slices.SortStableFunc(entityAliases, func(a, b rez.KnowledgeEntityAlias) int {
			return b.LastRecordedAt.Compare(a.LastRecordedAt)
		})
	}
	return byEntity, nil
}

func (s *KnowledgeEntityAliasService) ListEntitiesWithAliases(ctx context.Context, params rez.ListKnowledgeEntitiesParams) (*ent.ListResult[rez.KnowledgeEntityWithAliases], error) {
	entities, listErr := s.query.ListEntities(ctx, params)
	if listErr != nil {
		return nil, listErr
	}
	ids := make([]uuid.UUID, len(entities.Data))
	for i, entity := range entities.Data {
		ids[i] = entity.ID
	}
	aliases, aliasesErr := s.ListEntityAliases(ctx, ids)
	if aliasesErr != nil {
		return nil, fmt.Errorf("list entity aliases: %w", aliasesErr)
	}
	result := &ent.ListResult[rez.KnowledgeEntityWithAliases]{
		Data:     make([]*rez.KnowledgeEntityWithAliases, len(entities.Data)),
		Page:     entities.Page,
		PageSize: entities.PageSize,
		Total:    entities.Total,
	}
	for i, entity := range entities.Data {
		result.Data[i] = &rez.KnowledgeEntityWithAliases{Entity: entity, Aliases: aliases[entity.ID]}
	}
	return result, nil
}

func (s *KnowledgeEntityAliasService) GetEntityWithAliases(ctx context.Context, id uuid.UUID, expand rez.KnowledgeEntityExpand) (*rez.KnowledgeEntityWithAliases, error) {
	entity, getErr := s.query.GetEntity(ctx, id, expand)
	if getErr != nil {
		return nil, getErr
	}
	aliases, aliasesErr := s.ListEntityAliases(ctx, []uuid.UUID{entity.ID})
	if aliasesErr != nil {
		return nil, fmt.Errorf("list entity aliases: %w", aliasesErr)
	}
	return &rez.KnowledgeEntityWithAliases{Entity: entity, Aliases: aliases[entity.ID]}, nil
}
