package db

import (
	"context"
	"fmt"
	"maps"
	"unicode/utf8"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
)

const knowledgeProposalRationaleMaxLen = 1000

// knowledgeProposalDraft is a proposal resolution found, written once the evidence it cites exists.
type knowledgeProposalDraft struct {
	category  kne.Category
	kind      string
	attribute string
	valueA    string
	valueB    string
	origin    kp.Origin
	rationale string
}

func (r *resolvedSubjectAlias) propose(ref rez.KnowledgeEntityRef, attribute, first, second string, origin kp.Origin, rationale string) {
	if second < first {
		first, second = second, first
	}
	if len(rationale) > knowledgeProposalRationaleMaxLen {
		cut := knowledgeProposalRationaleMaxLen - len("…")
		for !utf8.RuneStart(rationale[cut]) {
			cut--
		}
		rationale = rationale[:cut] + "…"
	}
	r.proposals = append(r.proposals, knowledgeProposalDraft{
		category:  ref.Category,
		kind:      ref.Kind,
		attribute: attribute,
		valueA:    first,
		valueB:    second,
		origin:    origin,
		rationale: rationale,
	})
}

// proposeIdentityConflicts handles the ref's linking values held by an entity other than its existing alias's.
// For kinds a merge can combine, the alias wins: each such value is proposed as a match for the alias entity's
// own value and left off the entity. It returns the values that do not conflict.
func (s *KnowledgeGraphIngestionService) proposeIdentityConflicts(ctx context.Context, ref rez.KnowledgeEntityRef, resolved *resolvedSubjectAlias, values map[string]string, linkingAttributes []*ent.KnowledgeEntityLinkingAttribute) (map[string]string, error) {
	unconflicted := maps.Clone(values)
	for _, linkingAttribute := range linkingAttributes {
		if linkingAttribute.EntityID == resolved.subjectId {
			continue
		}
		// A kind a confirmed match can merge has an identity conflict proposed instead of failing.
		if !knowledgegraph.IsMergeable(ref.Category, ref.Kind) {
			return nil, fmt.Errorf("%w: linking attribute identifies another entity", errs.ErrConflict)
		}
		other, entityErr := linkingAttribute.Edges.EntityOrErr()
		if entityErr != nil {
			return nil, fmt.Errorf("load linked entity: %w", entityErr)
		}
		if other.Category != ref.Category || other.Kind != ref.Kind {
			return nil, fmt.Errorf("%w: linking attribute identifies %q/%q, evidence expects %q/%q",
				errs.ErrConflict, other.Category, other.Kind, ref.Category, ref.Kind)
		}

		ownValueQuery := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
			Where(kela.EntityID(resolved.subjectId), kela.Attribute(linkingAttribute.Attribute)).
			Order(kela.ByValue())
		ownValue, ownErr := ownValueQuery.First(ctx)
		if ownErr != nil {
			if ent.IsNotFound(ownErr) {
				return nil, fmt.Errorf("%w: linking attribute identifies another entity", errs.ErrConflict)
			}
			return nil, fmt.Errorf("query alias entity linking attribute: %w", ownErr)
		}

		rationale := fmt.Sprintf("`%s` now names the %s `%s` as `%s`, which another %s already has.",
			ref.ProviderResourceRef.Provider, ref.Kind, ownValue.Value, linkingAttribute.Value, ref.Kind)
		resolved.propose(ref, linkingAttribute.Attribute, ownValue.Value, linkingAttribute.Value, kp.OriginIdentityConflict, rationale)
		delete(unconflicted, linkingAttribute.Attribute)
	}
	return unconflicted, nil
}

// proposeSimilarServiceNames proposes a match between a service name new to the organization and every other
// service's name that differs from it only by a generic term.
func (s *KnowledgeGraphIngestionService) proposeSimilarServiceNames(ctx context.Context, ref rez.KnowledgeEntityRef, resolved *resolvedSubjectAlias, name string) error {
	stem := projections.ServiceNameStem(name)
	if stem == "" {
		return nil
	}
	forms := []string{stem}
	for _, term := range projections.GenericServiceNameTerms {
		forms = append(forms, term+"-"+stem, stem+"-"+term)
	}

	similarQuery := s.db.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(
			kela.Attribute(projections.LinkingAttributeServiceName),
			kela.ValueIn(forms...),
			kela.EntityIDNEQ(resolved.subjectId),
			kela.HasEntityWith(kne.CategoryEQ(ref.Category), kne.Kind(ref.Kind)),
		).
		Order(kela.ByValue())
	similar, queryErr := similarQuery.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query similar service names: %w", queryErr)
	}

	for _, linkingAttribute := range similar {
		reason, isSimilar := projections.SimilarServiceNames(name, linkingAttribute.Value)
		if !isSimilar {
			continue
		}
		rationale := reason + ", so they may name the same service."
		resolved.propose(ref, projections.LinkingAttributeServiceName, name, linkingAttribute.Value, kp.OriginNameSimilarity, rationale)
	}
	return nil
}

var knowledgeProposalUniqueColumns = sql.ConflictColumns(kp.FieldTenantID, kp.FieldAttribute, kp.FieldValueA, kp.FieldValueB)

// createProposals writes the proposals resolution found, each citing the event's evidence on its alias. A pair
// already proposed, in any status, is skipped: a constraint error would abort the transaction and lose the event.
func (s *KnowledgeGraphIngestionService) createProposals(ctx context.Context, eventID uuid.UUID, resolved []*resolvedSubjectAlias) error {
	if len(resolved) == 0 {
		return nil
	}
	client := s.db.Client(ctx)
	builders := make([]*ent.KnowledgeProposalCreate, 0, len(resolved))
	for _, alias := range resolved {
		evidenceQuery := client.KnowledgeEvidence.Query().
			Where(ke.EventID(eventID), ke.SubjectAliasID(alias.aliasId))
		evidenceID, evidenceErr := evidenceQuery.OnlyID(ctx)
		if evidenceErr != nil {
			return fmt.Errorf("query proposal evidence: %w", evidenceErr)
		}
		for _, draft := range alias.proposals {
			builders = append(builders, client.KnowledgeProposal.Create().
				SetCategory(kp.Category(draft.category)).
				SetKind(draft.kind).
				SetAttribute(draft.attribute).
				SetValueA(draft.valueA).
				SetValueB(draft.valueB).
				SetOrigin(draft.origin).
				SetRationale(draft.rationale).
				SetEvidenceID(evidenceID))
		}
	}
	createProposals := client.KnowledgeProposal.CreateBulk(builders...).
		OnConflict(knowledgeProposalUniqueColumns).
		DoNothing()
	return createProposals.Exec(ctx)
}
