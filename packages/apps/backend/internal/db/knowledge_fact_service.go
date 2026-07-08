package db

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeFactService struct {
	db rez.Database
}

func NewKnowledgeFactService(db rez.Database) *KnowledgeFactService {
	return &KnowledgeFactService{db: db}
}

func (s *KnowledgeFactService) GetEntity(ctx context.Context, p predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().Where(p)
	return query.Only(ctx)
}

//func (s *KnowledgeService) SetEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityMutation)) (*ent.KnowledgeEntity, error) {
//	var mutator ent.EntityMutator[*ent.KnowledgeEntity, *ent.KnowledgeEntityMutation]
//	if id == uuid.Nil {
//		create := s.db.Client(ctx).KnowledgeEntity.Create()
//		mutator = create
//	} else {
//		mutator = s.db.Client(ctx).KnowledgeEntity.UpdateOneID(id)
//	}
//	setFn(mutator.Mutation())
//	return mutator.Save(ctx)
//}

func (s *KnowledgeFactService) GetRelationship(ctx context.Context, pred predicate.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().Where(pred)
	return query.Only(ctx)
}

func (s *KnowledgeFactService) getRelationshipFromRef(ctx context.Context, rr *ent.KnowledgeRelationshipRef) (*ent.KnowledgeRelationship, error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(rr.Predicate())
	return query.Only(ctx)
}

func (s *KnowledgeFactService) GetSubjectAlias(ctx context.Context, p predicate.KnowledgeSubjectAlias) (*ent.KnowledgeSubjectAlias, error) {
	return s.db.Client(ctx).KnowledgeSubjectAlias.Query().Where(p).Only(ctx)
}

var (
	knowledgeEntityUniqueCols       = sql.ConflictColumns(kne.FieldTenantID, kne.FieldKind, kne.FieldReference)
	knowledgeRelationshipUniqueCols = sql.ConflictColumns(knr.FieldTenantID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
	knowledgeSubjectAliasUniqueCols = sql.ConflictColumns(ksa.FieldTenantID, ksa.FieldSubjectKind, ksa.FieldProvider, ksa.FieldProviderSubjectRef)
	knowledgeEvidenceUniqueCols     = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldAliasID, ke.FieldEvidenceKind)
)

func (s *KnowledgeFactService) setEntityFromRef(ctx context.Context, er ent.KnowledgeEntityRef) (uuid.UUID, error) {
	upsert := s.db.Client(ctx).KnowledgeEntity.Create().
		SetKind(er.Kind).
		SetReference(er.Reference).
		SetDisplayName(er.DisplayName).
		SetDescription(er.Description).
		OnConflict(knowledgeEntityUniqueCols).
		UpdateUpdatedAt()
	return upsert.ID(ctx)
}

func (s *KnowledgeFactService) setRelationshipFromRef(ctx context.Context, rr ent.KnowledgeRelationshipRef) (uuid.UUID, error) {
	sourceId, sourceErr := s.setEntityFromRef(ctx, rr.EntityRefs[0])
	if sourceErr != nil {
		return uuid.Nil, fmt.Errorf("source entity: %w", sourceErr)
	}
	targetId, targetErr := s.setEntityFromRef(ctx, rr.EntityRefs[1])
	if targetErr != nil {
		return uuid.Nil, fmt.Errorf("target entity: %w", targetErr)
	}

	upsert := s.db.Client(ctx).KnowledgeRelationship.Create().
		SetKind(rr.Kind).
		SetSourceEntityID(sourceId).
		SetTargetEntityID(targetId).
		SetDescription(rr.Description).
		OnConflict(knowledgeRelationshipUniqueCols).
		UpdateUpdatedAt()
	return upsert.ID(ctx)
}

func (s *KnowledgeFactService) setSubjectAliasFromProjection(ctx context.Context, proj rez.ProjectedKnowledgeEvidenceSubjectAlias) (uuid.UUID, error) {
	sar := proj.AliasRef
	create := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetSubjectKind(sar.Kind).
		SetProvider(proj.AliasRef.Provider).
		SetProviderSubjectRef(proj.AliasRef.ProviderSubjectRef).
		SetDescription(proj.Description)

	if sar.Kind == ksa.SubjectKindEntity && proj.SubjectEntityRef != nil {
		id, setEntityErr := s.setEntityFromRef(ctx, *proj.SubjectEntityRef)
		if setEntityErr != nil {
			return uuid.Nil, fmt.Errorf("set entity: %w", setEntityErr)
		}
		create.SetEntityID(id)
	} else if sar.Kind == ksa.SubjectKindRelationship && proj.SubjectRelationshipRef != nil {
		id, setRelErr := s.setRelationshipFromRef(ctx, *proj.SubjectRelationshipRef)
		if setRelErr != nil {
			return uuid.Nil, fmt.Errorf("set relationship: %w", setRelErr)
		}
		create.SetRelationshipID(id)
	} else {
		return uuid.Nil, fmt.Errorf("invalid subject %s (missing ref?)", sar.Kind)
	}

	upsert := create.OnConflict(knowledgeSubjectAliasUniqueCols).
		SetUpdatedAt(time.Now()).
		UpdateNewValues()

	return upsert.ID(ctx)
}

func (s *KnowledgeFactService) IngestProjectedEventEvidence(ctx context.Context, event *ent.NormalizedEvent, projected []rez.ProjectedKnowledgeEvidence) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		evidenceBuilders := make([]*ent.KnowledgeEvidenceCreate, len(projected))
		for i, projEv := range projected {
			aliasId, setAliasErr := s.setSubjectAliasFromProjection(ctx, projEv.SubjectAlias)
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

func (s *KnowledgeFactService) lookupAliasesFromRefs(ctx context.Context, refs ...ent.KnowledgeSubjectAliasRef) (ent.KnowledgeSubjectAliasSlice, error) {
	preds := make([]predicate.KnowledgeSubjectAlias, len(refs))
	for i, ref := range refs {
		preds[i] = ref.Predicate()
	}
	return s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.Or(preds...)).
		All(ctx)
}

func (s *KnowledgeFactService) LookupEntityIdByAliasRefs(ctx context.Context, refs ...ent.KnowledgeSubjectAliasRef) (uuid.UUID, error) {
	aliases, aliasesErr := s.lookupAliasesFromRefs(ctx, refs...)
	if aliasesErr != nil {
		return uuid.Nil, fmt.Errorf("query aliases: %w", aliasesErr)
	}
	var entityId uuid.UUID
	for _, a := range aliases {
		if entityId == uuid.Nil {
			entityId = a.EntityID
			continue
		}
		if entityId != a.EntityID {
			return uuid.Nil, fmt.Errorf("projected aliases resolve to different entities: %s -> %s (expected %s)",
				a.ProviderSubjectRef, a.EntityID, entityId)
		}
	}
	return entityId, nil
}
