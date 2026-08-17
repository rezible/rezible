package db

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

var (
	knowledgeRelationshipUniqueColumns = sql.ConflictColumns(knr.FieldTenantID, knr.FieldKind, knr.FieldSubkind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
	knowledgeSubjectAliasUniqueColumns = sql.ConflictColumns(ksa.FieldTenantID, ksa.FieldSubjectKind, ksa.FieldProvider, ksa.FieldProviderSource, ksa.FieldProviderSubjectRef)
	knowledgeEvidenceUniqueColumns     = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldSubjectAliasID)
)

func (s *KnowledgeGraphService) createSubjectAlias(ctx context.Context, ref ent.KnowledgeAliasRef) *ent.KnowledgeSubjectAliasCreate {
	return s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetProvider(ref.Provider).
		SetProviderSource(ref.ProviderSource).
		SetProviderSubjectRef(ref.ProviderSubjectRef)
}

func (s *KnowledgeGraphService) setEntityFromRef(ctx context.Context, ref ent.KnowledgeEntityRef) (*ent.KnowledgeSubjectAlias, error) {
	queryAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ref.Alias.SubjectPredicate(ksa.SubjectKindEntity)).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query entity alias: %w", queryErr)
	} else if alias != nil {
		entity, entityErr := alias.Edges.EntityOrErr()
		if entityErr != nil {
			return nil, fmt.Errorf("load alias entity: %w", entityErr)
		}
		if entity.Kind != ref.Kind || entity.Subkind != ref.Subkind {
			return nil, fmt.Errorf("%w: alias identifies %q/%q, evidence expects %q/%q", rez.ErrConflict, entity.Kind, entity.Subkind, ref.Kind, ref.Subkind)
		}
		return alias, nil
	}

	createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
		SetKind(ref.Kind).
		SetSubkind(ref.Subkind)
	createdEntity, createEntityErr := createEntity.Save(ctx)
	if createEntityErr != nil {
		return nil, fmt.Errorf("create knowledge entity: %w", createEntityErr)
	}

	createAlias := s.createSubjectAlias(ctx, ref.Alias).
		SetSubjectKind(ksa.SubjectKindEntity).
		SetEntityID(createdEntity.ID)
	createdAlias, createAliasErr := createAlias.Save(ctx)
	if createAliasErr != nil {
		return nil, fmt.Errorf("create entity alias: %w", createAliasErr)
	}
	return createdAlias, nil
}

func (s *KnowledgeGraphService) setRelationshipFromRef(ctx context.Context, ref ent.KnowledgeRelationshipRef) (*ent.KnowledgeSubjectAlias, error) {
	sourceAlias, sourceErr := s.setEntityFromRef(ctx, ref.Source)
	if sourceErr != nil {
		return nil, fmt.Errorf("resolve source entity: %w", sourceErr)
	} else if sourceAlias.EntityID == nil {
		return nil, fmt.Errorf("nil source alias entity id")
	}
	sourceId := *sourceAlias.EntityID

	targetAlias, targetErr := s.setEntityFromRef(ctx, ref.Target)
	if targetErr != nil {
		return nil, fmt.Errorf("resolve target entity: %w", targetErr)
	} else if targetAlias.EntityID == nil {
		return nil, fmt.Errorf("nil target alias entity id")
	}
	targetId := *targetAlias.EntityID

	queryExisting := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ref.Alias.SubjectPredicate(ksa.SubjectKindRelationship)).
		WithRelationship()
	existingAlias, queryExistingErr := queryExisting.Only(ctx)
	if queryExistingErr != nil && !ent.IsNotFound(queryExistingErr) {
		return nil, fmt.Errorf("query existing relationship alias: %w", queryExistingErr)
	} else if existingAlias != nil {
		rel, edgeErr := existingAlias.Edges.RelationshipOrErr()
		if edgeErr != nil {
			return nil, fmt.Errorf("load aliased relationship: %w", edgeErr)
		}
		if rel.Kind != ref.Kind || rel.Subkind != ref.Subkind || rel.SourceEntityID != sourceId || rel.TargetEntityID != targetId {
			return nil, fmt.Errorf("%w: relationship alias identifies different topology", rez.ErrConflict)
		}
		return existingAlias, nil
	}

	upsertRel := s.db.Client(ctx).KnowledgeRelationship.Create().
		SetKind(ref.Kind).
		SetSubkind(ref.Subkind).
		SetSourceEntityID(sourceId).
		SetTargetEntityID(targetId).
		OnConflict(knowledgeRelationshipUniqueColumns).
		SetUpdatedAt(time.Now().UTC())
	relId, upsertRelErr := upsertRel.ID(ctx)
	if upsertRelErr != nil {
		return nil, fmt.Errorf("upsert relationship: %w", upsertRelErr)
	}
	createAlias := s.createSubjectAlias(ctx, ref.Alias).
		SetSubjectKind(ksa.SubjectKindRelationship).
		SetRelationshipID(relId)
	createdAlias, createAliasErr := createAlias.Save(ctx)
	if createAliasErr != nil {
		return nil, fmt.Errorf("create alias: %w", createAliasErr)
	}
	return createdAlias, nil
}

func (s *KnowledgeGraphService) setEvidenceSubjectAlias(ctx context.Context, ref ent.KnowledgeEvidenceRef) (*ent.KnowledgeSubjectAlias, error) {
	switch {
	case ref.SubjectEntity != nil && ref.SubjectRelationship == nil:
		return s.setEntityFromRef(ctx, *ref.SubjectEntity)
	case ref.SubjectRelationship != nil && ref.SubjectEntity == nil:
		return s.setRelationshipFromRef(ctx, *ref.SubjectRelationship)
	default:
		return nil, fmt.Errorf("evidence must contain exactly one entity or relationship")
	}
}

func (s *KnowledgeGraphService) getRefTransactionLocks(refs []ent.KnowledgeEvidenceRef) ([]string, error) {
	locks := make([]string, len(refs))
	for i, ref := range refs {
		if ref.SubjectEntity != nil && ref.SubjectRelationship == nil {
			locks[i] = ref.SubjectEntity.Alias.LockKey(ksa.SubjectKindEntity)
		} else if ref.SubjectRelationship != nil && ref.SubjectEntity == nil {
			locks[i] = ref.SubjectRelationship.Alias.LockKey(ksa.SubjectKindRelationship)
		} else {
			return nil, fmt.Errorf("evidence must contain exactly one entity or relationship")
		}
	}
	return locks, nil
}

func (s *KnowledgeGraphService) IngestEntityEvidence(ctx context.Context, event *ent.NormalizedEvent, ref ent.KnowledgeEvidenceRef) (*ent.KnowledgeEntity, error) {
	locks, locksErr := s.getRefTransactionLocks([]ent.KnowledgeEvidenceRef{ref})
	if locksErr != nil {
		return nil, fmt.Errorf("make transaction locks: %w", locksErr)
	}

	var entity *ent.KnowledgeEntity
	return entity, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "knowledge_subject_alias", locks...); lockErr != nil {
			return fmt.Errorf("failed to acquire tx locks: %w", lockErr)
		}
		alias, aliasErr := s.setEvidenceSubjectAlias(ctx, ref)
		if aliasErr != nil {
			return fmt.Errorf("set subject alias: %w", aliasErr)
		}
		createEvidence := tx.KnowledgeEvidence.Create().
			SetEventID(event.ID).
			SetSubjectAliasID(alias.ID).
			SetKind(ref.Kind).
			SetAssertion(ref.Assertion).
			SetEffectiveAt(ref.EffectiveAt).
			SetSubjectState(ref.SubjectState).
			OnConflict(knowledgeEvidenceUniqueColumns).
			Ignore()
		if createErr := createEvidence.Exec(ctx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}
		aliasEntity, aliasEntityErr := alias.QueryEntity().Only(ctx)
		if aliasEntityErr != nil {
			return fmt.Errorf("load alias entity: %w", aliasEntityErr)
		}
		entity = aliasEntity.Unwrap()
		return nil
	})
}

func (s *KnowledgeGraphService) IngestEvidenceBulk(ctx context.Context, event *ent.NormalizedEvent, refs ...ent.KnowledgeEvidenceRef) (ent.KnowledgeSubjectAliasSlice, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	locks, locksErr := s.getRefTransactionLocks(refs)
	if locksErr != nil {
		return nil, fmt.Errorf("make transaction locks: %w", locksErr)
	}
	aliases := make(ent.KnowledgeSubjectAliasSlice, len(refs))
	builders := make([]*ent.KnowledgeEvidenceCreate, len(refs))
	return aliases, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "knowledge_subject_alias", locks...); lockErr != nil {
			return fmt.Errorf("failed to acquire tx locks: %w", lockErr)
		}
		for i, ref := range refs {
			alias, aliasErr := s.setEvidenceSubjectAlias(ctx, ref)
			if aliasErr != nil {
				return fmt.Errorf("set subject alias: %w", aliasErr)
			}
			aliases[i] = alias.Unwrap()
			builders[i] = tx.KnowledgeEvidence.Create().
				SetEventID(event.ID).
				SetSubjectAliasID(alias.ID).
				SetKind(ref.Kind).
				SetAssertion(ref.Assertion).
				SetEffectiveAt(ref.EffectiveAt).
				SetSubjectState(ref.SubjectState)
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
