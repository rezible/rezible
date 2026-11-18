package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"

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
		if !ref.Authority.IsValid() {
			return fmt.Errorf("%w: evidence must state its authority", errs.ErrInvalidInput)
		}
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
		refresh := newSubjectRefresh()
		if entsErr := s.ingestEvidenceRefs(ctx, event, entityRefs, refresh); entsErr != nil {
			return fmt.Errorf("entities: %w", entsErr)
		}
		if relsErr := s.ingestEvidenceRefs(ctx, event, relationshipRefs, refresh); relsErr != nil {
			return fmt.Errorf("relationships: %w", relsErr)
		}
		return s.flush(ctx, refresh)
	})
}

func (s *KnowledgeGraphIngestionService) ResolveInternalSubject(ctx context.Context, ref rez.KnowledgeSubjectRef) (*ent.KnowledgeSubjectAlias, error) {
	ik := &knowledgeEntityIdentityLockKeys{}
	if ref.Entity != nil {
		if entityErr := ik.addEntity(*ref.Entity); entityErr != nil {
			return nil, fmt.Errorf("entity: %w", entityErr)
		}
	} else if ref.Relationship != nil {
		ik.addAlias(ref.Relationship.ProviderResourceRef)
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
		// not written. Rezible is the system of record for its internal subjects.
		resolvedAlias, resolveErr := s.resolveSubjectAliasFromRef(ctx, ref, schematypes.KnowledgeAuthorityDeclared, true)
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

// ensureAlias returns the alias for the resource ref, creating it with the authority when it is new. An existing
// alias's authority is raised by the evidence applied to it (applyToAlias).
func (s *KnowledgeGraphIngestionService) ensureAlias(ctx context.Context, ref rez.ProviderResourceRef, authority schematypes.KnowledgeAuthority, subjectKind ksa.SubjectKind, entityID, relationshipID uuid.UUID) (*ent.KnowledgeSubjectAlias, error) {
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
		SetSubjectKind(subjectKind).
		SetAuthority(authority)
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
	// The entity is loaded after the alias, so a merge that committed in between leaves it a redirect, which
	// lockResolvedEntity follows.
	queryAlias := s.makeAliasQuery(ctx, ref.ProviderResourceRef).
		WithEntity()
	alias, queryErr := queryAlias.Only(ent.IncludeMerged(ctx))
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

// queryLinkingAttributes loads the stored linking attributes holding any of the values, with their entities. The
// entities are loaded after the attributes, so one may be a redirect a merge left in between, which
// lockResolvedEntity follows.
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
	linkingAttributes, queryErr := lookupAttrsQuery.All(ent.IncludeMerged(ctx))
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

func (s *KnowledgeGraphIngestionService) resolveEntityFromRef(ctx context.Context, ref rez.KnowledgeEntityRef, authority schematypes.KnowledgeAuthority) (*resolvedSubjectAlias, error) {
	linkingValues, valuesErr := getKnowledgeEntityLinkingValues(ref)
	if valuesErr != nil {
		return nil, valuesErr
	}

	existingAlias, lookupAliasErr := s.lookupExistingEntityAlias(ctx, ref)
	if lookupAliasErr != nil {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupAliasErr)
	}

	var resolvedAlias *resolvedSubjectAlias
	var existingValues []*ent.KnowledgeEntityLinkingAttribute
	var valuesQueryErr error
	if existingAlias != nil {
		resolvedAlias = existingAlias
		entityID, lockErr := s.lockResolvedEntity(ctx, resolvedAlias.subjectId)
		if lockErr != nil {
			return nil, lockErr
		}
		resolvedAlias.subjectId = entityID
		// Read after the lock, so a merge that committed while it waited has moved them.
		existingValues, valuesQueryErr = s.queryLinkingAttributes(ctx, linkingValues)
		if valuesQueryErr != nil {
			return nil, valuesQueryErr
		}
		unconflicted, conflictErr := s.proposeIdentityConflicts(ctx, ref, resolvedAlias, linkingValues, existingValues)
		if conflictErr != nil {
			return nil, conflictErr
		}
		linkingValues = unconflicted
	} else {
		existingValues, valuesQueryErr = s.queryLinkingAttributes(ctx, linkingValues)
		if valuesQueryErr != nil {
			return nil, valuesQueryErr
		}
		entity, linkedErr := s.linkedEntity(existingValues)
		var entityID uuid.UUID
		if linkedErr != nil {
			return nil, linkedErr
		} else if entity != nil {
			if entity.Category != ref.Category || entity.Kind != ref.Kind {
				return nil, fmt.Errorf("%w: linking attribute identifies %q/%q, evidence expects %q/%q",
					errs.ErrConflict, entity.Category, entity.Kind, ref.Category, ref.Kind)
			}
			lockedID, lockErr := s.lockResolvedEntity(ctx, entity.ID)
			if lockErr != nil {
				return nil, lockErr
			}
			entityID = lockedID
			if entityID != entity.ID {
				// A merge that committed while the lock waited moved the attributes read before it.
				existingValues, valuesQueryErr = s.queryLinkingAttributes(ctx, linkingValues)
				if valuesQueryErr != nil {
					return nil, valuesQueryErr
				}
			}
		} else {
			createEntity := s.db.Client(ctx).KnowledgeEntity.Create().
				SetCategory(ref.Category).
				SetKind(ref.Kind)
			created, createErr := createEntity.Save(ctx)
			if createErr != nil {
				return nil, fmt.Errorf("create entity: %w", createErr)
			}
			entityID = created.ID
		}
		alias, aliasErr := s.ensureAlias(ctx, ref.ProviderResourceRef, authority, ksa.SubjectKindEntity, entityID, uuid.Nil)
		if aliasErr != nil {
			return nil, fmt.Errorf("ensure entity alias: %w", aliasErr)
		}
		resolvedAlias = &resolvedSubjectAlias{aliasId: alias.ID, subjectId: entityID}
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

// lockResolvedEntity key-share locks an existing entity's row before anything is written for it, and returns the
// entity to write for, following any redirect a merge left. Evidence references the alias, not the entity, so this
// lock rather than a foreign key keeps the entity from being merged away until the projection commits.
func (s *KnowledgeGraphIngestionService) lockResolvedEntity(ctx context.Context, entityID uuid.UUID) (uuid.UUID, error) {
	resolvedID, lockErr := ent.ResolveKnowledgeEntityID(ctx, s.db.Client(ctx), entityID)
	if lockErr != nil {
		return uuid.Nil, fmt.Errorf("lock entity: %w", lockErr)
	}
	return resolvedID, nil
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
		return nil, &invalidRelationshipRefError{schematypes.SkippedEvidenceReasonTopologyMismatch, fmt.Errorf("%w: resource already identifies an entity", errs.ErrConflict)}
	}
	rel, relationshipErr := alias.Edges.RelationshipOrErr()
	if relationshipErr != nil {
		return nil, fmt.Errorf("load aliased relationship: %w", relationshipErr)
	}
	if rel.Predicate != ref.Predicate || rel.SourceEntityID != sourceId || rel.TargetEntityID != targetId {
		return nil, &invalidRelationshipRefError{schematypes.SkippedEvidenceReasonTopologyMismatch, fmt.Errorf("%w: relationship alias identifies different topology", errs.ErrConflict)}
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: rel.ID}, nil
}

var knowledgeRelationshipUniqueColumns = sql.ConflictColumns(
	knr.FieldTenantID,
	knr.FieldPredicate,
	knr.FieldSourceEntityID,
	knr.FieldTargetEntityID,
)

type invalidRelationshipRefError struct {
	reason schematypes.SkippedEvidenceReason
	cause  error
}

func (e *invalidRelationshipRefError) Error() string { return e.cause.Error() }
func (e *invalidRelationshipRefError) Unwrap() error { return e.cause }

var errNoExistingEndpointEntityAlias = &invalidRelationshipRefError{schematypes.SkippedEvidenceReasonMissingEndpoint, fmt.Errorf("endpoint entity does not exist")}

func (s *KnowledgeGraphIngestionService) resolveRelationshipFromRef(ctx context.Context, ref rez.KnowledgeRelationshipRef, authority schematypes.KnowledgeAuthority, source, target *resolvedSubjectAlias) (*resolvedSubjectAlias, error) {
	// Endpoint resolution rejects an alias whose entity differs from the ref's category, so the ref's
	// categories are the resolved endpoints' categories.
	if validErr := knowledgegraph.ValidateRelationship(ref.Predicate, ref.Source.Category, ref.Target.Category); validErr != nil {
		return nil, &invalidRelationshipRefError{schematypes.SkippedEvidenceReasonPredicateCategories, validErr}
	}
	existing, lookupExistingErr := s.lookupExistingRelationshipAlias(ctx, ref, source.subjectId, target.subjectId)
	if lookupExistingErr != nil {
		return nil, fmt.Errorf("lookup existing alias: %w", lookupExistingErr)
	} else if existing != nil {
		return existing, nil
	}

	if ref.Predicate == knr.PredicateContains && source.subjectId == target.subjectId {
		return nil, &invalidRelationshipRefError{schematypes.SkippedEvidenceReasonSelfContains, fmt.Errorf("%w: an entity cannot contain itself", errs.ErrInvalidInput)}
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

	alias, aliasErr := s.ensureAlias(ctx, ref.ProviderResourceRef, authority, ksa.SubjectKindRelationship, uuid.Nil, relationshipID)
	if aliasErr != nil {
		return nil, aliasErr
	}
	return &resolvedSubjectAlias{aliasId: alias.ID, subjectId: relationshipID}, nil
}

var knowledgeEvidenceUniqueColumns = sql.ConflictColumns(ke.FieldTenantID, ke.FieldEventID, ke.FieldSubjectAliasID)

// resolveSubjectAliasFromRef resolves the subject's alias, creating any alias it needs with the authority.
func (s *KnowledgeGraphIngestionService) resolveSubjectAliasFromRef(ctx context.Context, ref rez.KnowledgeSubjectRef, authority schematypes.KnowledgeAuthority, resolveRelEndpoints bool) (*resolvedSubjectAlias, error) {
	if ref.Entity != nil {
		return s.resolveEntityFromRef(ctx, *ref.Entity, authority)
	}

	if refErr := ref.Relationship.ProviderResourceRef.Validate(); refErr != nil {
		return nil, fmt.Errorf("validate resource ref: %w", refErr)
	}

	var source *resolvedSubjectAlias
	var target *resolvedSubjectAlias
	var endpointErr error
	if resolveRelEndpoints {
		source, endpointErr = s.resolveEntityFromRef(ctx, ref.Relationship.Source, authority)
		if endpointErr != nil {
			return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
		}
		target, endpointErr = s.resolveEntityFromRef(ctx, ref.Relationship.Target, authority)
		if endpointErr != nil {
			return nil, fmt.Errorf("target: %w", errNoExistingEndpointEntityAlias)
		}
	} else {
		source, endpointErr = s.lookupExistingEntityAlias(ctx, ref.Relationship.Source)
		if endpointErr != nil {
			return nil, fmt.Errorf("source: %w", endpointErr)
		}
		if source == nil {
			return nil, fmt.Errorf("source: %w", errNoExistingEndpointEntityAlias)
		}
		target, endpointErr = s.lookupExistingEntityAlias(ctx, ref.Relationship.Target)
		if endpointErr != nil {
			return nil, fmt.Errorf("target: %w", endpointErr)
		}
		if target == nil {
			return nil, fmt.Errorf("target: %w", errNoExistingEndpointEntityAlias)
		}
		sourceID, lockErr := s.lockResolvedEntity(ctx, source.subjectId)
		if lockErr != nil {
			return nil, fmt.Errorf("source: %w", lockErr)
		}
		source.subjectId = sourceID
		targetID, lockErr := s.lockResolvedEntity(ctx, target.subjectId)
		if lockErr != nil {
			return nil, fmt.Errorf("target: %w", lockErr)
		}
		target.subjectId = targetID
	}

	return s.resolveRelationshipFromRef(ctx, *ref.Relationship, authority, source, target)
}

// ingestEvidenceRefs writes the refs' evidence and touches their subjects in refresh, for the caller to flush.
func (s *KnowledgeGraphIngestionService) ingestEvidenceRefs(ctx context.Context, event *ent.NormalizedEvent, refs []rez.KnowledgeEvidenceRef, refresh *subjectRefresh) error {
	eventID := event.ID
	if len(refs) == 0 {
		return nil
	}
	resolvedRefs := make([]*resolvedSubjectAlias, len(refs))
	aliasIDs := make([]uuid.UUID, len(refs))
	proposing := make([]*resolvedSubjectAlias, 0)
	for i, ref := range refs {
		resolved, aliasErr := s.resolveSubjectAliasFromRef(ctx, ref.Subject, ref.Authority, false)
		if aliasErr != nil {
			// Only the projection runner installs diagnostics; internal writes must reject invalid refs.
			diagnostics := projectionDiagnosticsFrom(ctx)
			var invalid *invalidRelationshipRefError
			if ref.Subject.Relationship != nil && diagnostics != nil && errors.As(aliasErr, &invalid) {
				if recordErr := diagnostics.record(ref, invalid.reason, aliasErr); recordErr != nil {
					return recordErr
				}
				continue
			}
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

	// An alias takes the highest authority among the event's refs to it, whichever of them is applied as evidence,
	// so the ref order within the event does not matter.
	authorities := make(map[uuid.UUID]schematypes.KnowledgeAuthority, len(refs))
	for i, ref := range refs {
		if resolved := resolvedRefs[i]; resolved != nil && ref.Authority.Compare(authorities[resolved.aliasId]) > 0 {
			authorities[resolved.aliasId] = ref.Authority
		}
	}
	touch := func(ref rez.KnowledgeEvidenceRef, resolved *resolvedSubjectAlias, reranked bool) {
		switch {
		case ref.Subject.Entity != nil:
			refresh.entities.Add(resolved.subjectId)
		case reranked:
			refresh.rerank(resolved.subjectId)
		default:
			refresh.relationships.Add(resolved.subjectId)
		}
	}

	evClient := s.db.Client(ctx).KnowledgeEvidence
	builders := make([]*ent.KnowledgeEvidenceCreate, 0, len(refs))
	for i, ref := range refs {
		resolved := resolvedRefs[i]
		if resolved == nil {
			continue
		}
		authority, pending := authorities[resolved.aliasId]
		if !pending {
			continue
		}
		delete(authorities, resolved.aliasId)
		if recorded.Contains(resolved.aliasId) {
			raised, reranked, raiseErr := s.raiseAliasAuthority(ctx, resolved.aliasId, authority)
			if raiseErr != nil {
				return fmt.Errorf("raise alias authority: %w", raiseErr)
			}
			if raised {
				touch(ref, resolved, reranked)
			}
			continue
		}
		evidence := knowledgegraph.AliasEvidence{
			EventID:     eventID,
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
		reranked, applyErr := s.applyToAlias(ctx, resolved.aliasId, evidence, authority)
		if applyErr != nil {
			return fmt.Errorf("apply evidence to alias: %w", applyErr)
		}
		touch(ref, resolved, reranked)
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
	if event.LastOccurredAt.After(event.OccurredAt) {
		// Resolution already holds these aliases' identity and entity locks. Catch up only this batch,
		// without taking new identity locks after ingestion's sorted acquisition.
		rows, queryErr := s.repeatableEvidenceQuery(ctx, eventID).
			Where(ke.SubjectAliasIDIn(aliasIDs...)).
			All(ctx)
		if queryErr != nil {
			return fmt.Errorf("load ingested repeatable evidence: %w", queryErr)
		}
		if repeatErr := s.applyReobservation(ctx, rows, event.LastOccurredAt, refresh); repeatErr != nil {
			return repeatErr
		}
	}
	if proposalsErr := s.createProposals(ctx, eventID, proposing); proposalsErr != nil {
		return fmt.Errorf("create knowledge proposals: %w", proposalsErr)
	}
	return nil
}

// applyToAlias applies one evidence row to its alias's state and observation times, and raises the alias to the
// evidence's authority when that is higher; a lower authority leaves it unchanged. It reports whether the change
// reorders a relationship alias's claim (aliasRanksClaim). The identity locks taken by the ingesting transaction
// cover the alias's resource ref, so no other ingestion updates the alias meanwhile.
func (s *KnowledgeGraphIngestionService) applyToAlias(ctx context.Context, aliasID uuid.UUID, evidence knowledgegraph.AliasEvidence, authority schematypes.KnowledgeAuthority) (bool, error) {
	client := s.db.Client(ctx)
	alias, queryErr := client.KnowledgeSubjectAlias.Get(ctx, aliasID)
	if queryErr != nil {
		return false, fmt.Errorf("load alias: %w", queryErr)
	}
	before := aliasRecord(alias)
	record := knowledgegraph.ApplyEvidence(before, evidence)
	updateAlias := client.KnowledgeSubjectAlias.UpdateOneID(aliasID).
		SetState(record.State)
	raised := authority.Compare(alias.Authority) > 0
	if raised {
		updateAlias.SetAuthority(authority)
	}
	if mutationErr := setTimeFields(updateAlias.Mutation(), map[string]*time.Time{
		ksa.FieldFirstObservedAt: record.FirstObservedAt,
		ksa.FieldLastObservedAt:  record.LastObservedAt,
		ksa.FieldWithdrawnAt:     record.WithdrawnAt,
		ksa.FieldExpiresAt:       record.ExpiresAt,
	}); mutationErr != nil {
		return false, mutationErr
	}
	if updateErr := updateAlias.Exec(ctx); updateErr != nil {
		return false, updateErr
	}
	return aliasRanksClaim(before, record, raised), nil
}

// raiseAliasAuthority raises an alias whose evidence for the event is already recorded, a replay or a second call
// for the event, to the authority when that is higher. It reports whether it raised the alias, and whether that
// reorders the alias's claim (aliasRanksClaim).
func (s *KnowledgeGraphIngestionService) raiseAliasAuthority(ctx context.Context, aliasID uuid.UUID, authority schematypes.KnowledgeAuthority) (bool, bool, error) {
	client := s.db.Client(ctx)
	alias, queryErr := client.KnowledgeSubjectAlias.Get(ctx, aliasID)
	if queryErr != nil {
		return false, false, fmt.Errorf("load alias: %w", queryErr)
	}
	if authority.Compare(alias.Authority) <= 0 {
		return false, false, nil
	}
	if updateErr := client.KnowledgeSubjectAlias.UpdateOneID(aliasID).SetAuthority(authority).Exec(ctx); updateErr != nil {
		return false, false, updateErr
	}
	record := aliasRecord(alias)
	return true, aliasRanksClaim(record, record, true), nil
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
		subjectAliases[i] = knowledgegraph.SubjectAlias{ID: alias.ID, Authority: alias.Authority, Record: aliasRecord(alias)}
	}
	return subjectAliases, nil
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

func (k *knowledgeEntityIdentityLockKeys) addAlias(ref rez.ProviderResourceRef) {
	k.keys = append(k.keys, k.makeKey("alias", ref.Provider, ref.ProviderNamespace, ref.ResourceRef))
}

func (k *knowledgeEntityIdentityLockKeys) addEntity(ref rez.KnowledgeEntityRef) error {
	values, valuesErr := getKnowledgeEntityLinkingValues(ref)
	if valuesErr != nil {
		return valuesErr
	}
	k.addAlias(ref.ProviderResourceRef)
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
			k.addAlias(subj.Relationship.ProviderResourceRef)
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

func (s *KnowledgeGraphIngestionService) repeatableEvidenceQuery(ctx context.Context, eventID uuid.UUID) *ent.KnowledgeEvidenceQuery {
	return s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(eventID), ke.KindEQ(ke.KindObserved), ke.ExpiresAtNotNil()).
		WithSubjectAlias(func(q *ent.KnowledgeSubjectAliasQuery) { q.WithRelationship() })
}

// reobserveEvidence runs under the event's projection lock, before any identity locks in the transaction.
func (s *KnowledgeGraphIngestionService) reobserveEvidence(ctx context.Context, eventID uuid.UUID, at time.Time) error {
	query := s.repeatableEvidenceQuery(ctx, eventID)
	rows, queryErr := query.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("load repeatable evidence: %w", queryErr)
	}
	if len(rows) == 0 {
		return nil
	}
	ik := &knowledgeEntityIdentityLockKeys{}
	lockEntityIDs := mapset.NewSet[uuid.UUID]()
	for _, row := range rows {
		alias := row.Edges.SubjectAlias
		ik.addAlias(rez.ProviderResourceRef{Provider: alias.Provider,
			ProviderNamespace: alias.ProviderNamespace, ResourceRef: alias.ProviderResourceRef})
		if alias.EntityID != nil {
			lockEntityIDs.Add(*alias.EntityID)
		} else {
			relationship := alias.Edges.Relationship
			if relationship == nil {
				return errs.ErrKnowledgeSubjectMerged
			}
			lockEntityIDs.Add(relationship.SourceEntityID)
			lockEntityIDs.Add(relationship.TargetEntityID)
		}
	}
	// Alias keys use ingestion's derivation, including relationship aliases. Take the complete set once,
	// after the projection lock and before entity locks, so concurrent ingestion cannot overwrite repeats.
	if lockErr := s.db.AcquireTxLocks(ctx, knowledgeEntityIdentityLockNamespace, ik.keys...); lockErr != nil {
		return lockErr
	}
	// As in projection, key-share locks prevent merges from moving these aliases or endpoints. A merge that
	// committed since the first read left a redirect, which lockResolvedEntity follows to the survivor.
	lockedEntityIDs := mapset.NewSet[uuid.UUID]()
	for _, entityID := range sortedUUIDs(lockEntityIDs) {
		lockedID, lockErr := s.lockResolvedEntity(ctx, entityID)
		if lockErr != nil {
			return lockErr
		}
		lockedEntityIDs.Add(lockedID)
	}
	// Reload after identity locks: a later repeat or a merge may have changed the rows while waiting.
	rows, queryErr = query.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("reload repeatable evidence: %w", queryErr)
	}
	for _, row := range rows {
		alias := row.Edges.SubjectAlias
		if alias.EntityID != nil {
			if !lockedEntityIDs.Contains(*alias.EntityID) {
				return errs.ErrKnowledgeSubjectMerged
			}
		} else if relationship := alias.Edges.Relationship; relationship == nil ||
			!lockedEntityIDs.Contains(relationship.SourceEntityID) || !lockedEntityIDs.Contains(relationship.TargetEntityID) {
			return errs.ErrKnowledgeSubjectMerged
		}
	}
	refresh := newSubjectRefresh()
	if repeatErr := s.applyReobservation(ctx, rows, at, refresh); repeatErr != nil {
		return repeatErr
	}
	return s.flush(ctx, refresh)
}

// applyReobservation touches its subjects in refresh, for the caller to flush.
func (s *KnowledgeGraphIngestionService) applyReobservation(ctx context.Context, rows []*ent.KnowledgeEvidence, at time.Time, refresh *subjectRefresh) error {
	client := s.db.Client(ctx)
	for _, row := range rows {
		alias := row.Edges.SubjectAlias
		if row.ReobservedAt != nil && !row.ReobservedAt.Before(at) {
			continue
		}
		updateEvidence := client.KnowledgeEvidence.UpdateOneID(row.ID).
			SetReobservedAt(at)
		if updateErr := updateEvidence.Exec(ctx); updateErr != nil {
			return updateErr
		}
		original := knowledgegraph.AliasEvidence{EventID: row.EventID, EffectiveAt: row.EffectiveAt, ExpiresAt: row.ExpiresAt}
		repeat := knowledgegraph.AliasEvidence{Kind: ke.KindObserved, EffectiveAt: at, ReobservationOf: &original}
		// A repeat restates the original evidence, so it carries no authority of its own.
		reranked, applyErr := s.applyToAlias(ctx, row.SubjectAliasID, repeat, alias.Authority)
		if applyErr != nil {
			return applyErr
		}
		switch {
		case alias.EntityID != nil:
			refresh.entities.Add(*alias.EntityID)
		case reranked:
			refresh.rerank(*alias.RelationshipID)
		default:
			refresh.relationships.Add(*alias.RelationshipID)
		}
	}
	return nil
}
