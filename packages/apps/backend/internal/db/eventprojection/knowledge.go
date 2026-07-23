package eventprojection

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
)

var (
	knowledgeEntityUniqueColumns       = sql.ConflictColumns(kne.FieldTenantID, kne.FieldKind, kne.FieldReference)
	knowledgeRelationshipUniqueColumns = sql.ConflictColumns(knr.FieldTenantID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
	knowledgeSubjectAliasUniqueColumns = sql.ConflictColumns(ksa.FieldTenantID, ksa.FieldSubjectKind, ksa.FieldProvider, ksa.FieldProviderSubjectRef)
	knowledgeEvidenceUniqueColumns     = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldAliasID, ke.FieldEvidenceKind, ke.FieldAssertion)
)

func (s *ProjectionService) setEntityFromRef(ctx context.Context, entityRef *ent.KnowledgeEntityRef, applyState bool) (uuid.UUID, error) {
	if entityRef == nil {
		return uuid.Nil, fmt.Errorf("entity reference is required")
	}

	queryEntity := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.Kind(entityRef.Kind), kne.Reference(entityRef.Reference))
	existing, queryErr := queryEntity.Only(ctx)
	if queryErr == nil {
		if !applyState {
			return existing.ID, nil
		}
		updateEntity := s.db.Client(ctx).KnowledgeEntity.UpdateOne(existing)
		if entityRef.DisplayName != "" {
			updateEntity.SetDisplayName(entityRef.DisplayName)
		}
		if entityRef.Description != "" {
			updateEntity.SetDescription(entityRef.Description)
		}
		if entityRef.Properties != nil {
			properties := cloneMap(existing.LiveProperties)
			for key, value := range entityRef.Properties {
				properties[key] = value
			}
			updateEntity.SetLiveProperties(properties)
		}
		updated, updateErr := updateEntity.Save(ctx)
		if updateErr != nil {
			return uuid.Nil, fmt.Errorf("update entity: %w", updateErr)
		}
		return updated.ID, nil
	}
	if !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query entity: %w", queryErr)
	}

	displayName := entityRef.DisplayName
	if displayName == "" {
		displayName = entityRef.Reference
	}
	createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
		SetKind(entityRef.Kind).
		SetReference(entityRef.Reference).
		SetDisplayName(displayName).
		SetDescription(entityRef.Description)
	if entityRef.Properties != nil {
		createEntity.SetLiveProperties(entityRef.Properties)
	}
	upsertEntity := createEntity.OnConflict(knowledgeEntityUniqueColumns).
		Update(func(update *ent.KnowledgeEntityUpsert) {
			if entityRef.DisplayName != "" {
				update.SetDisplayName(entityRef.DisplayName)
			}
			if entityRef.Description != "" {
				update.SetDescription(entityRef.Description)
			}
			if entityRef.Properties != nil {
				update.SetLiveProperties(entityRef.Properties)
			}
			update.SetUpdatedAt(time.Now())
		})
	entityID, createErr := upsertEntity.ID(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create entity: %w", createErr)
	}
	return entityID, nil
}

func (s *ProjectionService) setRelationshipFromRef(ctx context.Context, relationshipRef *ent.KnowledgeRelationshipRef, applyState bool) (uuid.UUID, error) {
	if relationshipRef == nil {
		return uuid.Nil, fmt.Errorf("relationship reference is required")
	}

	sourceID, sourceErr := s.setEntityFromRef(ctx, &relationshipRef.EntityRefs[0], applyState)
	if sourceErr != nil {
		return uuid.Nil, fmt.Errorf("resolve source entity: %w", sourceErr)
	}
	targetID, targetErr := s.setEntityFromRef(ctx, &relationshipRef.EntityRefs[1], applyState)
	if targetErr != nil {
		return uuid.Nil, fmt.Errorf("resolve target entity: %w", targetErr)
	}

	queryRelationship := s.db.Client(ctx).KnowledgeRelationship.Query().Where(
		knr.Kind(relationshipRef.Kind),
		knr.SourceEntityID(sourceID),
		knr.TargetEntityID(targetID),
	)
	existing, queryErr := queryRelationship.Only(ctx)
	if queryErr == nil {
		if !applyState {
			return existing.ID, nil
		}
		updateRelationship := s.db.Client(ctx).KnowledgeRelationship.UpdateOne(existing)
		if relationshipRef.Description != "" {
			updateRelationship.SetDescription(relationshipRef.Description)
		}
		if relationshipRef.Properties != nil {
			properties := cloneMap(existing.Properties)
			for key, value := range relationshipRef.Properties {
				properties[key] = value
			}
			updateRelationship.SetProperties(properties)
		}
		updated, updateErr := updateRelationship.Save(ctx)
		if updateErr != nil {
			return uuid.Nil, fmt.Errorf("update relationship: %w", updateErr)
		}
		return updated.ID, nil
	}
	if !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query relationship: %w", queryErr)
	}

	createRelationship := s.db.Client(ctx).KnowledgeRelationship.Create().
		SetKind(relationshipRef.Kind).
		SetSourceEntityID(sourceID).
		SetTargetEntityID(targetID).
		SetDescription(relationshipRef.Description)
	if relationshipRef.Properties != nil {
		createRelationship.SetProperties(relationshipRef.Properties)
	}
	upsertRelationship := createRelationship.OnConflict(knowledgeRelationshipUniqueColumns).
		Update(func(update *ent.KnowledgeRelationshipUpsert) {
			if relationshipRef.Description != "" {
				update.SetDescription(relationshipRef.Description)
			}
			if relationshipRef.Properties != nil {
				update.SetProperties(relationshipRef.Properties)
			}
			update.SetUpdatedAt(time.Now())
		})
	relationshipID, createErr := upsertRelationship.ID(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create relationship: %w", createErr)
	}
	return relationshipID, nil
}

func (s *ProjectionService) setSubjectAliasFromProjection(ctx context.Context, aliasRef ent.KnowledgeSubjectAliasRef, evidenceKind ke.EvidenceKind, observedAt time.Time) (uuid.UUID, error) {
	queryAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Query().Where(
		ksa.SubjectKindEQ(aliasRef.Kind),
		ksa.Provider(aliasRef.Provider),
		ksa.ProviderSubjectRef(aliasRef.ProviderSubjectRef),
	)
	alias, queryErr := queryAlias.Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query subject alias: %w", queryErr)
	}
	isLatest := alias == nil || !observedAt.Before(alias.LastObservedAt)
	applyState := isLatest && evidenceKind != ke.EvidenceKindDeleted

	var entityID, relationshipID uuid.UUID
	switch aliasRef.Kind {
	case ksa.SubjectKindEntity:
		var entityErr error
		entityID, entityErr = s.setEntityFromRef(ctx, aliasRef.SubjectEntityRef, applyState)
		if entityErr != nil {
			return uuid.Nil, fmt.Errorf("resolve alias entity: %w", entityErr)
		}
	case ksa.SubjectKindRelationship:
		var relationshipErr error
		relationshipID, relationshipErr = s.setRelationshipFromRef(ctx, aliasRef.SubjectRelationshipRef, applyState)
		if relationshipErr != nil {
			return uuid.Nil, fmt.Errorf("resolve alias relationship: %w", relationshipErr)
		}
	default:
		return uuid.Nil, fmt.Errorf("invalid alias subject kind %q", aliasRef.Kind)
	}

	if alias == nil {
		createAlias := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
			SetProvider(aliasRef.Provider).
			SetProviderSubjectRef(aliasRef.ProviderSubjectRef).
			SetSubjectKind(aliasRef.Kind).
			SetDescription(aliasRef.Description).
			SetFirstObservedAt(observedAt).
			SetLastObservedAt(observedAt)
		if evidenceKind == ke.EvidenceKindDeleted {
			createAlias.SetDeletedAt(observedAt)
		}
		if entityID != uuid.Nil {
			createAlias.SetEntityID(entityID)
		}
		if relationshipID != uuid.Nil {
			createAlias.SetRelationshipID(relationshipID)
		}
		createErr := createAlias.OnConflict(knowledgeSubjectAliasUniqueColumns).DoNothing().Exec(ctx)
		if createErr != nil {
			return uuid.Nil, fmt.Errorf("create subject alias: %w", createErr)
		}
		alias, queryErr = queryAlias.Only(ctx)
		if queryErr != nil {
			return uuid.Nil, fmt.Errorf("query created subject alias: %w", queryErr)
		}
	}

	if alias.EntityID != entityID || alias.RelationshipID != relationshipID {
		return uuid.Nil, fmt.Errorf(
			"%w: provider subject %q is already linked to another knowledge subject",
			rez.ErrConflict,
			aliasRef.ProviderSubjectRef,
		)
	}
	if observedAt.Before(alias.FirstObservedAt) {
		if _, updateErr := s.db.Client(ctx).KnowledgeSubjectAlias.Update().
			Where(ksa.ID(alias.ID), ksa.FirstObservedAtGT(observedAt)).
			SetFirstObservedAt(observedAt).
			Save(ctx); updateErr != nil {
			return uuid.Nil, fmt.Errorf("update alias first observation: %w", updateErr)
		}
	}
	if observedAt.After(alias.LastObservedAt) {
		if _, updateErr := s.db.Client(ctx).KnowledgeSubjectAlias.Update().
			Where(ksa.ID(alias.ID), ksa.LastObservedAtLT(observedAt)).
			SetLastObservedAt(observedAt).
			Save(ctx); updateErr != nil {
			return uuid.Nil, fmt.Errorf("update alias last observation: %w", updateErr)
		}
	}
	if isLatest {
		updateAlias := s.db.Client(ctx).KnowledgeSubjectAlias.UpdateOneID(alias.ID)
		updateRequired := false
		if aliasRef.Description != "" && alias.Description != aliasRef.Description {
			updateAlias.SetDescription(aliasRef.Description)
			updateRequired = true
		}
		if evidenceKind == ke.EvidenceKindDeleted {
			updateAlias.SetDeletedAt(observedAt)
			updateRequired = true
		} else if alias.DeletedAt != nil {
			updateAlias.ClearDeletedAt()
			updateRequired = true
		}
		if updateRequired {
			if _, updateErr := updateAlias.Save(ctx); updateErr != nil {
				return uuid.Nil, fmt.Errorf("update subject alias: %w", updateErr)
			}
		}
	}
	return alias.ID, nil
}

func (s *ProjectionService) ingestProjectedEvidence(ctx context.Context, event *ent.NormalizedEvent, evidenceRefs ...ent.KnowledgeEvidenceRef) error {
	if len(evidenceRefs) == 0 {
		return nil
	}
	return s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		builders := make([]*ent.KnowledgeEvidenceCreate, len(evidenceRefs))
		for i, evidenceRef := range evidenceRefs {
			aliasID, aliasErr := s.setSubjectAliasFromProjection(txCtx, evidenceRef.SubjectAliasRef, evidenceRef.Kind, evidenceRef.EffectiveAt)
			if aliasErr != nil {
				return fmt.Errorf("set subject alias: %w", aliasErr)
			}
			builders[i] = tx.KnowledgeEvidence.Create().
				SetEventID(event.ID).
				SetAliasID(aliasID).
				SetEvidenceKind(evidenceRef.Kind).
				SetAssertion(evidenceRef.Assertion).
				SetEffectiveAt(evidenceRef.EffectiveAt).
				SetProperties(evidenceRef.Properties).
				SetSubjectState(knowledgeEvidenceSubjectState(evidenceRef))
		}
		createEvidence := tx.KnowledgeEvidence.CreateBulk(builders...).
			OnConflict(knowledgeEvidenceUniqueColumns).
			UpdateUpdatedAt()
		if createErr := createEvidence.Exec(txCtx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}
		return nil
	})
}

func knowledgeEvidenceSubjectState(evidenceRef ent.KnowledgeEvidenceRef) map[string]any {
	aliasRef := evidenceRef.SubjectAliasRef
	if aliasRef.SubjectEntityRef != nil {
		return map[string]any{
			"subject_kind": "entity",
			"entity":       aliasRef.SubjectEntityRef,
		}
	}
	if aliasRef.SubjectRelationshipRef != nil {
		return map[string]any{
			"subject_kind": "relationship",
			"relationship": aliasRef.SubjectRelationshipRef,
		}
	}
	return map[string]any{}
}

func (s *ProjectionService) ingestDomainEntityEvidence(ctx context.Context, event *ent.NormalizedEvent, evidenceRef ent.KnowledgeEvidenceRef) (uuid.UUID, error) {
	if evidenceRef.SubjectAliasRef.SubjectEntityRef == nil {
		return uuid.Nil, fmt.Errorf("evidence subject entity is required")
	}

	var entityID uuid.UUID
	return entityID, s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		aliasID, aliasErr := s.setSubjectAliasFromProjection(txCtx, evidenceRef.SubjectAliasRef, evidenceRef.Kind, evidenceRef.EffectiveAt)
		if aliasErr != nil {
			return fmt.Errorf("set subject alias: %w", aliasErr)
		}
		createEvidence := tx.KnowledgeEvidence.Create().
			SetEventID(event.ID).
			SetAliasID(aliasID).
			SetEvidenceKind(evidenceRef.Kind).
			SetAssertion(evidenceRef.Assertion).
			SetEffectiveAt(evidenceRef.EffectiveAt).
			SetProperties(evidenceRef.Properties).
			SetSubjectState(knowledgeEvidenceSubjectState(evidenceRef)).
			OnConflict(knowledgeEvidenceUniqueColumns).
			UpdateUpdatedAt()
		if createErr := createEvidence.Exec(txCtx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}

		alias, getErr := tx.KnowledgeSubjectAlias.Get(txCtx, aliasID)
		if getErr != nil {
			return fmt.Errorf("get subject alias: %w", getErr)
		}
		entityID = alias.EntityID
		return nil
	})
}

func cloneMap(values map[string]any) map[string]any {
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
