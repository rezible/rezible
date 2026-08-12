package ent

import (
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
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

func (ev *NormalizedEvent) KnowledgeAliasRef() KnowledgeAliasRef {
	return KnowledgeAliasRef{
		Provider:           ev.Provider,
		ProviderSource:     ev.ProviderSource,
		ProviderSubjectRef: ev.ProviderSubjectRef,
	}
}

type KnowledgeAliasRef struct {
	Provider           string
	ProviderSource     string
	ProviderSubjectRef string
}

func (a KnowledgeAliasRef) SubjectPredicate(kind ksa.SubjectKind) predicate.KnowledgeSubjectAlias {
	return ksa.And(
		ksa.SubjectKindEQ(kind),
		ksa.Provider(a.Provider),
		ksa.ProviderSource(a.ProviderSource),
		ksa.ProviderSubjectRef(a.ProviderSubjectRef))
}

func (a KnowledgeAliasRef) LockKey(kind ksa.SubjectKind) string {
	return fmt.Sprintf("%s:%s:%s:%s", kind, a.Provider, a.ProviderSource, a.ProviderSubjectRef)
}

type (
	KnowledgeEntityRef struct {
		Kind  string
		Alias KnowledgeAliasRef
	}

	KnowledgeRelationshipRef struct {
		Kind   string
		Alias  KnowledgeAliasRef
		Source KnowledgeEntityRef
		Target KnowledgeEntityRef
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

func (aliases KnowledgeSubjectAliasSlice) LatestEvidence() *KnowledgeEvidence {
	var latest *KnowledgeEvidence
	for _, alias := range aliases {
		for _, ev := range alias.Edges.Evidence {
			if latest == nil || ev.EffectiveAt.After(latest.EffectiveAt) {
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
