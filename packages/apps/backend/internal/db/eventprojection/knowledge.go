package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

// ingestSubjectEvidence ingests the refs and returns the alias the event's evidence was recorded on. It returns
// nil when the evidence is a relationship ref that ingestion skipped as invalid: the alias for that resource is
// then missing, or still identifies the relationship an earlier event claimed.
func (s *ProjectionService) ingestSubjectEvidence(ctx context.Context, event *ent.NormalizedEvent, evidence rez.KnowledgeEvidenceRef, supportingEvidence ...rez.KnowledgeEvidenceRef) (*ent.KnowledgeSubjectAlias, error) {
	refs := append(supportingEvidence, evidence)
	if ingestErr := s.knowledge.IngestEvidence(ctx, event, refs...); ingestErr != nil {
		return nil, fmt.Errorf("ingest evidence: %w", ingestErr)
	}
	var ref rez.ProviderResourceRef
	subj := evidence.Subject
	if subj.Entity != nil {
		ref = subj.Entity.ProviderResourceRef
	} else if subj.Relationship != nil {
		ref = subj.Relationship.ProviderResourceRef
	}
	query := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(
			ksa.Provider(ref.Provider),
			ksa.ProviderNamespace(ref.ProviderNamespace),
			ksa.ProviderResourceRef(ref.ResourceRef),
			ksa.HasEvidenceWith(ke.EventID(event.ID)),
		).
		WithEntity().
		WithRelationship()
	alias, queryAliasErr := query.Only(ctx)
	if queryAliasErr != nil {
		if ent.IsNotFound(queryAliasErr) && subj.Relationship != nil {
			return nil, nil
		}
		return nil, fmt.Errorf("query alias: %w", queryAliasErr)
	}
	return alias, nil
}
