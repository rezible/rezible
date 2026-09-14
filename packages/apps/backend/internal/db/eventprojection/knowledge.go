package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

const (
	knowledgeEntityKindAlert      = "alert"
	knowledgeEntityKindCodeChange = "code_change"
	knowledgeEntityKindIncident   = "incident"
	knowledgeEntityKindRepository = "repository"
	knowledgeEntityKindTeam       = "team"
	knowledgeEntityKindUser       = "user"
)

func (s *ProjectionService) ingestEntityEvidence(ctx context.Context, evt *ent.NormalizedEvent, evRef rez.KnowledgeEvidenceRef) (uuid.UUID, error) {
	subj, ingestErr := s.ingestSubjectEvidence(ctx, evt, evRef)
	if ingestErr != nil {
		return uuid.Nil, fmt.Errorf("incident knowledge evidence: %w", ingestErr)
	} else if subj.EntityID == nil {
		return uuid.Nil, fmt.Errorf("nil subject entity")
	}
	return *subj.EntityID, nil
}

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
		).
		WithEntity().
		WithRelationship()
	alias, queryAliasErr := query.Only(ctx)
	if queryAliasErr != nil {
		return nil, fmt.Errorf("query alias: %w", queryAliasErr)
	}
	return alias, nil
}
