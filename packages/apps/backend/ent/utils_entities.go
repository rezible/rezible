package ent

import (
	"bytes"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incidentmilestone"
	invfvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	vc "github.com/rezible/rezible/ent/videoconference"
)

// GetLatest returns the newest milestone of the given kinds, or of any kind when none are given. Equal
// timestamps are ordered by ID.
func (ims IncidentMilestones) GetLatest(kinds ...incidentmilestone.Kind) *IncidentMilestone {
	var latest *IncidentMilestone
	for _, im := range ims {
		if len(kinds) > 0 && !slices.Contains(kinds, im.Kind) {
			continue
		}
		if latest == nil || im.Timestamp.After(latest.Timestamp) ||
			(im.Timestamp.Equal(latest.Timestamp) && bytes.Compare(im.ID[:], latest.ID[:]) > 0) {
			latest = im
		}
	}
	return latest
}

func (vcs VideoConferences) GetPrimary() *VideoConference {
	var active *VideoConference
	var latest *VideoConference
	for _, conference := range vcs {
		if latest == nil || conference.CreatedAt.After(latest.CreatedAt) {
			latest = conference
		}
		if conference.Status == vc.StatusActive {
			if active == nil || conference.CreatedAt.After(active.CreatedAt) {
				active = conference
			}
		}
	}
	if active != nil {
		return active
	}
	if latest != nil {
		return latest
	}
	return nil
}

func (ie IncidentEdges) GetPrimaryVideoConference() *VideoConference {
	//conferences, confErr := ie.VideoConferencesOrErr()
	//if confErr != nil || len(conferences) == 0 {
	//	return nil
	//}
	//return VideoConferences(conferences).GetPrimary()
	return nil
}

func (aliases KnowledgeSubjectAliasSlice) LatestEvidence() *KnowledgeEvidence {
	var latest *KnowledgeEvidence
	for _, alias := range aliases {
		for _, ev := range alias.Edges.Evidence {
			if latest == nil || ev.EffectiveAt.After(latest.EffectiveAt) {
				latest = ev
				continue
			}
			if ev.EffectiveAt.Before(latest.EffectiveAt) || ev.CreatedAt.Before(latest.CreatedAt) {
				continue
			}
			if ev.CreatedAt.After(latest.CreatedAt) || ev.ID.String() > latest.ID.String() {
				latest = ev
			}
		}
	}
	return latest
}

func (e *KnowledgeEntity) LatestEvidence() *KnowledgeEvidence {
	return KnowledgeSubjectAliasSlice(e.Edges.Aliases).LatestEvidence()
}

func (r *KnowledgeRelationship) LatestEvidence() *KnowledgeEvidence {
	return KnowledgeSubjectAliasSlice(r.Edges.Aliases).LatestEvidence()
}

func (am *AgentMessage) MakeGenkitMessage() *ai.Message {
	return ai.NewMessage(ai.Role(am.Role), am.Metadata, am.Content...)
}

// CitedIDs returns the knowledge evidence or provider read each reference cites, in loaded order.
func (refs InvestigationOutputReferences) CitedIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		if ref.KnowledgeEvidenceID != nil {
			ids = append(ids, *ref.KnowledgeEvidenceID)
		} else if ref.AgentProviderReadID != nil {
			ids = append(ids, *ref.AgentProviderReadID)
		}
	}
	return ids
}

// KnowledgeEvidenceIDs returns the evidence IDs cited by the references, in loaded order.
func (refs InvestigationOutputReferences) KnowledgeEvidenceIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		if ref.KnowledgeEvidenceID != nil {
			ids = append(ids, *ref.KnowledgeEvidenceID)
		}
	}
	return ids
}

// InvalidatedByVersionIDs returns the source versions of the loaded incoming invalidation links.
func (v *InvestigationFindingVersion) InvalidatedByVersionIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(v.Edges.IncomingLinks))
	for _, link := range v.Edges.IncomingLinks {
		if link.Relation == invfvl.RelationInvalidates {
			ids = append(ids, link.SourceVersionID)
		}
	}
	return ids
}

// CurrentAnswerVersion returns the first loaded version of the input's answer finding.
// Investigation services load answer versions newest first.
func (ui *InvestigationUserInput) CurrentAnswerVersion() *InvestigationFindingVersion {
	for _, finding := range ui.Edges.Findings {
		if len(finding.Edges.Versions) > 0 {
			return finding.Edges.Versions[0]
		}
	}
	return nil
}
