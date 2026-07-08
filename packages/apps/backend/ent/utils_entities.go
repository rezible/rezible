package ent

import (
	"time"

	ke "github.com/rezible/rezible/ent/knowledgeentity"
	kr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
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

func (ev *NormalizedEvent) DeriveObservedAt() time.Time {
	if !ev.OccurredAt.IsZero() {
		return ev.OccurredAt
	}
	if !ev.ReceivedAt.IsZero() {
		return ev.ReceivedAt
	}
	return time.Now()
}

func (ev *NormalizedEvent) MakeSubjectAliasRef(kind ksa.SubjectKind) KnowledgeSubjectAliasRef {
	return KnowledgeSubjectAliasRef{
		Kind:               kind,
		Provider:           ev.Provider,
		ProviderSubjectRef: ev.ProviderSubjectRef,
	}
}

type KnowledgeEntityRef struct {
	Kind        string
	Reference   string
	DisplayName string
	Description string
}

func (r *KnowledgeEntityRef) Predicate() predicate.KnowledgeEntity {
	return ke.And(ke.Kind(r.Kind), ke.Reference(r.Reference))
}

type KnowledgeRelationshipRef struct {
	Kind        string
	Description string
	EntityRefs  [2]KnowledgeEntityRef
}

func (r *KnowledgeRelationshipRef) Predicate() predicate.KnowledgeRelationship {
	return kr.And(
		kr.Kind(r.Kind),
		kr.HasSourceEntityWith(r.EntityRefs[0].Predicate()),
		kr.HasTargetEntityWith(r.EntityRefs[1].Predicate()))
}

type KnowledgeSubjectAliasRef struct {
	Kind               ksa.SubjectKind
	Provider           string
	ProviderSubjectRef string
}

func (r *KnowledgeSubjectAliasRef) Predicate() predicate.KnowledgeSubjectAlias {
	return ksa.And(
		ksa.SubjectKindEQ(r.Kind),
		ksa.Provider(r.Provider),
		ksa.ProviderSubjectRef(r.ProviderSubjectRef))
}

func (r *KnowledgeSubjectAliasRef) SortKey() string {
	return r.Provider + "\x1f" + r.ProviderSubjectRef
}
