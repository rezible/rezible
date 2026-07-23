package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	atkc "github.com/rezible/rezible/ent/agentturnknowledgecitation"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
)

const (
	maxKnowledgeViewDepth         = 4
	maxKnowledgeViewEntities      = 200
	maxKnowledgeViewRelationships = 400
	maxKnowledgeViewEvidence      = 1000
)

type KnowledgeGraphService struct {
	db rez.Database
}

func NewKnowledgeGraphService(db rez.Database) (*KnowledgeGraphService, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &KnowledgeGraphService{db: db}, nil
}

func (s *KnowledgeGraphService) ListEntities(ctx context.Context, params rez.ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		WithSourceRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		}).
		WithTargetRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		})
	if params.Search != "" {
		query.Where(kne.DisplayNameContainsFold(params.Search))
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases().
		WithSourceRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.WithTargetEntity()
		}).
		WithTargetRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.WithSourceEntity()
		})
	return query.Only(ctx)
}

func (s *KnowledgeGraphService) GetEntityAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeEntity, error) {
	entity, queryErr := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases().
		Only(ctx)
	if queryErr != nil {
		return nil, queryErr
	}
	if referencedAt.IsZero() {
		return entity, nil
	}

	evidence, evidenceErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EffectiveAtLTE(referencedAt), ke.HasAliasWith(ksa.EntityID(id))).
		WithAlias().
		Order(ent.Desc(ke.FieldEffectiveAt), ent.Desc(ke.FieldCreatedAt)).
		All(ctx)
	if evidenceErr != nil {
		return nil, fmt.Errorf("query entity state: %w", evidenceErr)
	}
	seenAliases := mapset.NewSet[uuid.UUID]()
	for _, item := range evidence {
		alias, aliasErr := item.Edges.AliasOrErr()
		if aliasErr != nil || seenAliases.Contains(alias.ID) {
			continue
		}
		seenAliases.Add(alias.ID)
		if item.EvidenceKind == ke.EvidenceKindDeleted {
			continue
		}
		var state struct {
			Entity *ent.KnowledgeEntityRef `json:"entity"`
		}
		encoded, encodeErr := json.Marshal(item.SubjectState)
		if encodeErr != nil || json.Unmarshal(encoded, &state) != nil || state.Entity == nil {
			continue
		}
		historical := *entity
		historical.Kind = state.Entity.Kind
		historical.Reference = state.Entity.Reference
		historical.DisplayName = state.Entity.DisplayName
		historical.Description = state.Entity.Description
		historical.LiveProperties = state.Entity.Properties
		return &historical, nil
	}
	if referencedAt.Before(entity.CreatedAt) {
		return nil, fmt.Errorf("%w: entity did not exist at %s", rez.ErrConflict, referencedAt.Format(time.RFC3339))
	}
	return entity, nil
}

func (s *KnowledgeGraphService) GetRelationshipAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeRelationship, error) {
	relationship, queryErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases().
		WithSourceEntity().
		WithTargetEntity().
		Only(ctx)
	if queryErr != nil {
		return nil, queryErr
	}
	if referencedAt.IsZero() {
		return relationship, nil
	}

	evidence, evidenceErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EffectiveAtLTE(referencedAt), ke.HasAliasWith(ksa.RelationshipID(id))).
		WithAlias().
		Order(ent.Desc(ke.FieldEffectiveAt), ent.Desc(ke.FieldCreatedAt)).
		All(ctx)
	if evidenceErr != nil {
		return nil, fmt.Errorf("query relationship state: %w", evidenceErr)
	}
	seenAliases := mapset.NewSet[uuid.UUID]()
	for _, item := range evidence {
		alias, aliasErr := item.Edges.AliasOrErr()
		if aliasErr != nil || seenAliases.Contains(alias.ID) {
			continue
		}
		seenAliases.Add(alias.ID)
		if item.EvidenceKind == ke.EvidenceKindDeleted {
			continue
		}
		var state struct {
			Relationship *ent.KnowledgeRelationshipRef `json:"relationship"`
		}
		encoded, encodeErr := json.Marshal(item.SubjectState)
		if encodeErr != nil || json.Unmarshal(encoded, &state) != nil || state.Relationship == nil {
			continue
		}
		historical := *relationship
		historical.Kind = state.Relationship.Kind
		historical.Description = state.Relationship.Description
		historical.Properties = state.Relationship.Properties
		return &historical, nil
	}
	if referencedAt.Before(relationship.CreatedAt) {
		return nil, fmt.Errorf("%w: relationship did not exist at %s", rez.ErrConflict, referencedAt.Format(time.RFC3339))
	}
	return relationship, nil
}

func (s *KnowledgeGraphService) RecordTurnKnowledgeCitations(ctx context.Context, turnID uuid.UUID, citations []rez.KnowledgeCitation) error {
	if len(citations) == 0 {
		return nil
	}

	uniqueCitations := make(map[uuid.UUID]rez.KnowledgeCitation, len(citations))
	evidenceIDs := make([]uuid.UUID, 0, len(citations))
	for _, citation := range citations {
		if citation.EvidenceID == uuid.Nil {
			return fmt.Errorf("%w: knowledge evidence ID is required", rez.ErrInvalidInput)
		}
		citation.Summary = strings.TrimSpace(citation.Summary)
		if citation.Summary == "" {
			return fmt.Errorf("%w: knowledge citation summary is required", rez.ErrInvalidInput)
		}
		if _, exists := uniqueCitations[citation.EvidenceID]; exists {
			continue
		}
		uniqueCitations[citation.EvidenceID] = citation
		evidenceIDs = append(evidenceIDs, citation.EvidenceID)
	}

	evidence, queryErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.IDIn(evidenceIDs...)).
		WithAlias().
		All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query knowledge evidence: %w", queryErr)
	}
	if len(evidence) != len(evidenceIDs) {
		return fmt.Errorf("%w: one or more knowledge evidence records were not found", rez.ErrInvalidInput)
	}

	creates := make([]*ent.AgentTurnKnowledgeCitationCreate, 0, len(evidence))
	for _, item := range evidence {
		alias, aliasErr := item.Edges.AliasOrErr()
		if aliasErr != nil {
			return fmt.Errorf("get knowledge evidence alias: %w", aliasErr)
		}
		create := s.db.Client(ctx).AgentTurnKnowledgeCitation.Create().
			SetAgentTurnID(turnID).
			SetKnowledgeEvidenceID(item.ID).
			SetSummary(uniqueCitations[item.ID].Summary)
		if alias.EntityID != uuid.Nil {
			create.SetKnowledgeEntityID(alias.EntityID)
		}
		if alias.RelationshipID != uuid.Nil {
			create.SetKnowledgeRelationshipID(alias.RelationshipID)
		}
		creates = append(creates, create)
	}

	return s.db.Client(ctx).AgentTurnKnowledgeCitation.CreateBulk(creates...).
		OnConflictColumns(atkc.FieldTenantID, atkc.FieldAgentTurnID, atkc.FieldKnowledgeEvidenceID).
		DoNothing().
		Exec(ctx)
}

func (s *KnowledgeGraphService) ListRelationships(ctx context.Context, params rez.ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithSourceEntity().
		WithTargetEntity().
		WithAliases()
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetView(ctx context.Context, rootID uuid.UUID, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	depth := params.Depth
	if depth < 1 {
		depth = 1
	}
	if depth > maxKnowledgeViewDepth {
		depth = maxKnowledgeViewDepth
	}

	queryRoot := s.db.Client(ctx).KnowledgeEntity.Query().Where(kne.ID(rootID)).WithAliases()
	root, queryErr := queryRoot.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query root entity: %w", queryErr)
	}
	if !knowledgeAliasesAreLive(root.Edges.Aliases) {
		return nil, fmt.Errorf("%w: root entity is deleted", rez.ErrConflict)
	}

	entityIDs := mapset.NewSet(rootID)
	relationshipIDs := mapset.NewSet[uuid.UUID]()
	frontier := []uuid.UUID{rootID}
	truncated := false

	for level := 0; level < depth && len(frontier) > 0; level++ {
		queryRelationships := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.Or(knr.SourceEntityIDIn(frontier...), knr.TargetEntityIDIn(frontier...))).
			WithAliases().
			WithSourceEntity(func(query *ent.KnowledgeEntityQuery) { query.WithAliases() }).
			WithTargetEntity(func(query *ent.KnowledgeEntityQuery) { query.WithAliases() }).
			Order(ent.Asc(knr.FieldID)).
			Limit(maxKnowledgeViewRelationships + 1)
		if len(params.RelationshipKinds) > 0 {
			queryRelationships.Where(knr.KindIn(params.RelationshipKinds...))
		}
		relationships, queryErr := queryRelationships.All(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("query graph level %d: %w", level+1, queryErr)
		}
		if len(relationships) > maxKnowledgeViewRelationships {
			relationships = relationships[:maxKnowledgeViewRelationships]
			truncated = true
		}

		next := make([]uuid.UUID, 0)
		for _, relationship := range relationships {
			if !knowledgeRelationshipIsLive(relationship) {
				continue
			}
			if relationshipIDs.Contains(relationship.ID) {
				continue
			}
			if relationshipIDs.Cardinality() >= maxKnowledgeViewRelationships {
				truncated = true
				break
			}

			newEntityIDs := make([]uuid.UUID, 0, 2)
			if !entityIDs.Contains(relationship.SourceEntityID) {
				newEntityIDs = append(newEntityIDs, relationship.SourceEntityID)
			}
			if !entityIDs.Contains(relationship.TargetEntityID) {
				newEntityIDs = append(newEntityIDs, relationship.TargetEntityID)
			}
			if entityIDs.Cardinality()+len(newEntityIDs) > maxKnowledgeViewEntities {
				truncated = true
				continue
			}

			relationshipIDs.Add(relationship.ID)
			for _, entityID := range newEntityIDs {
				entityIDs.Add(entityID)
				next = append(next, entityID)
			}
		}
		frontier = next
	}

	selectedEntityIDs := entityIDs.ToSlice()
	queryEntities := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.IDIn(selectedEntityIDs...)).
		WithAliases().
		Order(ent.Asc(kne.FieldKind), ent.Asc(kne.FieldDisplayName), ent.Asc(kne.FieldID))
	entities, queryEntitiesErr := queryEntities.All(ctx)
	if queryEntitiesErr != nil {
		return nil, fmt.Errorf("load graph entities: %w", queryEntitiesErr)
	}

	selectedRelationshipIDs := relationshipIDs.ToSlice()
	relationships := ent.KnowledgeRelationships{}
	if len(selectedRelationshipIDs) > 0 {
		queryRelationships := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.IDIn(selectedRelationshipIDs...)).
			WithSourceEntity(func(query *ent.KnowledgeEntityQuery) { query.WithAliases() }).
			WithTargetEntity(func(query *ent.KnowledgeEntityQuery) { query.WithAliases() }).
			WithAliases().
			Order(ent.Asc(knr.FieldKind), ent.Asc(knr.FieldID))
		var queryRelationshipsErr error
		relationships, queryRelationshipsErr = queryRelationships.All(ctx)
		if queryRelationshipsErr != nil {
			return nil, fmt.Errorf("load graph relationships: %w", queryRelationshipsErr)
		}
	}

	evidence, evidenceTruncated, evidenceErr := s.queryViewEvidence(ctx, selectedEntityIDs, selectedRelationshipIDs)
	if evidenceErr != nil {
		return nil, evidenceErr
	}
	truncated = truncated || evidenceTruncated
	warnings := make([]string, 0, 1)
	if truncated {
		warnings = append(warnings, "graph context reached its result limit")
	}
	return &rez.KnowledgeGraphView{
		Entities:      entities,
		Relationships: relationships,
		Evidence:      evidence,
		Truncated:     truncated,
		Warnings:      warnings,
	}, nil
}

func knowledgeAliasesAreLive(aliases ent.KnowledgeSubjectAliasSlice) bool {
	if len(aliases) == 0 {
		return true
	}
	for _, alias := range aliases {
		if alias.DeletedAt == nil {
			return true
		}
	}
	return false
}

func knowledgeRelationshipIsLive(relationship *ent.KnowledgeRelationship) bool {
	if !knowledgeAliasesAreLive(relationship.Edges.Aliases) {
		return false
	}
	source, sourceErr := relationship.Edges.SourceEntityOrErr()
	target, targetErr := relationship.Edges.TargetEntityOrErr()
	return sourceErr == nil && targetErr == nil &&
		knowledgeAliasesAreLive(source.Edges.Aliases) && knowledgeAliasesAreLive(target.Edges.Aliases)
}

func (s *KnowledgeGraphService) queryViewEvidence(ctx context.Context, entityIDs, relationshipIDs []uuid.UUID) (ent.KnowledgeEvidences, bool, error) {
	aliasConditions := make([]predicate.KnowledgeSubjectAlias, 0, 2)
	if len(entityIDs) > 0 {
		aliasConditions = append(aliasConditions, ksa.EntityIDIn(entityIDs...))
	}
	if len(relationshipIDs) > 0 {
		aliasConditions = append(aliasConditions, ksa.RelationshipIDIn(relationshipIDs...))
	}
	if len(aliasConditions) == 0 {
		return ent.KnowledgeEvidences{}, false, nil
	}

	queryEvidence := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.HasAliasWith(ksa.Or(aliasConditions...))).
		WithAlias().
		WithEvent().
		Order(ent.Desc(ke.FieldEffectiveAt), ent.Asc(ke.FieldID)).
		Limit(maxKnowledgeViewEvidence + 1)
	evidence, queryErr := queryEvidence.All(ctx)
	if queryErr != nil {
		return nil, false, fmt.Errorf("load graph evidence: %w", queryErr)
	}
	truncated := len(evidence) > maxKnowledgeViewEvidence
	if truncated {
		evidence = evidence[:maxKnowledgeViewEvidence]
	}
	return evidence, truncated, nil
}
