package ent

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	invfvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	vc "github.com/rezible/rezible/ent/videoconference"
)

func (ims IncidentMilestones) GetLatest() *IncidentMilestone {
	var latest *IncidentMilestone
	for _, im := range ims {
		if latest == nil || latest.Timestamp.IsZero() || im.Timestamp.After(latest.Timestamp) {
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

func (sogs SituationObservationGroups) SignalGroupCounts() map[string]int {
	eventCounts := mapset.NewSet[uuid.UUID]()
	alertEpCounts := mapset.NewSet[uuid.UUID]()
	for _, group := range sogs {
		for _, event := range group.Edges.Events {
			eventCounts.Add(event.ID)
		}
		for _, ep := range group.Edges.AlertEpisodes {
			alertEpCounts.Add(ep.ID)
		}
	}
	return map[string]int{
		"event":         eventCounts.Cardinality(),
		"alert_episode": alertEpCounts.Cardinality(),
	}
}

func (sogs SituationObservationGroups) SignalCount() int {
	total := 0
	for _, count := range sogs.SignalGroupCounts() {
		total += count
	}
	return total
}

// KnowledgeEvidenceIDs returns the evidence IDs cited by the references, in loaded order.
func (refs InvestigationOutputReferences) KnowledgeEvidenceIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(refs))
	for i, ref := range refs {
		ids[i] = ref.KnowledgeEvidenceID
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
