package ent

import (
	"github.com/firebase/genkit/go/ai"
	vc "github.com/rezible/rezible/ent/videoconference"
)

func (ims IncidentMilestones) GetLatest() *IncidentMilestone {
	var latest *IncidentMilestone
	for _, im := range ims {
		if latest == nil || latest.Timestamp.After(im.Timestamp) {
			latest = im
		}
	}
	return latest
}

func (ie IncidentEdges) GetLatestMilestone() *IncidentMilestone {
	return IncidentMilestones(ie.Milestones).GetLatest()
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
	conferences, confErr := ie.VideoConferencesOrErr()
	if confErr != nil || len(conferences) == 0 {
		return nil
	}
	return VideoConferences(conferences).GetPrimary()
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
