package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

const (
	defaultEntityQueryLimit = 100
	maxEntityQueryLimit     = 500
)

func (s *KnowledgeGraphQueryService) SelectGraphEntities(ctx context.Context, params rez.SelectKnowledgeGraphEntitiesParams) (*rez.KnowledgeGraphEntitiesPage, error) {
	selection := &knowledgeEntitySelection{
		Categories: params.Filter.Categories,
		Kinds:      params.Filter.Kinds,
		Limit:      defaultEntityQueryLimit,
	}
	if params.Limit != nil {
		selection.Limit = *params.Limit
	}
	if normalizeErr := selection.normalize(); normalizeErr != nil {
		return nil, fmt.Errorf("normalize entity selection: %w", normalizeErr)
	}

	fingerprint, fingerprintErr := selection.makeEntityFingerprint()
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint entity selection: %w", fingerprintErr)
	}

	if params.Cursor != nil {
		cursor, cursorErr := s.decodeGraphQueryCursor(string(*params.Cursor), fingerprint, true)
		if cursorErr != nil {
			return nil, fmt.Errorf("decode entity selection cursor: %w", cursorErr)
		}
		selection.AfterID = cursor.AfterID
	}

	reference, encodeErr := s.encodeEntitySelectionRef(selection)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode entity selection reference: %w", encodeErr)
	}

	page := &rez.KnowledgeGraphEntitiesPage{
		Entities:           make([]*ent.KnowledgeEntity, 0),
		EntitySelectionRef: reference,
	}

	selectEntitySummaryFields := s.selectedEntitiesQuery(ctx, selection, +1).
		Select(kne.FieldID, kne.FieldCategory, kne.FieldKind)
	entities, queryErr := selectEntitySummaryFields.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("select knowledge graph entities: %w", queryErr)
	}
	if len(entities) > selection.Limit {
		page.Entities = entities[:selection.Limit]
		descriptor := knowledgeGraphQueryCursor{
			Version:       knowledgeGraphQueryVersion,
			IsEntityQuery: true,
			AfterID:       new(page.Entities[len(page.Entities)-1].ID),
			Fingerprint:   fingerprint,
		}
		encodedCursor, cursorErr := descriptor.encode()
		if cursorErr != nil {
			return nil, fmt.Errorf("encode entity selection cursor: %w", cursorErr)
		}
		page.NextCursor = new(rez.EntitySelectionCursor(encodedCursor))
	} else if entities != nil {
		page.Entities = entities
	}
	return page, nil
}

func (s *KnowledgeGraphQueryService) selectedEntitiesQuery(ctx context.Context, selection *knowledgeEntitySelection, extraAmount int) *ent.KnowledgeEntityQuery {
	query := s.db.Client(ctx).KnowledgeEntity.Query()
	if len(selection.Categories) > 0 {
		query.Where(kne.CategoryIn(selection.Categories...))
	}
	if len(selection.Kinds) > 0 {
		query.Where(kne.KindIn(selection.Kinds...))
	}
	if selection.AfterID != nil {
		query.Where(kne.IDGT(*selection.AfterID))
	}
	return query.Order(kne.ByID()).Limit(selection.Limit + extraAmount)
}

const (
	defaultRelationshipLimit = 200
	maxRelationshipLimit     = 1000
)

func (s *KnowledgeGraphQueryService) ExpandGraphRelationships(ctx context.Context, params rez.ExpandKnowledgeGraphRelationshipsParams) (*rez.KnowledgeGraphRelationshipsPage, error) {
	selection, selectionErr := s.decodeEntitySelectionRef(params.EntitySelectionRef)
	if selectionErr != nil {
		return nil, fmt.Errorf("decode entity selection reference: %w", selectionErr)
	}

	predicates := slices.Clone(params.Predicates)
	for _, relationshipPredicate := range predicates {
		if validationErr := knr.PredicateValidator(relationshipPredicate); validationErr != nil {
			return nil, fmt.Errorf("%w: invalid relationship predicate", rez.ErrInvalidInput)
		}
	}
	slices.Sort(predicates)
	predicates = slices.Compact(predicates)

	limit := defaultRelationshipLimit
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > maxRelationshipLimit {
			return nil, fmt.Errorf("%w: limit is outside the supported range", rez.ErrInvalidInput)
		}
		limit = *params.Limit
	}

	fingerprint, fingerprintErr := selection.makeRelationshipFingerprint(predicates, limit)
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint relationship expansion: %w", fingerprintErr)
	}

	var afterID *uuid.UUID
	if params.Cursor != nil {
		cursor, cursorErr := s.decodeGraphQueryCursor(string(*params.Cursor), fingerprint, false)
		if cursorErr != nil {
			return nil, fmt.Errorf("decode relationship expansion cursor: %w", cursorErr)
		}
		afterID = cursor.AfterID
	}

	entityQuery := s.selectedEntitiesQuery(ctx, selection, 0)
	entityIDs, entityQueryErr := entityQuery.IDs(ctx)
	if entityQueryErr != nil {
		return nil, fmt.Errorf("replay selected knowledge graph entities: %w", entityQueryErr)
	}

	page := &rez.KnowledgeGraphRelationshipsPage{
		Relationships: make([]*ent.KnowledgeRelationship, 0),
	}
	if len(entityIDs) == 0 {
		return page, nil
	}

	relationshipQuery := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		Order(knr.ByID()).
		Limit(limit + 1)
	if len(predicates) > 0 {
		relationshipQuery.Where(knr.PredicateIn(predicates...))
	}
	if afterID != nil {
		relationshipQuery.Where(knr.IDGT(*afterID))
	}
	selectSummaryFields := relationshipQuery.Select(
		knr.FieldID,
		knr.FieldSourceEntityID,
		knr.FieldTargetEntityID,
		knr.FieldPredicate,
	)
	relationships, queryErr := selectSummaryFields.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query knowledge graph relationships: %w", queryErr)
	}
	if len(relationships) > limit {
		page.Relationships = relationships[:limit]
		descriptor := knowledgeGraphQueryCursor{
			Version:       knowledgeGraphQueryVersion,
			IsEntityQuery: false,
			AfterID:       new(page.Relationships[len(page.Relationships)-1].ID),
			Fingerprint:   fingerprint,
		}
		encodedCursor, cursorErr := descriptor.encode()
		if cursorErr != nil {
			return nil, fmt.Errorf("encode relationship expansion cursor: %w", cursorErr)
		}
		page.NextCursor = new(rez.RelationshipExpansionCursor(encodedCursor))
	} else if relationships != nil {
		page.Relationships = relationships
	}
	return page, nil
}

const knowledgeGraphQueryVersion = 1

type knowledgeEntitySelection struct {
	Version    int            `json:"version"`
	Categories []kne.Category `json:"categories,omitempty"`
	Kinds      []string       `json:"kinds,omitempty"`
	AfterID    *uuid.UUID     `json:"afterId,omitempty"`
	Limit      int            `json:"limit"`
}

func (es *knowledgeEntitySelection) normalize() error {
	if es.Limit < 1 || es.Limit > maxEntityQueryLimit {
		return fmt.Errorf("%w: entity limit is outside the supported range", rez.ErrInvalidInput)
	}

	for _, category := range es.Categories {
		if validationErr := kne.CategoryValidator(category); validationErr != nil {
			return fmt.Errorf("%w: invalid entity category", rez.ErrInvalidInput)
		}
	}
	categories := slices.Clone(es.Categories)
	slices.Sort(categories)

	kinds := slices.Clone(es.Kinds)
	if slices.Contains(kinds, "") {
		return fmt.Errorf("%w: entity kinds must be non-empty", rez.ErrInvalidInput)
	}
	slices.Sort(kinds)

	es.Version = knowledgeGraphQueryVersion
	es.Categories = slices.Compact(categories)
	es.Kinds = slices.Compact(kinds)
	return nil
}

func (es *knowledgeEntitySelection) encodeFingerprint[F any](payload F) (string, error) {
	jsonPayload, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal fingerprint: %w", marshalErr)
	}
	hash := sha256.Sum256(jsonPayload)
	return hex.EncodeToString(hash[:]), nil
}

type selectKnowledgeEntitiesFingerprint struct {
	Categories []kne.Category `json:"categories,omitempty"`
	Kinds      []string       `json:"kinds,omitempty"`
	Limit      int            `json:"limit"`
}

func (es *knowledgeEntitySelection) makeEntityFingerprint() (string, error) {
	return es.encodeFingerprint(selectKnowledgeEntitiesFingerprint{
		Categories: es.Categories,
		Kinds:      es.Kinds,
		Limit:      es.Limit,
	})
}

type expandKnowledgeRelationshipsFingerprint struct {
	Selection  knowledgeEntitySelection `json:"selection"`
	Predicates []knr.Predicate          `json:"predicates,omitempty"`
	Limit      int                      `json:"limit"`
}

func (es *knowledgeEntitySelection) makeRelationshipFingerprint(predicates []knr.Predicate, limit int) (string, error) {
	return es.encodeFingerprint(expandKnowledgeRelationshipsFingerprint{
		Selection:  *es,
		Predicates: predicates,
		Limit:      limit,
	})
}

const (
	maxKnowledgeEntitySelectionRefSize        = 4 * 1024
	knowledgeEntitySelectionBoundaryAllowance = len(`,"afterId":""`) + 36
)

func (s *KnowledgeGraphQueryService) encodeEntitySelectionRef(es *knowledgeEntitySelection) (rez.EntitySelectionRef, error) {
	jsonBytes, marshalErr := json.Marshal(es)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal entity selection reference: %w", marshalErr)
	}

	// check that accepted filters also fit on continuation pages with the afterId field set
	reservedSize := len(jsonBytes)
	if es.AfterID == nil {
		reservedSize += knowledgeEntitySelectionBoundaryAllowance
	}
	if base64.RawURLEncoding.EncodedLen(reservedSize) > maxKnowledgeEntitySelectionRefSize {
		return "", fmt.Errorf("%w: entity selection reference exceeds the maximum size", rez.ErrInvalidInput)
	}

	return rez.EntitySelectionRef(base64.RawURLEncoding.EncodeToString(jsonBytes)), nil
}

func (s *KnowledgeGraphQueryService) decodeRef[T any](ref string) (*T, error) {
	refJson, decodeErr := base64.RawURLEncoding.Strict().DecodeString(ref)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid encoding", rez.ErrInvalidInput)
	}
	decoder := json.NewDecoder(bytes.NewReader(refJson))
	decoder.DisallowUnknownFields()
	var t T
	if jsonErr := decoder.Decode(&t); jsonErr != nil {
		return nil, fmt.Errorf("%w: invalid reference JSON", rez.ErrInvalidInput)
	}
	if _, trailingErr := decoder.Token(); trailingErr != io.EOF {
		return nil, fmt.Errorf("%w: trailing data in reference", rez.ErrInvalidInput)
	}
	return &t, nil
}

func (s *KnowledgeGraphQueryService) decodeEntitySelectionRef(ref rez.EntitySelectionRef) (*knowledgeEntitySelection, error) {
	if ref == "" || len(ref) > maxKnowledgeEntitySelectionRefSize {
		return nil, fmt.Errorf("%w: invalid entity selection reference size", rez.ErrInvalidInput)
	}
	selection, decodeErr := s.decodeRef[knowledgeEntitySelection](string(ref))
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid entity selection reference: %w", rez.ErrInvalidInput, decodeErr)
	}
	if selection.Version != knowledgeGraphQueryVersion {
		return nil, fmt.Errorf("%w: unsupported entity selection reference", rez.ErrInvalidInput)
	}
	if normalizeErr := selection.normalize(); normalizeErr != nil {
		return nil, fmt.Errorf("%w: invalid entity selection reference", rez.ErrInvalidInput)
	}
	return selection, nil
}

type knowledgeGraphQueryCursor struct {
	Version       int        `json:"version"`
	IsEntityQuery bool       `json:"is_entity_query"`
	AfterID       *uuid.UUID `json:"afterId"`
	Fingerprint   string     `json:"fingerprint"`
}

const maxQueryCursorSize = 1 * 1024

func (cursor *knowledgeGraphQueryCursor) encode() (string, error) {
	jsonBytes, marshalErr := json.Marshal(cursor)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal query cursor: %w", marshalErr)
	}
	encoded := base64.RawURLEncoding.EncodeToString(jsonBytes)
	if len(encoded) > maxQueryCursorSize {
		return "", fmt.Errorf("%w: query cursor exceeds the maximum size", rez.ErrInvalidInput)
	}
	return encoded, nil
}

func (s *KnowledgeGraphQueryService) decodeGraphQueryCursor(encoded string, fingerprint string, isEntityQuery bool) (*knowledgeGraphQueryCursor, error) {
	if encoded == "" || len(encoded) > maxQueryCursorSize {
		return nil, fmt.Errorf("%w: invalid query cursor size", rez.ErrInvalidInput)
	}
	cursor, decodeErr := s.decodeRef[knowledgeGraphQueryCursor](encoded)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid cursor: %w", rez.ErrInvalidInput, decodeErr)
	}
	if cursor.Version != knowledgeGraphQueryVersion || cursor.IsEntityQuery != isEntityQuery {
		return nil, fmt.Errorf("%w: unsupported query cursor", rez.ErrInvalidInput)
	}
	if cursor.Fingerprint != fingerprint {
		return nil, fmt.Errorf("%w: query cursor does not match the query", rez.ErrInvalidInput)
	}
	if cursor.AfterID == nil {
		return nil, fmt.Errorf("%w: missing query cursor boundary", rez.ErrInvalidInput)
	}
	return cursor, nil
}
