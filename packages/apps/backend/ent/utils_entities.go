package ent

import (
	"fmt"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
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

type (
	ProviderResourceRef struct {
		Provider          string `json:"provider"`
		ProviderNamespace string `json:"provider_namespace"`
		ResourceRef       string `json:"resource_ref"`
	}

	KnowledgeEntityRef struct {
		Category            kne.Category
		Kind                string
		ProviderResourceRef ProviderResourceRef
	}

	KnowledgeRelationshipRef struct {
		Predicate           knr.Predicate
		ProviderResourceRef ProviderResourceRef
		Source              KnowledgeEntityRef
		Target              KnowledgeEntityRef
	}

	KnowledgeEvidenceRef struct {
		Kind                kev.Kind
		Assertion           string
		EffectiveAt         time.Time
		SubjectState        schematypes.KnowledgeGraphSubjectState
		SubjectEntity       *KnowledgeEntityRef
		SubjectRelationship *KnowledgeRelationshipRef
	}
)

func (ref ProviderResourceRef) Validate() error {
	if strings.TrimSpace(ref.Provider) == "" {
		return fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(ref.ResourceRef) == "" {
		return fmt.Errorf("resource_ref is required")
	}
	if ref.Provider != "rezible" && strings.TrimSpace(ref.ProviderNamespace) == "" {
		return fmt.Errorf("provider_namespace is required for provider %q", ref.Provider)
	}
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
