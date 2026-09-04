package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"

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

func (s *KnowledgeGraphService) makeAliasQuery(ctx context.Context, ref ent.ProviderResourceRef, preds ...predicate.KnowledgeSubjectAlias) *ent.KnowledgeSubjectAliasQuery {
	refPred := ksa.And(
		ksa.Provider(ref.Provider),
		ksa.ProviderNamespace(ref.ProviderNamespace),
		ksa.ProviderResourceRef(ref.ResourceRef),
	)
	return s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(preds...).Where(refPred)
}

var knowledgeSubjectAliasUniqueColumns = sql.ConflictColumns(
	ksa.FieldTenantID,
	ksa.FieldProvider,
	ksa.FieldProviderNamespace,
	ksa.FieldProviderResourceRef,
)

func (s *KnowledgeGraphService) ensureAlias(ctx context.Context, ref rez.ProviderResourceRef, subjectKind ksa.SubjectKind, entityID, relationshipID uuid.UUID) (*ent.KnowledgeSubjectAlias, error) {
	query := s.makeAliasQuery(ctx, ref)
	alias, queryErr := query.Only(ctx)
	if alias != nil {
		if alias.SubjectKind != subjectKind ||
			(entityID != uuid.Nil && (alias.EntityID == nil || *alias.EntityID != entityID)) ||
			(relationshipID != uuid.Nil && (alias.RelationshipID == nil || *alias.RelationshipID != relationshipID)) {
			return nil, fmt.Errorf("%w: provider resource is already mapped to another knowledge subject", rez.ErrConflict)
		}
		return alias, nil
	}
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("load subject alias: %w", queryErr)
	}
	create := s.db.Client(ctx).KnowledgeSubjectAlias.Create().
		SetProvider(ref.Provider).
		SetProviderNamespace(ref.ProviderNamespace).
		SetProviderResourceRef(ref.ResourceRef).
		SetSubjectKind(subjectKind)
	if entityID != uuid.Nil {
		create.SetEntityID(entityID)
	} else {
		create.SetRelationshipID(relationshipID)
	}
	createAlias := create.OnConflict(knowledgeSubjectAliasUniqueColumns).
		Ignore()
	_, queryErr = createAlias.ID(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("create subject alias: %w", queryErr)
	}
	alias, queryErr = query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load subject alias after create: %w", queryErr)
	}
	return alias, nil
}

func (s *KnowledgeGraphService) lookupExistingEntityAlias(ctx context.Context, ref ent.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	if refErr := ref.ProviderResourceRef.Validate(); refErr != nil {
		return nil, refErr
	}
	queryAlias := s.makeAliasQuery(ctx, ref.ProviderResourceRef).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	if queryErr != nil || alias == nil {
		return nil, fmt.Errorf("query entity alias: %w", queryErr)
	}
	if alias.SubjectKind != ksa.SubjectKindEntity {
		return nil, fmt.Errorf("%w: resource already identifies a relationship", rez.ErrConflict)
	}
	entity, entityErr := alias.Edges.EntityOrErr()
	if entityErr != nil {
		return nil, fmt.Errorf("load alias entity: %w", entityErr)
	}
	if entity.Category != ref.Category || entity.Kind != ref.Kind {
		return nil, fmt.Errorf("%w: alias identifies %q/%q, evidence expects %q/%q", rez.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entity.ID}, nil
}

func (s *KnowledgeGraphService) resolveEntityFromRef(ctx context.Context, ref ent.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	if refErr := ref.ProviderResourceRef.Validate(); refErr != nil {
		return nil, fmt.Errorf("entity resource ref: %w", refErr)
	}
	existingAlias, lookupAliasErr := s.lookupExistingEntityAlias(ctx, ref)
	if lookupAliasErr != nil && !ent.IsNotFound(lookupAliasErr) {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupAliasErr)
	} else if existingAlias != nil {
		return existingAlias, nil
	}

	createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
		SetCategory(ref.Category).
		SetKind(ref.Kind)
	entity, createErr := createEntity.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create entity: %w", createErr)
	}
	alias, aliasErr := s.ensureAlias(ctx, ref.ProviderResourceRef, ksa.SubjectKindEntity, entity.ID, uuid.Nil)
	if aliasErr != nil {
		return nil, fmt.Errorf("ensure entity alias: %w", aliasErr)
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entity.ID}, nil
}

func (s *KnowledgeGraphService) lookupExistingRelationshipAlias(ctx context.Context, ref ent.KnowledgeRelationshipRef, sourceId, targetId uuid.UUID) (*resolvedSubjectAlias, error) {
	queryRelationshipAlias := s.makeAliasQuery(ctx, ref.ProviderResourceRef).
		WithRelationship()
	alias, queryErr := queryRelationshipAlias.Only(ctx)
	if alias == nil || ent.IsNotFound(queryErr) {
		return nil, nil
	}
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query relationship alias: %w", queryErr)
	}
	if alias.SubjectKind != ksa.SubjectKindRelationship {
		return nil, fmt.Errorf("%w: resource already identifies an entity", rez.ErrConflict)
	}
	rel, relationshipErr := alias.Edges.RelationshipOrErr()
	if relationshipErr != nil {
		return nil, fmt.Errorf("load aliased relationship: %w", relationshipErr)
	}
	if rel.Predicate != ref.Predicate || rel.SourceEntityID != sourceId || rel.TargetEntityID != targetId {
		return nil, fmt.Errorf("%w: relationship alias identifies different topology", rez.ErrConflict)
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: rel.ID}, nil
}

var knowledgeRelationshipUniqueColumns = sql.ConflictColumns(
	knr.FieldTenantID,
	knr.FieldPredicate,
	knr.FieldSourceEntityID,
	knr.FieldTargetEntityID,
)

var errNoExistingEndpointEntityAlias = fmt.Errorf("endpoint entity does not exist")

func (s *KnowledgeGraphService) resolveRelationshipFromRef(ctx context.Context, ref ent.KnowledgeRelationshipRef) (*resolvedSubjectAlias, error) {
	if refErr := ref.ProviderResourceRef.Validate(); refErr != nil {
		return nil, fmt.Errorf("validate resource ref: %w", refErr)
	}

	source, sourceErr := s.lookupExistingEntityAlias(ctx, ref.Source)
	if sourceErr != nil {
		if ent.IsNotFound(sourceErr) {
			sourceErr = errNoExistingEndpointEntityAlias
		}
		return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
	}
	target, targetErr := s.lookupExistingEntityAlias(ctx, ref.Target)
	if targetErr != nil {
		if ent.IsNotFound(targetErr) {
			targetErr = errNoExistingEndpointEntityAlias
		}
		return nil, fmt.Errorf("resolve target: %w", targetErr)
	}

	existing, lookupExistingErr := s.lookupExistingRelationshipAlias(ctx, ref, source.subjectId, target.subjectId)
	if lookupExistingErr != nil {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupExistingErr)
	} else if existing != nil {
		return existing, nil
	}

	create := s.db.Client(ctx).KnowledgeRelationship.Create().
		SetPredicate(ref.Predicate).
		SetSourceEntityID(source.subjectId).
		SetTargetEntityID(target.subjectId).
		OnConflict(knowledgeRelationshipUniqueColumns).
		Ignore()
	relationshipID, upsertErr := create.ID(ctx)
	if upsertErr != nil {
		return nil, fmt.Errorf("create relationship: %w", upsertErr)
	}

	alias, aliasErr := s.ensureAlias(ctx, ref.ProviderResourceRef, ksa.SubjectKindRelationship, uuid.Nil, relationshipID)
	if aliasErr != nil {
		return nil, aliasErr
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: relationshipID}, nil
}

var knowledgeEvidenceUniqueColumns = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldSubjectAliasID)

func (s *KnowledgeGraphService) ingestEvidenceRefs(ctx context.Context, eventID uuid.UUID, kind ksa.SubjectKind, refs []ent.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	resolveAlias := func(ctx context.Context, ref ent.KnowledgeEvidenceRef) (*resolvedSubjectAlias, error) {
		return s.resolveEntityFromRef(ctx, *ref.SubjectEntity)
	}
	if kind == ksa.SubjectKindRelationship {
		resolveAlias = func(ctx context.Context, ref ent.KnowledgeEvidenceRef) (*resolvedSubjectAlias, error) {
			return s.resolveRelationshipFromRef(ctx, *ref.SubjectRelationship)
		}
	}
	evClient := s.db.Client(ctx).KnowledgeEvidence
	builders := make([]*ent.KnowledgeEvidenceCreate, 0, len(refs))
	for _, ref := range refs {
		resolved, aliasErr := resolveAlias(ctx, ref)
		if aliasErr != nil {
			return fmt.Errorf("resolve subject alias: %w", aliasErr)
		}
		builders = append(builders, evClient.Create().
			SetEventID(eventID).
			SetSubjectAliasID(resolved.aliasId).
			SetKind(ref.Kind).
			SetAssertion(ref.Assertion).
			SetEffectiveAt(ref.EffectiveAt).
			SetSubjectState(ref.SubjectState))
	}
	createEvidence := evClient.CreateBulk(builders...).
		OnConflict(knowledgeEvidenceUniqueColumns).
		Ignore()
	if createErr := createEvidence.Exec(ctx); createErr != nil {
		return fmt.Errorf("create knowledge evidence: %w", createErr)
	}
	return nil
}

func (s *KnowledgeGraphService) IngestEvidence(ctx context.Context, event *ent.NormalizedEvent, refs ...ent.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	entityRefs := make([]ent.KnowledgeEvidenceRef, 0, len(refs))
	relationshipRefs := make([]ent.KnowledgeEvidenceRef, 0, len(refs))
	for _, ref := range refs {
		switch {
		case ref.SubjectEntity != nil && ref.SubjectRelationship == nil:
			entityRefs = append(entityRefs, ref)
		case ref.SubjectRelationship != nil && ref.SubjectEntity == nil:
			relationshipRefs = append(relationshipRefs, ref)
		default:
			return fmt.Errorf("evidence must contain exactly one entity or relationship")
		}
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if ingestErr := s.ingestEvidenceRefs(ctx, event.ID, ksa.SubjectKindEntity, entityRefs); ingestErr != nil {
			return fmt.Errorf("entities: %w", ingestErr)
		}
		if ingestErr := s.ingestEvidenceRefs(ctx, event.ID, ksa.SubjectKindRelationship, relationshipRefs); ingestErr != nil {
			return fmt.Errorf("relationships: %w", ingestErr)
		}
		return nil
	})
}
