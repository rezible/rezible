package ent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	"github.com/rezible/rezible/ent/predicate"
	vc "github.com/rezible/rezible/ent/videoconference"
)

func (ims IncidentMilestones) GetLatest() *IncidentMilestone {
	if len(ims) == 0 {
		return nil
	}
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

func (ev *NormalizedEvent) MakeEntityAliasRef() KnowledgeEntityAliasRef {
	return KnowledgeEntityAliasRef{Provider: ev.Provider, ProviderSubjectRef: ev.ProviderSubjectRef}
}

type KnowledgeEntityAliasRef struct {
	Provider           string
	ProviderSubjectRef string
}

func (ref KnowledgeEntityAliasRef) Predicate() predicate.KnowledgeEntityAlias {
	return knea.And(knea.Provider(ref.Provider), knea.ProviderSubjectRef(ref.ProviderSubjectRef))
}

func (ref KnowledgeEntityAliasRef) SortKey() string {
	return ref.Provider + "\x1f" + ref.ProviderSubjectRef
}

func (r *AgentRun) GetSubjectEntityId(subjectKind string) (uuid.UUID, error) {
	subjects, subjectsErr := r.Edges.SubjectsOrErr()
	if subjectsErr != nil {
		return uuid.Nil, subjectsErr
	}
	for _, sub := range subjects {
		if sub.SubjectKind == subjectKind {
			if sub.DomainEntityID == nil {
				return uuid.Nil, fmt.Errorf("subject kind with nil domain entity id")
			}
			return *sub.DomainEntityID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("subject kind %s not found", subjectKind)
}
