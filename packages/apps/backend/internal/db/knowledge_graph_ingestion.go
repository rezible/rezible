package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
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
	"github.com/rezible/rezible/pkg/knowledgegraph"
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
		if lockErr := s.lockResolvedEntity(ctx, resolvedAlias.subjectId); lockErr != nil {
			return nil, lockErr
		}
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
			if lockErr := s.lockResolvedEntity(ctx, entity.ID); lockErr != nil {
				return nil, lockErr
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

	if newName, isNewServiceName := insertedValues[projections.LinkingAttributeServiceName]; isNewServiceName && knowledgegraph.IsService(ref.Category, ref.Kind) {
		if similarErr := s.proposeSimilarServiceNames(ctx, ref, resolvedAlias, newName); similarErr != nil {
			return nil, fmt.Errorf("propose similar service names: %w", similarErr)
		}
	}

	return resolvedAlias, nil
}

// lockResolvedEntity key-share locks an existing entity's row before anything is written for it. A projection
// can resolve an existing alias and write only evidence, which references the alias and not the entity, so no
// foreign key protects the entity it resolved. A merge (MergeEntities) locks both its entities FOR UPDATE, in ID
// order, and deletes the one merged away: with this lock, the merge waits for in-flight projections of the
// entity, and a projection that resolved an entity the merge has since deleted finds no row and returns
// errs.ErrKnowledgeSubjectMerged, so its job runs again on the merged graph. FOR KEY SHARE rather than FOR SHARE,
// so projections of one entity still run side by side: they also update its state, which another projection's
// FOR SHARE would block. A projection naming both of a merge's entities locks them one at a time, in its own
// order; Postgres aborts one side of such a cycle as a deadlock, and its job or request retries.
func (s *KnowledgeGraphIngestionService) lockResolvedEntity(ctx context.Context, entityID uuid.UUID) error {
	lockEntity := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(entityID)).
		ForShare(ent.KeyShareLock)
	if _, lockErr := lockEntity.OnlyID(ctx); lockErr != nil {
		if ent.IsNotFound(lockErr) {
			return fmt.Errorf("%w: knowledge entity %s", errs.ErrKnowledgeSubjectMerged, entityID)
		}
		return fmt.Errorf("lock entity: %w", lockErr)
	}
	return nil
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
	// Endpoint resolution rejects an alias whose entity differs from the ref's category, so the ref's
	// categories are the resolved endpoints' categories.
	if validErr := knowledgegraph.ValidateRelationship(ref.Predicate, ref.Source.Category, ref.Target.Category); validErr != nil {
		return nil, validErr
	}
	if ref.Predicate == knr.PredicateContains && source.subjectId == target.subjectId {
		return nil, fmt.Errorf("%w: an entity cannot contain itself", errs.ErrInvalidInput)
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
		if lockErr := s.lockResolvedEntity(ctx, source.subjectId); lockErr != nil {
			return nil, fmt.Errorf("source: %w", lockErr)
		}
		if lockErr := s.lockResolvedEntity(ctx, target.subjectId); lockErr != nil {
			return nil, fmt.Errorf("target: %w", lockErr)
		}
	}

	return s.resolveRelationshipFromRef(ctx, *ref.Relationship, source, target)
}

func (s *KnowledgeGraphIngestionService) ingestEvidenceRefs(ctx context.Context, eventID uuid.UUID, refs []rez.KnowledgeEvidenceRef) error {
	if len(refs) == 0 {
		return nil
	}
	resolvedRefs := make([]*resolvedSubjectAlias, len(refs))
	aliasIDs := make([]uuid.UUID, len(refs))
	proposing := make([]*resolvedSubjectAlias, 0)
	for i, ref := range refs {
		resolved, aliasErr := s.resolveSubjectAliasFromRef(ctx, ref.Subject, false)
		if aliasErr != nil {
			return fmt.Errorf("resolve subject alias: %w", aliasErr)
		}
		resolvedRefs[i], aliasIDs[i] = resolved, resolved.aliasId
		if len(resolved.proposals) > 0 {
			proposing = append(proposing, resolved)
		}
	}

	// An event states one thing per alias: evidence it already recorded for an alias (a replay, or a second
	// call for the same event) is kept and not applied again, and within the batch the first statement wins.
	queryRecorded := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(eventID), ke.SubjectAliasIDIn(aliasIDs...))
	recordedEvidence, recordedErr := queryRecorded.All(ctx)
	if recordedErr != nil {
		return fmt.Errorf("query recorded evidence: %w", recordedErr)
	}
	recorded := mapset.NewSet[uuid.UUID]()
	for _, evidence := range recordedEvidence {
		recorded.Add(evidence.SubjectAliasID)
	}

	evClient := s.db.Client(ctx).KnowledgeEvidence
	builders := make([]*ent.KnowledgeEvidenceCreate, 0, len(refs))
	entityIDs := mapset.NewSet[uuid.UUID]()
	relationshipIDs := mapset.NewSet[uuid.UUID]()
	for i, ref := range refs {
		resolved := resolvedRefs[i]
		if !recorded.Add(resolved.aliasId) {
			continue
		}
		evidence := knowledgegraph.AliasEvidence{
			Kind:        ref.Kind,
			StateMode:   ref.StateMode,
			EffectiveAt: ref.EffectiveAt,
			State:       ref.SubjectState,
		}
		if evidence.StateMode == "" {
			evidence.StateMode = ke.StateModeMerge
		}
		if ref.ValidFor > 0 {
			evidence.ExpiresAt = new(ref.EffectiveAt.Add(ref.ValidFor))
		}
		if applyErr := s.applyToAlias(ctx, resolved.aliasId, evidence); applyErr != nil {
			return fmt.Errorf("apply evidence to alias: %w", applyErr)
		}
		if ref.Subject.Entity != nil {
			entityIDs.Add(resolved.subjectId)
		} else {
			relationshipIDs.Add(resolved.subjectId)
		}
		builders = append(builders, evClient.Create().
			SetEventID(eventID).
			SetSubjectAliasID(resolved.aliasId).
			SetKind(ref.Kind).
			SetAssertion(ref.Assertion).
			SetEffectiveAt(ref.EffectiveAt).
			SetSubjectState(ref.SubjectState).
			SetStateMode(evidence.StateMode).
			SetNillableExpiresAt(evidence.ExpiresAt))
	}
	if len(builders) > 0 {
		createEvidence := evClient.CreateBulk(builders...).
			OnConflict(knowledgeEvidenceUniqueColumns).
			Ignore()
		if createErr := createEvidence.Exec(ctx); createErr != nil {
			return fmt.Errorf("create knowledge evidence: %w", createErr)
		}
	}
	for _, entityID := range sortedUUIDs(entityIDs) {
		if mergeErr := s.mergeEntityState(ctx, entityID); mergeErr != nil {
			return fmt.Errorf("merge entity state: %w", mergeErr)
		}
	}
	children := mapset.NewSet[uuid.UUID]()
	for _, relationshipID := range sortedUUIDs(relationshipIDs) {
		relationship, mergeErr := s.mergeRelationshipState(ctx, relationshipID)
		if mergeErr != nil {
			return fmt.Errorf("merge relationship state: %w", mergeErr)
		}
		if relationship.Predicate == knr.PredicateContains {
			children.Add(relationship.TargetEntityID)
		}
	}
	for _, childID := range sortedUUIDs(children) {
		if parentErr := s.recomputeParent(ctx, childID); parentErr != nil {
			return fmt.Errorf("recompute parent: %w", parentErr)
		}
	}
	if proposalsErr := s.createProposals(ctx, eventID, proposing); proposalsErr != nil {
		return fmt.Errorf("create knowledge proposals: %w", proposalsErr)
	}
	return nil
}

// applyToAlias applies one evidence row to its alias's state and observation times. The identity locks taken
// by the ingesting transaction cover the alias's resource ref, so no other ingestion updates the alias meanwhile.
func (s *KnowledgeGraphIngestionService) applyToAlias(ctx context.Context, aliasID uuid.UUID, evidence knowledgegraph.AliasEvidence) error {
	client := s.db.Client(ctx)
	alias, queryErr := client.KnowledgeSubjectAlias.Get(ctx, aliasID)
	if queryErr != nil {
		return fmt.Errorf("load alias: %w", queryErr)
	}
	record := knowledgegraph.ApplyEvidence(aliasRecord(alias), evidence)
	updateAlias := client.KnowledgeSubjectAlias.UpdateOneID(aliasID).
		SetState(record.State)
	if mutationErr := setTimeFields(updateAlias.Mutation(), map[string]*time.Time{
		ksa.FieldFirstObservedAt: record.FirstObservedAt,
		ksa.FieldLastObservedAt:  record.LastObservedAt,
		ksa.FieldWithdrawnAt:     record.WithdrawnAt,
		ksa.FieldExpiresAt:       record.ExpiresAt,
	}); mutationErr != nil {
		return mutationErr
	}
	return updateAlias.Exec(ctx)
}

func aliasRecord(alias *ent.KnowledgeSubjectAlias) knowledgegraph.AliasRecord {
	return knowledgegraph.AliasRecord{
		State:           alias.State,
		FirstObservedAt: alias.FirstObservedAt,
		LastObservedAt:  alias.LastObservedAt,
		WithdrawnAt:     alias.WithdrawnAt,
		ExpiresAt:       alias.ExpiresAt,
	}
}

func (s *KnowledgeGraphIngestionService) subjectAliases(ctx context.Context, pred predicate.KnowledgeSubjectAlias) ([]knowledgegraph.SubjectAlias, error) {
	aliases, queryErr := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(pred).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query subject aliases: %w", queryErr)
	}
	subjectAliases := make([]knowledgegraph.SubjectAlias, len(aliases))
	for i, alias := range aliases {
		subjectAliases[i] = knowledgegraph.SubjectAlias{ID: alias.ID, CreatedAt: alias.CreatedAt, Record: aliasRecord(alias)}
	}
	return subjectAliases, nil
}

// mergeEntityState re-merges the entity's state from all of its aliases. The identity locks do not serialize
// projections of one entity under two names, which a merge leaves, so the read-modify-write locks the row
// FOR NO KEY UPDATE; that does not conflict with the FOR KEY SHARE lock of lockResolvedEntity.
func (s *KnowledgeGraphIngestionService) mergeEntityState(ctx context.Context, entityID uuid.UUID) error {
	client := s.db.Client(ctx)
	lockEntity := client.KnowledgeEntity.Query().
		Where(kne.ID(entityID)).
		ForUpdate(ent.KeyUpdateLock)
	entity, lockErr := lockEntity.Only(ctx)
	if lockErr != nil {
		return fmt.Errorf("lock entity: %w", lockErr)
	}
	aliases, aliasesErr := s.subjectAliases(ctx, ksa.EntityID(entityID))
	if aliasesErr != nil {
		return aliasesErr
	}
	merged := knowledgegraph.MergeSubjectState(aliases)
	update := client.KnowledgeEntity.UpdateOneID(entityID)
	if mutationErr := setSubjectState(update.Mutation(), merged); mutationErr != nil {
		return mutationErr
	}
	if updateErr := update.Exec(ctx); updateErr != nil {
		return updateErr
	}
	// Becoming withdrawn or current again changes which contains relationships are candidates, for the entity's
	// own parent and for the entities it contains.
	if (entity.WithdrawnAt == nil) != (merged.WithdrawnAt == nil) {
		return s.recomputeParentsAround(ctx, entityID)
	}
	return nil
}

// recomputeParentsAround recomputes the parent of the entity and of every entity it contains.
func (s *KnowledgeGraphIngestionService) recomputeParentsAround(ctx context.Context, entityID uuid.UUID) error {
	queryChildren := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateContains), knr.SourceEntityID(entityID))
	contains, childrenErr := queryChildren.All(ctx)
	if childrenErr != nil {
		return fmt.Errorf("query contained entities: %w", childrenErr)
	}
	ids := mapset.NewSet(entityID)
	for _, relationship := range contains {
		ids.Add(relationship.TargetEntityID)
	}
	for _, id := range sortedUUIDs(ids) {
		if parentErr := s.recomputeParent(ctx, id); parentErr != nil {
			return parentErr
		}
	}
	return nil
}

// recomputeParent chooses the entity's parent from the current contains relationships targeting it: the first
// claim Rezible saw (earliest created, then lowest ID) whose source is not the entity and does not have the
// entity among its ancestors. With no candidate the parent is cleared. Staleness does not affect the choice.
//
// The child is locked before the claims are read. Two ingestions can add claims on one child under different
// aliases, which the identity locks do not serialize; the later one then waits here and reads the earlier one's
// committed claim (READ COMMITTED), so the last write sees every claim. Callers lock children in ID order; a cycle with another
// transaction's locks aborts one side as a deadlock, and its job or request retries.
func (s *KnowledgeGraphIngestionService) recomputeParent(ctx context.Context, childID uuid.UUID) error {
	client := s.db.Client(ctx)
	lockChild := client.KnowledgeEntity.Query().
		Where(kne.ID(childID)).
		ForUpdate(ent.KeyUpdateLock)
	if _, lockErr := lockChild.OnlyID(ctx); lockErr != nil {
		return fmt.Errorf("lock entity: %w", lockErr)
	}
	queryClaims := client.KnowledgeRelationship.Query().
		Where(
			knr.PredicateEQ(knr.PredicateContains),
			knr.TargetEntityID(childID),
			knr.SourceEntityIDNEQ(childID),
			currentKnowledgeRelationship(),
		).
		Order(knr.ByCreatedAt(), knr.ByID())
	claims, claimsErr := queryClaims.All(ctx)
	if claimsErr != nil {
		return fmt.Errorf("query parent claims: %w", claimsErr)
	}
	var parentID *uuid.UUID
	for _, claim := range claims {
		cycles, walkErr := s.hasAncestor(ctx, claim.SourceEntityID, childID)
		if walkErr != nil {
			return walkErr
		}
		if !cycles {
			parentID = &claim.SourceEntityID
			break
		}
	}
	update := client.KnowledgeEntity.UpdateOneID(childID)
	if parentID == nil {
		update.ClearParentID()
	} else {
		update.SetParentID(*parentID)
	}
	return update.Exec(ctx)
}

// hasAncestor reports whether ancestorID is the entity or one of its ancestors, walking parent_id at most
// MaxAncestryDepth steps.
func (s *KnowledgeGraphIngestionService) hasAncestor(ctx context.Context, entityID, ancestorID uuid.UUID) (bool, error) {
	current := &entityID
	visited := mapset.NewSet[uuid.UUID]()
	for range knowledgegraph.MaxAncestryDepth {
		if current == nil || !visited.Add(*current) {
			return false, nil
		}
		if *current == ancestorID {
			return true, nil
		}
		entity, queryErr := s.db.Client(ctx).KnowledgeEntity.Get(ctx, *current)
		if queryErr != nil {
			return false, fmt.Errorf("walk ancestors: %w", queryErr)
		}
		current = entity.ParentID
	}
	return false, nil
}

// mergeRelationshipState re-merges the relationship's state from all of its aliases, locking the row as
// mergeEntityState does, for the same reason.
func (s *KnowledgeGraphIngestionService) mergeRelationshipState(ctx context.Context, relationshipID uuid.UUID) (*ent.KnowledgeRelationship, error) {
	client := s.db.Client(ctx)
	lockRelationship := client.KnowledgeRelationship.Query().
		Where(knr.ID(relationshipID)).
		ForUpdate(ent.KeyUpdateLock)
	relationship, lockErr := lockRelationship.Only(ctx)
	if lockErr != nil {
		return nil, fmt.Errorf("lock relationship: %w", lockErr)
	}
	aliases, aliasesErr := s.subjectAliases(ctx, ksa.RelationshipID(relationshipID))
	if aliasesErr != nil {
		return nil, aliasesErr
	}
	update := client.KnowledgeRelationship.UpdateOneID(relationshipID)
	if mutationErr := setSubjectState(update.Mutation(), knowledgegraph.MergeSubjectState(aliases)); mutationErr != nil {
		return nil, mutationErr
	}
	return relationship, update.Exec(ctx)
}

// fieldMutation is the part of a generated mutation that sets fields by name. Entities and relationships share
// the subject state columns, so one function writes both.
type fieldMutation interface {
	SetField(name string, value ent.Value) error
	ClearField(name string) error
}

func setSubjectState(m fieldMutation, merged knowledgegraph.SubjectState) error {
	if setErr := m.SetField(kne.FieldState, merged.State); setErr != nil {
		return fmt.Errorf("set state: %w", setErr)
	}
	if setErr := m.SetField(kne.FieldStateSources, merged.Sources); setErr != nil {
		return fmt.Errorf("set state sources: %w", setErr)
	}
	return setTimeFields(m, map[string]*time.Time{
		kne.FieldStateEffectiveAt: merged.StateEffectiveAt,
		kne.FieldLastObservedAt:   merged.LastObservedAt,
		kne.FieldWithdrawnAt:      merged.WithdrawnAt,
		kne.FieldExpiresAt:        merged.ExpiresAt,
	})
}

// setTimeFields sets each nullable time field, clearing it when nil.
func setTimeFields(m fieldMutation, fields map[string]*time.Time) error {
	for name, value := range fields {
		var fieldErr error
		if value == nil {
			fieldErr = m.ClearField(name)
		} else {
			fieldErr = m.SetField(name, *value)
		}
		if fieldErr != nil {
			return fmt.Errorf("set %s: %w", name, fieldErr)
		}
	}
	return nil
}

func sortedUUIDs(ids mapset.Set[uuid.UUID]) []uuid.UUID {
	sorted := ids.ToSlice()
	slices.SortFunc(sorted, func(a, b uuid.UUID) int {
		return bytes.Compare(a[:], b[:])
	})
	return sorted
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
