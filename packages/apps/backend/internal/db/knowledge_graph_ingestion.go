package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

type resolvedSubjectAlias struct {
	aliasId   uuid.UUID
	subjectId uuid.UUID
}

func (s *KnowledgeGraphService) lookupExistingEntityByRef(ctx context.Context, ref ent.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	queryAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ref.SubjectAliasRef.ResourcePredicate()).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query entity alias: %w", queryErr)
	}
	if alias == nil {
		return nil, nil
	} else if alias.SubjectKind != ksa.SubjectKindEntity {
		return nil, fmt.Errorf("%w: resource already identifies a relationship", rez.ErrConflict)
	}
	entity, entityErr := alias.Edges.EntityOrErr()
	if entityErr != nil || entity == nil {
		return nil, fmt.Errorf("load alias entity: %w", entityErr)
	} else if entity.Category != ref.Category || entity.Kind != ref.Kind {
		return nil, fmt.Errorf("%w: alias identifies %q/%q, evidence expects %q/%q", rez.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entity.ID}, nil
}

var knowledgeSubjectAliasUniqueColumns = sql.ConflictColumns(ksa.FieldTenantID, ksa.FieldProvider, ksa.FieldProviderSource, ksa.FieldProviderSubjectRef)

func (s *KnowledgeGraphService) resolveEntityFromRef(ctx context.Context, ref ent.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	existing, existingErr := s.lookupExistingEntityByRef(ctx, ref)
	if existingErr != nil {
		return nil, fmt.Errorf("lookup existing: %w", existingErr)
	} else if existing != nil {
		return existing, nil
	}

	createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
		SetCategory(ref.Category).
		SetKind(ref.Kind)
	createdEntity, createEntityErr := createEntity.Save(ctx)
	if createEntityErr != nil {
		return nil, fmt.Errorf("create entity: %w", createEntityErr)
	}

	createAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetProvider(ref.SubjectAliasRef.Provider).
		SetProviderSource(ref.SubjectAliasRef.ProviderSource).
		SetProviderSubjectRef(ref.SubjectAliasRef.ProviderSubjectRef).
		SetSubjectKind(ksa.SubjectKindEntity).
		SetEntityID(createdEntity.ID).
		OnConflict(knowledgeSubjectAliasUniqueColumns).
		Ignore()
	aliasId, createAliasErr := createAlias.ID(ctx)
	if createAliasErr != nil {
		return nil, fmt.Errorf("create alias: %w", createAliasErr)
	}
	return &resolvedSubjectAlias{aliasId: aliasId, subjectId: createdEntity.ID}, nil
}

func (s *KnowledgeGraphService) lookupExistingRelationshipByRef(ctx context.Context, ref ent.KnowledgeRelationshipRef, sourceEntityID, targetEntityID uuid.UUID) (*resolvedSubjectAlias, error) {
	queryAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ref.SubjectAliasRef.ResourcePredicate()).
		WithRelationship()
	existingAlias, queryExistingErr := queryAlias.Only(ctx)
	if queryExistingErr != nil && !ent.IsNotFound(queryExistingErr) {
		return nil, fmt.Errorf("query existing relationship alias: %w", queryExistingErr)
	}
	if existingAlias == nil {
		return nil, nil
	} else if existingAlias.SubjectKind != ksa.SubjectKindRelationship {
		return nil, fmt.Errorf("%w: resource already identifies an entity", rez.ErrConflict)
	}
	rel, edgeErr := existingAlias.Edges.RelationshipOrErr()
	if edgeErr != nil {
		return nil, fmt.Errorf("load aliased relationship: %w", edgeErr)
	}
	if rel.Predicate != ref.Predicate || rel.SourceEntityID != sourceEntityID || rel.TargetEntityID != targetEntityID {
		return nil, fmt.Errorf("%w: relationship alias identifies different topology", rez.ErrConflict)
	}
	return &resolvedSubjectAlias{aliasId: existingAlias.ID, subjectId: rel.ID}, nil
}

var knowledgeRelationshipUniqueColumns = sql.ConflictColumns(knr.FieldTenantID, knr.FieldPredicate, knr.FieldSourceEntityID, knr.FieldTargetEntityID)

func (s *KnowledgeGraphService) resolveRelationshipFromRef(ctx context.Context, ref ent.KnowledgeRelationshipRef) (*resolvedSubjectAlias, error) {
	source, sourceErr := s.resolveEntityFromRef(ctx, ref.Source)
	if sourceErr != nil || source == nil {
		return nil, fmt.Errorf("resolve source: %w", sourceErr)
	}

	target, targetErr := s.resolveEntityFromRef(ctx, ref.Target)
	if targetErr != nil || target == nil {
		return nil, fmt.Errorf("resolve target: %w", targetErr)
	}

	existing, lookupExistingErr := s.lookupExistingRelationshipByRef(ctx, ref, source.subjectId, target.subjectId)
	if lookupExistingErr != nil {
		return nil, fmt.Errorf("lookup existing: %w", lookupExistingErr)
	}
	if existing != nil {
		return existing, nil
	}

	upsertRel := s.db.Client(ctx).KnowledgeRelationship.Create().
		SetPredicate(ref.Predicate).
		SetSourceEntityID(source.subjectId).
		SetTargetEntityID(target.subjectId).
		OnConflict(knowledgeRelationshipUniqueColumns).
		Ignore()
	relationshipId, upsertRelErr := upsertRel.ID(ctx)
	if upsertRelErr != nil {
		return nil, fmt.Errorf("upsert relationship: %w", upsertRelErr)
	}

	createAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetProvider(ref.SubjectAliasRef.Provider).
		SetProviderSource(ref.SubjectAliasRef.ProviderSource).
		SetProviderSubjectRef(ref.SubjectAliasRef.ProviderSubjectRef).
		SetSubjectKind(ksa.SubjectKindRelationship).
		SetRelationshipID(relationshipId).
		OnConflict(knowledgeSubjectAliasUniqueColumns).
		Ignore()
	aliasId, createAliasErr := createAlias.ID(ctx)
	if createAliasErr != nil {
		return nil, fmt.Errorf("create alias: %w", createAliasErr)
	}
	return &resolvedSubjectAlias{aliasId: aliasId, subjectId: relationshipId}, nil
}

func (s *KnowledgeGraphService) resolveEvidenceSubjectAlias(ctx context.Context, ref ent.KnowledgeEvidenceRef) (*resolvedSubjectAlias, error) {
	switch {
	case ref.SubjectEntity != nil && ref.SubjectRelationship == nil:
		return s.resolveEntityFromRef(ctx, *ref.SubjectEntity)
	case ref.SubjectRelationship != nil && ref.SubjectEntity == nil:
		return s.resolveRelationshipFromRef(ctx, *ref.SubjectRelationship)
	default:
		return nil, fmt.Errorf("evidence must contain exactly one entity or relationship")
	}
}

var knowledgeEvidenceUniqueColumns = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldSubjectAliasID)

func (s *KnowledgeGraphService) makeEvidenceBuilderFromRef(client *ent.Client, eventId uuid.UUID, aliasId uuid.UUID, ref ent.KnowledgeEvidenceRef) *ent.KnowledgeEvidenceCreate {
	return client.KnowledgeEvidence.Create().
		SetEventID(eventId).
		SetSubjectAliasID(aliasId).
		SetKind(ref.Kind).
		SetAssertion(ref.Assertion).
		SetEffectiveAt(ref.EffectiveAt).
		SetSubjectState(ref.SubjectState)
}

func (s *KnowledgeGraphService) acquireRefTransactionLocks(ctx context.Context, refs []ent.KnowledgeEvidenceRef) error {
	locks := make([]string, len(refs))
	for i, ref := range refs {
		if ref.SubjectEntity != nil && ref.SubjectRelationship == nil {
			locks[i] = ref.SubjectEntity.SubjectAliasRef.LockKey()
		} else if ref.SubjectRelationship != nil && ref.SubjectEntity == nil {
			locks[i] = ref.SubjectRelationship.SubjectAliasRef.LockKey()
		} else {
			return fmt.Errorf("evidence must contain exactly one entity or relationship")
		}
	}
	if lockErr := s.db.AcquireTxLocks(ctx, "knowledge_evidence", locks...); lockErr != nil {
		return fmt.Errorf("failed to acquire tx locks: %w", lockErr)
	}
	return nil
}

func (s *KnowledgeGraphService) IngestEvidence(ctx context.Context, event *ent.NormalizedEvent, refs ...ent.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.acquireRefTransactionLocks(ctx, refs); lockErr != nil {
			return fmt.Errorf("failed to acquire tx locks: %w", lockErr)
		}
		builders := make([]*ent.KnowledgeEvidenceCreate, len(refs))
		for i, ref := range refs {
			ra, aliasErr := s.resolveEvidenceSubjectAlias(ctx, ref)
			if aliasErr != nil {
				return fmt.Errorf("set subject alias: %w", aliasErr)
			}
			builders[i] = s.makeEvidenceBuilderFromRef(tx, event.ID, ra.aliasId, ref)
		}
		createEvidence := tx.KnowledgeEvidence.CreateBulk(builders...).
			OnConflict(knowledgeEvidenceUniqueColumns).
			Ignore()
		if createErr := createEvidence.Exec(ctx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}
		return nil
	})
}

func (s *KnowledgeGraphService) IngestSubjectEvidence(ctx context.Context, event *ent.NormalizedEvent, ref ent.KnowledgeEvidenceRef) (*ent.KnowledgeSubjectAlias, error) {
	var alias *ent.KnowledgeSubjectAlias
	return alias, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.acquireRefTransactionLocks(ctx, []ent.KnowledgeEvidenceRef{ref}); lockErr != nil {
			return fmt.Errorf("failed to acquire tx locks: %w", lockErr)
		}

		sa, saErr := s.resolveEvidenceSubjectAlias(ctx, ref)
		if saErr != nil {
			return fmt.Errorf("set subject alias: %w", saErr)
		}

		createEvidence := s.makeEvidenceBuilderFromRef(tx, event.ID, sa.aliasId, ref).
			OnConflict(knowledgeEvidenceUniqueColumns).
			Ignore()
		if createErr := createEvidence.Exec(ctx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}

		subjAlias, subjAliasErr := tx.KnowledgeSubjectAlias.Get(ctx, sa.aliasId)
		if subjAliasErr != nil {
			return fmt.Errorf("load subject alias: %w", subjAliasErr)
		}
		alias = subjAlias.Unwrap()
		return nil
	})
}
