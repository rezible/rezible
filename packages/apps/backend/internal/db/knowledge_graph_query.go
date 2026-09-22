package db

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeGraphQueryService struct {
	db rez.Database
}

func NewKnowledgeGraphQueryService(db rez.Database) (*KnowledgeGraphQueryService, error) {
	return &KnowledgeGraphQueryService{db: db}, nil
}

func (s *KnowledgeGraphQueryService) ListEntities(ctx context.Context, params rez.ListKnowledgeEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		Order(kne.ByID(params.GetOrder()))

	if search := strings.TrimSpace(params.Search); search != "" {
		// explicitly do not support search for now
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}

	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.entityQueryWithEvidence(ctx, id).Only(ctx)
}

func (s *KnowledgeGraphQueryService) ListRelationships(ctx context.Context, params rez.ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases().
		Order(knr.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetRelationship(ctx context.Context, id uuid.UUID) (*ent.KnowledgeRelationship, error) {
	return s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases().
		WithSourceEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases()
		}).
		WithTargetEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases()
		}).
		Only(ctx)
}

func (s *KnowledgeGraphQueryService) relationshipQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeRelationshipQuery {
	return s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases(subjectAliasWithEvidence(evP...))
}

func (s *KnowledgeGraphQueryService) entityQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeEntityQuery {
	return s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases(subjectAliasWithEvidence(evP...))
}

func subjectAliasWithEvidence(evidencePreds ...predicate.KnowledgeEvidence) func(*ent.KnowledgeSubjectAliasQuery) {
	return func(aq *ent.KnowledgeSubjectAliasQuery) {
		//aq.Where(subjectAliasPred)
		aq.WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.Where(evidencePreds...)
		})
	}
}

func (s *KnowledgeGraphQueryService) ListSubjectAliases(ctx context.Context, params rez.ListKnowledgeSubjectAliasesParams) (*ent.ListResult[ent.KnowledgeSubjectAlias], error) {
	query := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(params.Predicates...).
		Order(ksa.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeSubjectAlias, *ent.KnowledgeSubjectAliasQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetSubjectAlias(ctx context.Context, id uuid.UUID) (*ent.KnowledgeSubjectAlias, error) {
	return s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ID(id)).
		Only(ctx)
}

func (s *KnowledgeGraphQueryService) ListEvidence(ctx context.Context, params rez.ListKnowledgeEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error) {
	query := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(params.Predicates...).
		Order(ke.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEvidence, *ent.KnowledgeEvidenceQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetEvidence(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEvidence, error) {
	return s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.ID(id)).
		WithSubjectAlias().
		Only(ctx)
}

func (s *KnowledgeGraphQueryService) Query(
	ctx context.Context,
	query rez.KnowledgeGraphQuery,
	params rez.KnowledgeGraphPageParams,
) (*rez.KnowledgeGraphQueryResult, error) {
	if params.Limit < 0 {
		return nil, fmt.Errorf("%w: page limit cannot be negative", rez.ErrInvalidInput)
	}

	return nil, nil
}

type graphQueryHashInput struct {
	TenantID int         `json:"tenant_id"`
	RootIDs  []uuid.UUID `json:"root_ids"`
	Level    int         `json:"level"`
}

func (hi *graphQueryHashInput) makeHash() (string, error) {
	slices.SortFunc(hi.RootIDs, func(a, b uuid.UUID) int {
		return cmp.Compare(a.String(), b.String())
	})
	encodedInput, marshalErr := json.Marshal(hi)
	if marshalErr != nil {
		return "", fmt.Errorf("hash knowledge graph query: %w", marshalErr)
	}
	return fmt.Sprintf("%x", sha256.Sum256(encodedInput)), nil
}

type graphLevelCursor struct {
	QueryHash string
	Offset    int
}

func (c *graphLevelCursor) encode() (rez.KnowledgeGraphPageCursor, error) {
	encodedCursor, marshalErr := json.Marshal(c)
	if marshalErr != nil {
		return "", fmt.Errorf("%w: encode knowledge graph cursor: %w", rez.ErrInvalidInput, marshalErr)
	}
	return rez.KnowledgeGraphPageCursor(base64.RawURLEncoding.EncodeToString(encodedCursor)), nil
}

func (c *graphLevelCursor) decode(encoded rez.KnowledgeGraphPageCursor) error {
	decodedCursor, decodeErr := base64.RawURLEncoding.DecodeString(string(encoded))
	if decodeErr != nil {
		decodedCursor, decodeErr = base64.URLEncoding.DecodeString(string(encoded))
	}
	if decodeErr != nil {
		return fmt.Errorf("%w: malformed page cursor", rez.ErrInvalidInput)
	}
	if unmarshalErr := json.Unmarshal(decodedCursor, c); unmarshalErr != nil || c.QueryHash == "" {
		return fmt.Errorf("%w: malformed page cursor", rez.ErrInvalidInput)
	}
	return nil
}

func (s *KnowledgeGraphQueryService) ListConnectionRelationships(
	ctx context.Context,
	params rez.ListKnowledgeGraphConnectionRelationshipsParams,
) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	return nil, nil
}
