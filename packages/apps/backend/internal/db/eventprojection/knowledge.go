package eventprojection

import (
	"context"

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

func (s *ProjectionService) ingestSubjectEvidence(ctx context.Context, event *ent.NormalizedEvent, evidence rez.KnowledgeEvidenceRef, supportingEvidence ...rez.KnowledgeEvidenceRef) (*ent.KnowledgeSubjectAlias, error) {
	refs := append(supportingEvidence, evidence)
	if err := s.knowledge.IngestEvidence(ctx, event, refs...); err != nil {
		return nil, err
	}
	var ref rez.ProviderResourceRef
	subj := evidence.Subject
	if subj.Entity != nil {
		ref = subj.Entity.ProviderResourceRef
	} else if subj.Relationship != nil {
		ref = subj.Relationship.ProviderResourceRef
	}
	query := s.db.Client(ctx).KnowledgeSubjectAlias.Query()
	query.Where(
		ksa.Provider(ref.Provider),
		ksa.ProviderNamespace(ref.ProviderNamespace),
		ksa.ProviderResourceRef(ref.ResourceRef),
	)
	query.WithEntity()
	query.WithRelationship()
	return query.Only(ctx)
}
