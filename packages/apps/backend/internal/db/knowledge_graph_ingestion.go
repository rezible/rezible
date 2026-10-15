package db

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/projections"
)

type KnowledgeGraphIngestionService struct {
	db rez.Database
}

func NewKnowledgeGraphIngestionService(db rez.Database) (*KnowledgeGraphIngestionService, error) {
	return &KnowledgeGraphIngestionService{db: db}, nil
}

func (s *KnowledgeGraphIngestionService) IngestEvidence(ctx context.Context, event *ent.NormalizedEvent, refs ...rez.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}

	ik := &knowledgeEntityIdentityLockKeys{}
	if lockKeysErr := ik.collectSubjects(refs); lockKeysErr != nil {
		return fmt.Errorf("get entity identity lock keys: %w", lockKeysErr)
	}

	entityRefs := make([]rez.KnowledgeEvidenceRef, 0, len(refs))
	relationshipRefs := make([]rez.KnowledgeEvidenceRef, 0, len(refs))
	for _, ref := range refs {
		subj := ref.Subject
		switch {
		case subj.Entity != nil && subj.Relationship == nil:
			entityRefs = append(entityRefs, ref)
		case subj.Relationship != nil && subj.Entity == nil:
			relationshipRefs = append(relationshipRefs, ref)
		default:
			return fmt.Errorf("evidence must contain exactly one entity or relationship")
		}
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
			return fmt.Errorf("lock knowledge entity identities: %w", lockErr)
		}
		if entsErr := s.ingestEvidenceRefs(ctx, event.ID, entityRefs); entsErr != nil {
			return fmt.Errorf("entities: %w", entsErr)
		}
		if relsErr := s.ingestEvidenceRefs(ctx, event.ID, relationshipRefs); relsErr != nil {
			return fmt.Errorf("relationships: %w", relsErr)
		}
		return nil
	})
}

func (s *KnowledgeGraphIngestionService) ResolveInternalSubject(ctx context.Context, ref rez.KnowledgeSubjectRef) (*ent.KnowledgeSubjectAlias, error) {
	ik := &knowledgeEntityIdentityLockKeys{}
	if ref.Entity != nil {
		if entityErr := ik.addEntity(*ref.Entity); entityErr != nil {
			return nil, fmt.Errorf("entity: %w", entityErr)
		}
	} else if ref.Relationship != nil {
		if sourceErr := ik.addEntity(ref.Relationship.Source); sourceErr != nil {
			return nil, fmt.Errorf("relationship source entity: %w", sourceErr)
		}
		if targetErr := ik.addEntity(ref.Relationship.Target); targetErr != nil {
			return nil, fmt.Errorf("relationship target entity: %w", targetErr)
		}
	} else {
		return nil, fmt.Errorf("invalid subject")
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.KnowledgeSubjectAlias, error) {
		if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
			return nil, fmt.Errorf("lock knowledge entity identities: %w", lockErr)
		}

		// Internal subjects carry no evidence, so any proposals resolution finds have nothing to cite and are
		// not written.
		resolvedAlias, resolveErr := s.resolveSubjectAliasFromRef(ctx, ref, true)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolve alias: %w", resolveErr)
		}

		subjAlias, aliasErr := tx.KnowledgeSubjectAlias.Get(ctx, resolvedAlias.aliasId)
		if aliasErr != nil {
			return nil, fmt.Errorf("get knowledge subject alias: %w", aliasErr)
		}
		return subjAlias, nil
	})
}

type resolvedSubjectAlias struct {
	aliasId   uuid.UUID
	subjectId uuid.UUID
	// proposals are the matches resolution found for the entity, written once the alias's evidence exists.
	proposals []knowledgeProposalDraft
}

func (s *KnowledgeGraphIngestionService) makeAliasQuery(ctx context.Context, ref rez.ProviderResourceRef, preds ...predicate.KnowledgeSubjectAlias) *ent.KnowledgeSubjectAliasQuery {
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

func (s *KnowledgeGraphIngestionService) ensureAlias(ctx context.Context, ref rez.ProviderResourceRef, subjectKind ksa.SubjectKind, entityID, relationshipID uuid.UUID) (*ent.KnowledgeSubjectAlias, error) {
	query := s.makeAliasQuery(ctx, ref)
	alias, queryErr := query.Only(ctx)
	if alias != nil {
		if alias.SubjectKind != subjectKind ||
			(entityID != uuid.Nil && (alias.EntityID == nil || *alias.EntityID != entityID)) ||
			(relationshipID != uuid.Nil && (alias.RelationshipID == nil || *alias.RelationshipID != relationshipID)) {
			return nil, fmt.Errorf("%w: provider resource is already mapped to another knowledge subject", errs.ErrConflict)
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

func (s *KnowledgeGraphIngestionService) lookupExistingEntityAlias(ctx context.Context, ref rez.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
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
		return nil, fmt.Errorf("%w: resource already identifies a relationship", errs.ErrConflict)
	}
	entity, entityErr := alias.Edges.EntityOrErr()
	if entityErr != nil {
		return nil, fmt.Errorf("load alias entity: %w", entityErr)
	}
	if entity.Category != ref.Category || entity.Kind != ref.Kind {
		return nil, fmt.Errorf("%w: alias identifies %q/%q, evidence expects %q/%q", errs.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entity.ID}, nil
}

// queryLinkingAttributes loads the stored linking attributes holding any of the values, with their entities.
func (s *KnowledgeGraphIngestionService) queryLinkingAttributes(ctx context.Context, values map[string]string) ([]*ent.KnowledgeEntityLinkingAttribute, error) {
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
	return linkingAttributes, nil
}

// linkedEntity returns the one entity holding the loaded linking attributes, if any.
func (s *KnowledgeGraphIngestionService) linkedEntity(linkingAttributes []*ent.KnowledgeEntityLinkingAttribute) (*ent.KnowledgeEntity, error) {
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
			return nil, fmt.Errorf("%w: linking attributes identify multiple entities", errs.ErrConflict)
		}
	}
	return linked, nil
}

var knowledgeEntityLinkingAttributeUniqueColumns = sql.ConflictColumns(kela.FieldTenantID, kela.FieldAttribute, kela.FieldValue)

// ensureEntityLinkingAttributes adds the values not already stored and returns them. existing are the stored
// linking attributes holding any of the values; callers have dropped or rejected values held by another entity.
func (s *KnowledgeGraphIngestionService) ensureEntityLinkingAttributes(ctx context.Context, entityID uuid.UUID, values map[string]string, existing []*ent.KnowledgeEntityLinkingAttribute) (map[string]string, error) {
	inserted := maps.Clone(values)
	for _, linkingAttribute := range existing {
		delete(inserted, linkingAttribute.Attribute)
	}
	if len(inserted) == 0 {
		return nil, nil
	}

	attrClient := s.db.Client(ctx).KnowledgeEntityLinkingAttribute
	builders := make([]*ent.KnowledgeEntityLinkingAttributeCreate, 0, len(inserted))
	for attribute, value := range inserted {
		builders = append(builders, attrClient.Create().
			SetEntityID(entityID).
			SetAttribute(attribute).
			SetValue(value))
	}
	upsertAttrs := attrClient.CreateBulk(builders...).
		OnConflict(knowledgeEntityLinkingAttributeUniqueColumns).
		Ignore()
	if upsertErr := upsertAttrs.Exec(ctx); upsertErr != nil {
		return nil, fmt.Errorf("save entity linking attributes: %w", upsertErr)
	}
	return inserted, nil
}

func getKnowledgeEntityLinkingValues(ref rez.KnowledgeEntityRef) (map[string]string, error) {
	if ref.LinkingAttributes == nil {
		return nil, nil
	}
	values := ref.LinkingAttributes.Values()
	for attribute, value := range values {
		if strings.TrimSpace(attribute) == "" || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%w: linking attribute and value must be non-empty", errs.ErrInvalidInput)
		}
	}
	return values, nil
}

func (s *KnowledgeGraphIngestionService) resolveEntityFromRef(ctx context.Context, ref rez.KnowledgeEntityRef) (*resolvedSubjectAlias, error) {
	linkingValues, valuesErr := getKnowledgeEntityLinkingValues(ref)
	if valuesErr != nil {
		return nil, valuesErr
	}

	existingAlias, lookupAliasErr := s.lookupExistingEntityAlias(ctx, ref)
	if lookupAliasErr != nil {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupAliasErr)
	}
	existingValues, valuesQueryErr := s.queryLinkingAttributes(ctx, linkingValues)
	if valuesQueryErr != nil {
		return nil, valuesQueryErr
	}

	var resolvedAlias *resolvedSubjectAlias
	if existingAlias != nil {
		resolvedAlias = existingAlias
		unconflicted, conflictErr := s.proposeIdentityConflicts(ctx, ref, resolvedAlias, linkingValues, existingValues)
		if conflictErr != nil {
			return nil, conflictErr
		}
		linkingValues = unconflicted
	} else {
		entity, linkedErr := s.linkedEntity(existingValues)
		if linkedErr != nil {
			return nil, linkedErr
		} else if entity != nil {
			if entity.Category != ref.Category || entity.Kind != ref.Kind {
				return nil, fmt.Errorf("%w: linking attribute identifies %q/%q, evidence expects %q/%q",
					errs.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
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

	insertedValues, linkingErr := s.ensureEntityLinkingAttributes(ctx, resolvedAlias.subjectId, linkingValues, existingValues)
	if linkingErr != nil {
		return nil, linkingErr
	}

	if newName, isNewServiceName := insertedValues[projections.LinkingAttributeServiceName]; isNewServiceName && isKnowledgeServiceRef(ref) {
		if similarErr := s.proposeSimilarServiceNames(ctx, ref, resolvedAlias, newName); similarErr != nil {
			return nil, fmt.Errorf("propose similar service names: %w", similarErr)
		}
	}

	return resolvedAlias, nil
}

func (s *KnowledgeGraphIngestionService) lookupExistingRelationshipAlias(ctx context.Context, ref rez.KnowledgeRelationshipRef, sourceId, targetId uuid.UUID) (*resolvedSubjectAlias, error) {
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
		return nil, fmt.Errorf("%w: resource already identifies an entity", errs.ErrConflict)
	}
	rel, relationshipErr := alias.Edges.RelationshipOrErr()
	if relationshipErr != nil {
		return nil, fmt.Errorf("load aliased relationship: %w", relationshipErr)
	}
	if rel.Predicate != ref.Predicate || rel.SourceEntityID != sourceId || rel.TargetEntityID != targetId {
		return nil, fmt.Errorf("%w: relationship alias identifies different topology", errs.ErrConflict)
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

func (s *KnowledgeGraphIngestionService) resolveRelationshipFromRef(ctx context.Context, ref rez.KnowledgeRelationshipRef, source, target *resolvedSubjectAlias) (*resolvedSubjectAlias, error) {
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

func (s *KnowledgeGraphIngestionService) resolveSubjectAliasFromRef(ctx context.Context, ref rez.KnowledgeSubjectRef, resolveRelEndpoints bool) (*resolvedSubjectAlias, error) {
	if ref.Entity != nil {
		return s.resolveEntityFromRef(ctx, *ref.Entity)
	}

	if refErr := ref.Relationship.ProviderResourceRef.Validate(); refErr != nil {
		return nil, fmt.Errorf("validate resource ref: %w", refErr)
	}

	var source *resolvedSubjectAlias
	var target *resolvedSubjectAlias
	var endpointErr error
	if resolveRelEndpoints {
		source, endpointErr = s.resolveEntityFromRef(ctx, ref.Relationship.Source)
		if endpointErr != nil {
			return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
		}
		target, endpointErr = s.resolveEntityFromRef(ctx, ref.Relationship.Target)
		if endpointErr != nil {
			return nil, fmt.Errorf("target: %w", errNoExistingEndpointEntityAlias)
		}
	} else {
		source, endpointErr = s.lookupExistingEntityAlias(ctx, ref.Relationship.Source)
		if endpointErr != nil || source == nil {
			return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
		}
		target, endpointErr = s.lookupExistingEntityAlias(ctx, ref.Relationship.Target)
		if endpointErr != nil || target == nil {
			return nil, fmt.Errorf("target: %w", errNoExistingEndpointEntityAlias)
		}
	}

	return s.resolveRelationshipFromRef(ctx, *ref.Relationship, source, target)
}

func (s *KnowledgeGraphIngestionService) ingestEvidenceRefs(ctx context.Context, eventID uuid.UUID, refs []rez.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	evClient := s.db.Client(ctx).KnowledgeEvidence
	builders := make([]*ent.KnowledgeEvidenceCreate, 0, len(refs))
	proposing := make([]*resolvedSubjectAlias, 0)
	for _, ref := range refs {
		resolved, aliasErr := s.resolveSubjectAliasFromRef(ctx, ref.Subject, false)
		if aliasErr != nil {
			return fmt.Errorf("resolve subject alias: %w", aliasErr)
		}
		if len(resolved.proposals) > 0 {
			proposing = append(proposing, resolved)
		}
		if ref.Subject.Entity != nil && ref.Kind == ke.KindObserved {
			if stateErr := s.updateEntityState(ctx, resolved.subjectId, ref); stateErr != nil {
				return fmt.Errorf("update entity state: %w", stateErr)
			}
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
	if proposalsErr := s.createProposals(ctx, eventID, proposing); proposalsErr != nil {
		return fmt.Errorf("create knowledge proposals: %w", proposalsErr)
	}
	return nil
}

// updateEntityState merges observed evidence into the entity's current state. Evidence with an empty state
// changes nothing, and evidence older than the stored state is skipped. The read-modify-write is safe
// without another lock because IngestEvidence holds the entity identity locks for its whole transaction.
func (s *KnowledgeGraphIngestionService) updateEntityState(ctx context.Context, entityID uuid.UUID, ref rez.KnowledgeEvidenceRef) error {
	if ref.SubjectState.IsEmpty() {
		return nil
	}
	client := s.db.Client(ctx)
	entityQuery := client.KnowledgeEntity.Query().
		Where(kne.ID(entityID)).
		Select(kne.FieldState, kne.FieldStateEffectiveAt)
	entity, queryErr := entityQuery.Only(ctx)
	if queryErr != nil {
		return fmt.Errorf("load entity: %w", queryErr)
	}
	if entity.StateEffectiveAt != nil && ref.EffectiveAt.Before(*entity.StateEffectiveAt) {
		return nil
	}
	updateState := client.KnowledgeEntity.UpdateOneID(entityID).
		SetState(entity.State.MergeFrom(ref.SubjectState)).
		SetStateEffectiveAt(ref.EffectiveAt)
	return updateState.Exec(ctx)
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
		// Two events naming similar services, such as `checkout` and `checkout-api`, would each miss the other's
		// uncommitted name and neither would propose the match. Locking on the stem serializes them, so the
		// later one sees the earlier one's name.
		if attribute == projections.LinkingAttributeServiceName {
			if stem := projections.ServiceNameStem(value); stem != "" {
				k.keys = append(k.keys, k.makeKey("service_name_stem", stem))
			}
		}
	}
	return nil
}

func (k *knowledgeEntityIdentityLockKeys) collectSubjects(refs []rez.KnowledgeEvidenceRef) error {
	for _, ref := range refs {
		subj := ref.Subject
		switch {
		case subj.Entity != nil && subj.Relationship == nil:
			if entityErr := k.addEntity(*subj.Entity); entityErr != nil {
				return fmt.Errorf("entity: %w", entityErr)
			}
		case subj.Relationship != nil && subj.Entity == nil:
			if sourceErr := k.addEntity(subj.Relationship.Source); sourceErr != nil {
				return fmt.Errorf("relationship source entity: %w", sourceErr)
			}
			if targetErr := k.addEntity(subj.Relationship.Target); targetErr != nil {
				return fmt.Errorf("relationship target entity: %w", targetErr)
			}
		}
	}
	return nil
}
