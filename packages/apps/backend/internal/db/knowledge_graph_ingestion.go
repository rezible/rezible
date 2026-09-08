package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

type resolvedSubjectAlias struct {
	aliasId   uuid.UUID
	subjectId uuid.UUID
}

func (s *KnowledgeGraphService) makeAliasQuery(ctx context.Context, ref rez.ProviderResourceRef, preds ...predicate.KnowledgeSubjectAlias) *ent.KnowledgeSubjectAliasQuery {
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

func (s *KnowledgeGraphService) lookupExistingEntityAlias(ctx context.Context, ref rez.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	if refErr := ref.ProviderResourceRef.Validate(); refErr != nil {
		return nil, fmt.Errorf("resource ref: %w", refErr)
	}
	queryAlias := s.makeAliasQuery(ctx, ref.ProviderResourceRef).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	if ent.IsNotFound(queryErr) {
		return nil, nil
	}
	if queryErr != nil {
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

func (s *KnowledgeGraphService) lookupLinkedEntity(ctx context.Context, values map[string]string) (*ent.KnowledgeEntity, error) {
	if len(values) == 0 {
		return nil, nil
	}

	preds := make([]predicate.KnowledgeEntityLinkingAttribute, 0, len(values))
	for attr, val := range values {
		preds = append(preds, kela.And(kela.Attribute(attr), kela.Value(val)))
	}
	lookupAttrsQuery := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(kela.Or(preds...)).
		WithEntity()
	linkingAttributes, queryErr := lookupAttrsQuery.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query entity linking attributes: %w", queryErr)
	}

	var linked *ent.KnowledgeEntity
	for _, linkingAttribute := range linkingAttributes {
		entity, entityErr := linkingAttribute.Edges.EntityOrErr()
		if entityErr != nil {
			return nil, fmt.Errorf("load linked entity: %w", entityErr)
		}
		if linked == nil {
			linked = entity
			continue
		}
		if linked.ID != entity.ID {
			return nil, fmt.Errorf("%w: linking attributes identify multiple entities", rez.ErrConflict)
		}
	}
	return linked, nil
}

var knowledgeEntityLinkingAttributeUniqueColumns = sql.ConflictColumns(kela.FieldTenantID, kela.FieldAttribute, kela.FieldValue)

func (s *KnowledgeGraphService) ensureEntityLinkingAttributes(ctx context.Context, entityID uuid.UUID, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	attrClient := s.db.Client(ctx).KnowledgeEntityLinkingAttribute
	builders := make([]*ent.KnowledgeEntityLinkingAttributeCreate, 0, len(values))
	for attribute, value := range values {
		builders = append(builders, attrClient.Create().
			SetEntityID(entityID).
			SetAttribute(attribute).
			SetValue(value))
	}
	upsertAttrs := attrClient.CreateBulk(builders...).
		OnConflict(knowledgeEntityLinkingAttributeUniqueColumns).
		Ignore()
	if upsertErr := upsertAttrs.Exec(ctx); upsertErr != nil {
		return fmt.Errorf("save entity linking attributes: %w", upsertErr)
	}
	linked, queryErr := s.lookupLinkedEntity(ctx, values)
	if queryErr != nil {
		return queryErr
	}
	if linked != nil && linked.ID != entityID {
		return fmt.Errorf("%w: linking attribute is already mapped to another entity", rez.ErrConflict)
	}
	return nil
}

func getKnowledgeEntityLinkingValues(ref rez.KnowledgeEntityRef) (map[string]string, error) {
	if ref.LinkingAttributes == nil {
		return nil, nil
	}
	values := ref.LinkingAttributes.Values()
	for attribute, value := range values {
		if strings.TrimSpace(attribute) == "" || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%w: linking attribute and value must be non-empty", rez.ErrInvalidInput)
		}
	}
	return values, nil
}

func (s *KnowledgeGraphService) resolveEntityFromRef(ctx context.Context, ref rez.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	linkingValues, valuesErr := getKnowledgeEntityLinkingValues(ref)
	if valuesErr != nil {
		return nil, valuesErr
	}

	existingAlias, lookupAliasErr := s.lookupExistingEntityAlias(ctx, ref)
	if lookupAliasErr != nil {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupAliasErr)
	}

	var resolvedAlias *resolvedSubjectAlias
	if existingAlias != nil {
		linked, linkedErr := s.lookupLinkedEntity(ctx, linkingValues)
		if linkedErr != nil {
			return nil, linkedErr
		} else if linked != nil {
			if linked.ID != existingAlias.subjectId {
				return nil, fmt.Errorf("%w: linking attribute identifies another entity", rez.ErrConflict)
			}
		}
		resolvedAlias = existingAlias
	} else {
		entity, linkedErr := s.lookupLinkedEntity(ctx, linkingValues)
		if linkedErr != nil {
			return nil, linkedErr
		} else if entity != nil {
			if entity.Category != ref.Category || entity.Kind != ref.Kind {
				return nil, fmt.Errorf("%w: linking attribute identifies %q/%q, evidence expects %q/%q",
					rez.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
			}
		} else {
			createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
				SetCategory(ref.Category).
				SetKind(ref.Kind)
			created, createErr := createEntity.Save(ctx)
			if createErr != nil {
				return nil, fmt.Errorf("create entity: %w", createErr)
			}
			entity = created
		}
		alias, aliasErr := s.ensureAlias(ctx, ref.ProviderResourceRef, ksa.SubjectKindEntity, entity.ID, uuid.Nil)
		if aliasErr != nil {
			return nil, fmt.Errorf("ensure entity alias: %w", aliasErr)
		}
		resolvedAlias = &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entity.ID}
	}

	if linkingErr := s.ensureEntityLinkingAttributes(ctx, resolvedAlias.subjectId, linkingValues); linkingErr != nil {
		return nil, linkingErr
	}

	return resolvedAlias, nil
}

func (s *KnowledgeGraphService) lookupExistingRelationshipAlias(ctx context.Context, ref rez.KnowledgeRelationshipRef, sourceId, targetId uuid.UUID) (*resolvedSubjectAlias, error) {
	queryRelationshipAlias := s.makeAliasQuery(ctx, ref.ProviderResourceRef).
		WithRelationship()
	alias, queryErr := queryRelationshipAlias.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil, nil
		}
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

func (s *KnowledgeGraphService) resolveRelationshipFromRef(ctx context.Context, ref rez.KnowledgeRelationshipRef, source, target *resolvedSubjectAlias) (*resolvedSubjectAlias, error) {
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

func (s *KnowledgeGraphService) ingestEvidenceRefs(ctx context.Context, eventID uuid.UUID, kind ksa.SubjectKind, refs []rez.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	resolveAlias := func(ctx context.Context, ref rez.KnowledgeEvidenceRef) (*resolvedSubjectAlias, error) {
		return s.resolveEntityFromRef(ctx, *ref.SubjectEntity)
	}
	if kind == ksa.SubjectKindRelationship {
		resolveAlias = func(ctx context.Context, ref rez.KnowledgeEvidenceRef) (*resolvedSubjectAlias, error) {
			rr := ref.SubjectRelationship
			if refErr := rr.ProviderResourceRef.Validate(); refErr != nil {
				return nil, fmt.Errorf("validate resource ref: %w", refErr)
			}

			source, sourceErr := s.lookupExistingEntityAlias(ctx, rr.Source)
			if sourceErr != nil || source == nil {
				return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
			}
			target, targetErr := s.lookupExistingEntityAlias(ctx, rr.Target)
			if targetErr != nil || target == nil {
				return nil, fmt.Errorf("resolve target: %w", errNoExistingEndpointEntityAlias)
			}

			return s.resolveRelationshipFromRef(ctx, *ref.SubjectRelationship, source, target)
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

const knowledgeEntityIdentityLockNamespace = "knowledge_entity_identity"

type knowledgeEntityIdentityLockKeys struct {
	keys []string
}

func (k *knowledgeEntityIdentityLockKeys) makeKey(parts ...string) string {
	encoded, _ := json.Marshal(parts)
	return string(encoded)
}

func (k *knowledgeEntityIdentityLockKeys) addEntity(ref rez.KnowledgeEntityRef) error {
	values, valuesErr := getKnowledgeEntityLinkingValues(ref)
	if valuesErr != nil {
		return valuesErr
	}
	rr := ref.ProviderResourceRef
	k.keys = append(k.keys, k.makeKey("alias", rr.Provider, rr.ProviderNamespace, rr.ResourceRef))
	for attribute, value := range values {
		k.keys = append(k.keys, k.makeKey("attribute", attribute, value))
	}
	return nil
}

func (k *knowledgeEntityIdentityLockKeys) collect(refs []rez.KnowledgeEvidenceRef) error {
	for _, ref := range refs {
		switch {
		case ref.SubjectEntity != nil && ref.SubjectRelationship == nil:
			if entityErr := k.addEntity(*ref.SubjectEntity); entityErr != nil {
				return fmt.Errorf("entity: %w", entityErr)
			}
		case ref.SubjectRelationship != nil && ref.SubjectEntity == nil:
			if sourceErr := k.addEntity(ref.SubjectRelationship.Source); sourceErr != nil {
				return fmt.Errorf("relationship source entity: %w", sourceErr)
			}
			if targetErr := k.addEntity(ref.SubjectRelationship.Target); targetErr != nil {
				return fmt.Errorf("relationship target entity: %w", targetErr)
			}
		}
	}
	return nil
}

func (s *KnowledgeGraphService) IngestEvidence(ctx context.Context, event *ent.NormalizedEvent, refs ...rez.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	entityRefs := make([]rez.KnowledgeEvidenceRef, 0, len(refs))
	relationshipRefs := make([]rez.KnowledgeEvidenceRef, 0, len(refs))
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

	ik := &knowledgeEntityIdentityLockKeys{}
	if lockKeysErr := ik.collect(refs); lockKeysErr != nil {
		return fmt.Errorf("get entity identity lock keys: %w", lockKeysErr)
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
			return fmt.Errorf("lock knowledge entity identities: %w", lockErr)
		}
		if entsErr := s.ingestEvidenceRefs(ctx, event.ID, ksa.SubjectKindEntity, entityRefs); entsErr != nil {
			return fmt.Errorf("entities: %w", entsErr)
		}
		if relsErr := s.ingestEvidenceRefs(ctx, event.ID, ksa.SubjectKindRelationship, relationshipRefs); relsErr != nil {
			return fmt.Errorf("relationships: %w", relsErr)
		}
		return nil
	})
}

func (s *KnowledgeGraphService) ResolveInternalEntity(ctx context.Context, ref rez.KnowledgeEntityRef) (*ent.KnowledgeSubjectAlias, error) {
	ik := &knowledgeEntityIdentityLockKeys{}
	if entityErr := ik.addEntity(ref); entityErr != nil {
		return nil, fmt.Errorf("entity: %w", entityErr)
	}

	var alias *ent.KnowledgeSubjectAlias
	return alias, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
			return fmt.Errorf("lock knowledge entity identities: %w", lockErr)
		}

		ea, entityErr := s.resolveEntityFromRef(ctx, ref)
		if entityErr != nil {
			return fmt.Errorf("subject entity: %w", entityErr)
		}

		entAlias, aliasErr := tx.KnowledgeSubjectAlias.Get(ctx, ea.aliasId)
		if aliasErr != nil {
			return fmt.Errorf("get knowledge subject alias: %w", aliasErr)
		}
		alias = entAlias.Unwrap()
		return nil
	})
}

func (s *KnowledgeGraphService) ResolveInternalRelationship(ctx context.Context, ref rez.KnowledgeRelationshipRef) (*ent.KnowledgeSubjectAlias, error) {
	ik := &knowledgeEntityIdentityLockKeys{}
	if sourceErr := ik.addEntity(ref.Source); sourceErr != nil {
		return nil, fmt.Errorf("relationship source entity: %w", sourceErr)
	}
	if targetErr := ik.addEntity(ref.Target); targetErr != nil {
		return nil, fmt.Errorf("relationship target entity: %w", targetErr)
	}

	var alias *ent.KnowledgeSubjectAlias
	return alias, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
			return fmt.Errorf("lock knowledge entity identities: %w", lockErr)
		}

		source, sourceErr := s.resolveEntityFromRef(ctx, ref.Source)
		if sourceErr != nil {
			return fmt.Errorf("relationship source entity: %w", sourceErr)
		}
		target, targetErr := s.resolveEntityFromRef(ctx, ref.Target)
		if targetErr != nil {
			return fmt.Errorf("relationship target entity: %w", targetErr)
		}
		ra, relErr := s.resolveRelationshipFromRef(ctx, ref, source, target)
		if relErr != nil {
			return fmt.Errorf("resolve relationship: %w", relErr)
		}

		relAlias, aliasErr := tx.KnowledgeSubjectAlias.Get(ctx, ra.aliasId)
		if aliasErr != nil {
			return fmt.Errorf("get knowledge subject alias: %w", aliasErr)
		}
		alias = relAlias.Unwrap()
		return nil
	})
}
