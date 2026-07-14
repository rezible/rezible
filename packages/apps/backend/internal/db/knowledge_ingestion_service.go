package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

type KnowledgeIngestionService struct {
	db rez.Database
}

func NewKnowledgeIngestionService(db rez.Database) *KnowledgeIngestionService {
	return &KnowledgeIngestionService{db: db}
}

var (
	knowledgeEntityUniqueCols       = sql.ConflictColumns(kne.FieldTenantID, kne.FieldKind, kne.FieldReference)
	knowledgeRelationshipUniqueCols = sql.ConflictColumns(knr.FieldTenantID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
	knowledgeSubjectAliasUniqueCols = sql.ConflictColumns(ksa.FieldTenantID, ksa.FieldSubjectKind, ksa.FieldProvider, ksa.FieldProviderSubjectRef)
	knowledgeEvidenceUniqueCols     = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldAliasID, ke.FieldEvidenceKind)
)

func (s *KnowledgeIngestionService) setEntityFromRef(ctx context.Context, er *ent.KnowledgeEntityRef) (uuid.UUID, error) {
	if er == nil {
		return uuid.Nil, fmt.Errorf("nil ref")
	}

	return s.db.Client(ctx).KnowledgeEntity.Create().
		SetKind(er.Kind).
		SetReference(er.Reference).
		SetDisplayName(er.DisplayName).
		SetDescription(er.Description).
		OnConflict(knowledgeEntityUniqueCols).
		UpdateUpdatedAt().
		ID(ctx)
}

func (s *KnowledgeIngestionService) setRelationshipFromRef(ctx context.Context, rr *ent.KnowledgeRelationshipRef) (uuid.UUID, error) {
	if rr == nil {
		return uuid.Nil, fmt.Errorf("nil ref")
	}

	sourceId, sourceErr := s.setEntityFromRef(ctx, &rr.EntityRefs[0])
	if sourceErr != nil {
		return uuid.Nil, fmt.Errorf("source entity: %w", sourceErr)
	}
	targetId, targetErr := s.setEntityFromRef(ctx, &rr.EntityRefs[1])
	if targetErr != nil {
		return uuid.Nil, fmt.Errorf("target entity: %w", targetErr)
	}

	return s.db.Client(ctx).KnowledgeRelationship.Create().
		SetKind(rr.Kind).
		SetSourceEntityID(sourceId).
		SetTargetEntityID(targetId).
		SetDescription(rr.Description).
		OnConflict(knowledgeRelationshipUniqueCols).
		UpdateUpdatedAt().
		ID(ctx)
}

func (s *KnowledgeIngestionService) setSubjectAliasFromProjection(ctx context.Context, sar ent.KnowledgeSubjectAliasRef) (uuid.UUID, error) {
	create := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetProvider(sar.Provider).
		SetProviderSubjectRef(sar.ProviderSubjectRef).
		SetSubjectKind(sar.Kind).
		SetDescription(sar.Description)

	if sar.Kind == ksa.SubjectKindEntity {
		id, setEntityErr := s.setEntityFromRef(ctx, sar.SubjectEntityRef)
		if setEntityErr != nil {
			return uuid.Nil, fmt.Errorf("entity: %w", setEntityErr)
		}
		create.SetEntityID(id)
	} else if sar.Kind == ksa.SubjectKindRelationship {
		id, setRelErr := s.setRelationshipFromRef(ctx, sar.SubjectRelationshipRef)
		if setRelErr != nil {
			return uuid.Nil, fmt.Errorf("relationship: %w", setRelErr)
		}
		create.SetRelationshipID(id)
	} else {
		return uuid.Nil, fmt.Errorf("invalid kind %s", sar.Kind)
	}

	upsert := create.OnConflict(knowledgeSubjectAliasUniqueCols).
		UpdateLastObservedAt()

	return upsert.ID(ctx)
}

func (s *KnowledgeIngestionService) IngestProjectedEvidence(ctx context.Context, event *ent.NormalizedEvent, projected ...ent.KnowledgeEvidenceRef) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		evidenceBuilders := make([]*ent.KnowledgeEvidenceCreate, len(projected))
		for i, projEv := range projected {
			aliasId, setAliasErr := s.setSubjectAliasFromProjection(ctx, projEv.SubjectAliasRef)
			if setAliasErr != nil {
				return fmt.Errorf("set subject alias: %w", setAliasErr)
			}
			evidenceBuilders[i] = tx.KnowledgeEvidence.Create().
				SetEventID(event.ID).
				SetAliasID(aliasId).
				SetEvidenceKind(projEv.Kind).
				SetAssertion(projEv.Assertion).
				SetEffectiveAt(projEv.EffectiveAt).
				SetProperties(projEv.Properties)
		}
		createBulk := tx.KnowledgeEvidence.CreateBulk(evidenceBuilders...).
			OnConflict(knowledgeEvidenceUniqueCols).
			UpdateUpdatedAt()
		if createErr := createBulk.Exec(ctx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}
		return nil
	})
}

func (s *KnowledgeIngestionService) LookupEntityIdByAliasRef(ctx context.Context, ref ent.KnowledgeSubjectAliasRef) (uuid.UUID, error) {
	queryAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ref.Predicate())
	alias, aliasErr := queryAlias.Only(ctx)
	if aliasErr != nil {
		return uuid.Nil, fmt.Errorf("alias: %w", aliasErr)
	}
	return alias.EntityID, nil
}

func (s *KnowledgeIngestionService) IngestDomainEntityEvidence(ctx context.Context, evt *ent.NormalizedEvent, evi ent.KnowledgeEvidenceRef) (uuid.UUID, error) {
	if evi.SubjectAliasRef.SubjectEntityRef == nil {
		return uuid.Nil, fmt.Errorf("missing subject entity ref")
	}
	var entityId uuid.UUID
	return entityId, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		aliasId, setAliasErr := s.setSubjectAliasFromProjection(ctx, evi.SubjectAliasRef)
		if setAliasErr != nil {
			return fmt.Errorf("set subject alias: %w", setAliasErr)
		}
		create := tx.KnowledgeEvidence.Create().
			SetEventID(evt.ID).
			SetAliasID(aliasId).
			SetEvidenceKind(evi.Kind).
			SetAssertion(evi.Assertion).
			SetEffectiveAt(evi.EffectiveAt).
			SetProperties(evi.Properties).
			OnConflict(knowledgeEvidenceUniqueCols).
			UpdateUpdatedAt()
		if saveErr := create.Exec(ctx); saveErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", saveErr)
		}
		alias, aliasErr := tx.KnowledgeSubjectAlias.Get(ctx, aliasId)
		if aliasErr != nil {
			return fmt.Errorf("get knowledge subject alias: %w", aliasErr)
		}
		entityId = alias.EntityID
		return nil
	})
}
