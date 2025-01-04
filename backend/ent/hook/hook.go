// Code generated by ent, DO NOT EDIT.

package hook

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent"
)

// The EnvironmentFunc type is an adapter to allow the use of ordinary
// function as Environment mutator.
type EnvironmentFunc func(context.Context, *ent.EnvironmentMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f EnvironmentFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.EnvironmentMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.EnvironmentMutation", m)
}

// The FunctionalityFunc type is an adapter to allow the use of ordinary
// function as Functionality mutator.
type FunctionalityFunc func(context.Context, *ent.FunctionalityMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f FunctionalityFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.FunctionalityMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.FunctionalityMutation", m)
}

// The IncidentFunc type is an adapter to allow the use of ordinary
// function as Incident mutator.
type IncidentFunc func(context.Context, *ent.IncidentMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentMutation", m)
}

// The IncidentDebriefFunc type is an adapter to allow the use of ordinary
// function as IncidentDebrief mutator.
type IncidentDebriefFunc func(context.Context, *ent.IncidentDebriefMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentDebriefFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentDebriefMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentDebriefMutation", m)
}

// The IncidentDebriefMessageFunc type is an adapter to allow the use of ordinary
// function as IncidentDebriefMessage mutator.
type IncidentDebriefMessageFunc func(context.Context, *ent.IncidentDebriefMessageMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentDebriefMessageFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentDebriefMessageMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentDebriefMessageMutation", m)
}

// The IncidentDebriefQuestionFunc type is an adapter to allow the use of ordinary
// function as IncidentDebriefQuestion mutator.
type IncidentDebriefQuestionFunc func(context.Context, *ent.IncidentDebriefQuestionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentDebriefQuestionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentDebriefQuestionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentDebriefQuestionMutation", m)
}

// The IncidentDebriefSuggestionFunc type is an adapter to allow the use of ordinary
// function as IncidentDebriefSuggestion mutator.
type IncidentDebriefSuggestionFunc func(context.Context, *ent.IncidentDebriefSuggestionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentDebriefSuggestionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentDebriefSuggestionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentDebriefSuggestionMutation", m)
}

// The IncidentEventFunc type is an adapter to allow the use of ordinary
// function as IncidentEvent mutator.
type IncidentEventFunc func(context.Context, *ent.IncidentEventMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentEventFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentEventMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentEventMutation", m)
}

// The IncidentEventContextFunc type is an adapter to allow the use of ordinary
// function as IncidentEventContext mutator.
type IncidentEventContextFunc func(context.Context, *ent.IncidentEventContextMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentEventContextFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentEventContextMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentEventContextMutation", m)
}

// The IncidentEventContributingFactorFunc type is an adapter to allow the use of ordinary
// function as IncidentEventContributingFactor mutator.
type IncidentEventContributingFactorFunc func(context.Context, *ent.IncidentEventContributingFactorMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentEventContributingFactorFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentEventContributingFactorMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentEventContributingFactorMutation", m)
}

// The IncidentEventEvidenceFunc type is an adapter to allow the use of ordinary
// function as IncidentEventEvidence mutator.
type IncidentEventEvidenceFunc func(context.Context, *ent.IncidentEventEvidenceMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentEventEvidenceFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentEventEvidenceMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentEventEvidenceMutation", m)
}

// The IncidentFieldFunc type is an adapter to allow the use of ordinary
// function as IncidentField mutator.
type IncidentFieldFunc func(context.Context, *ent.IncidentFieldMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentFieldFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentFieldMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentFieldMutation", m)
}

// The IncidentFieldOptionFunc type is an adapter to allow the use of ordinary
// function as IncidentFieldOption mutator.
type IncidentFieldOptionFunc func(context.Context, *ent.IncidentFieldOptionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentFieldOptionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentFieldOptionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentFieldOptionMutation", m)
}

// The IncidentLinkFunc type is an adapter to allow the use of ordinary
// function as IncidentLink mutator.
type IncidentLinkFunc func(context.Context, *ent.IncidentLinkMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentLinkFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentLinkMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentLinkMutation", m)
}

// The IncidentMilestoneFunc type is an adapter to allow the use of ordinary
// function as IncidentMilestone mutator.
type IncidentMilestoneFunc func(context.Context, *ent.IncidentMilestoneMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentMilestoneFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentMilestoneMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentMilestoneMutation", m)
}

// The IncidentRoleFunc type is an adapter to allow the use of ordinary
// function as IncidentRole mutator.
type IncidentRoleFunc func(context.Context, *ent.IncidentRoleMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentRoleFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentRoleMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentRoleMutation", m)
}

// The IncidentRoleAssignmentFunc type is an adapter to allow the use of ordinary
// function as IncidentRoleAssignment mutator.
type IncidentRoleAssignmentFunc func(context.Context, *ent.IncidentRoleAssignmentMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentRoleAssignmentFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentRoleAssignmentMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentRoleAssignmentMutation", m)
}

// The IncidentSeverityFunc type is an adapter to allow the use of ordinary
// function as IncidentSeverity mutator.
type IncidentSeverityFunc func(context.Context, *ent.IncidentSeverityMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentSeverityFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentSeverityMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentSeverityMutation", m)
}

// The IncidentTagFunc type is an adapter to allow the use of ordinary
// function as IncidentTag mutator.
type IncidentTagFunc func(context.Context, *ent.IncidentTagMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentTagFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentTagMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentTagMutation", m)
}

// The IncidentTeamAssignmentFunc type is an adapter to allow the use of ordinary
// function as IncidentTeamAssignment mutator.
type IncidentTeamAssignmentFunc func(context.Context, *ent.IncidentTeamAssignmentMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentTeamAssignmentFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentTeamAssignmentMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentTeamAssignmentMutation", m)
}

// The IncidentTypeFunc type is an adapter to allow the use of ordinary
// function as IncidentType mutator.
type IncidentTypeFunc func(context.Context, *ent.IncidentTypeMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f IncidentTypeFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.IncidentTypeMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.IncidentTypeMutation", m)
}

// The MeetingScheduleFunc type is an adapter to allow the use of ordinary
// function as MeetingSchedule mutator.
type MeetingScheduleFunc func(context.Context, *ent.MeetingScheduleMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f MeetingScheduleFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.MeetingScheduleMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.MeetingScheduleMutation", m)
}

// The MeetingSessionFunc type is an adapter to allow the use of ordinary
// function as MeetingSession mutator.
type MeetingSessionFunc func(context.Context, *ent.MeetingSessionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f MeetingSessionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.MeetingSessionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.MeetingSessionMutation", m)
}

// The OncallAlertFunc type is an adapter to allow the use of ordinary
// function as OncallAlert mutator.
type OncallAlertFunc func(context.Context, *ent.OncallAlertMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallAlertFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallAlertMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallAlertMutation", m)
}

// The OncallAlertInstanceFunc type is an adapter to allow the use of ordinary
// function as OncallAlertInstance mutator.
type OncallAlertInstanceFunc func(context.Context, *ent.OncallAlertInstanceMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallAlertInstanceFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallAlertInstanceMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallAlertInstanceMutation", m)
}

// The OncallHandoverTemplateFunc type is an adapter to allow the use of ordinary
// function as OncallHandoverTemplate mutator.
type OncallHandoverTemplateFunc func(context.Context, *ent.OncallHandoverTemplateMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallHandoverTemplateFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallHandoverTemplateMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallHandoverTemplateMutation", m)
}

// The OncallRosterFunc type is an adapter to allow the use of ordinary
// function as OncallRoster mutator.
type OncallRosterFunc func(context.Context, *ent.OncallRosterMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallRosterFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallRosterMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallRosterMutation", m)
}

// The OncallScheduleFunc type is an adapter to allow the use of ordinary
// function as OncallSchedule mutator.
type OncallScheduleFunc func(context.Context, *ent.OncallScheduleMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallScheduleFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallScheduleMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallScheduleMutation", m)
}

// The OncallScheduleParticipantFunc type is an adapter to allow the use of ordinary
// function as OncallScheduleParticipant mutator.
type OncallScheduleParticipantFunc func(context.Context, *ent.OncallScheduleParticipantMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallScheduleParticipantFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallScheduleParticipantMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallScheduleParticipantMutation", m)
}

// The OncallUserShiftFunc type is an adapter to allow the use of ordinary
// function as OncallUserShift mutator.
type OncallUserShiftFunc func(context.Context, *ent.OncallUserShiftMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallUserShiftFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallUserShiftMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallUserShiftMutation", m)
}

// The OncallUserShiftAnnotationFunc type is an adapter to allow the use of ordinary
// function as OncallUserShiftAnnotation mutator.
type OncallUserShiftAnnotationFunc func(context.Context, *ent.OncallUserShiftAnnotationMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallUserShiftAnnotationFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallUserShiftAnnotationMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallUserShiftAnnotationMutation", m)
}

// The OncallUserShiftCoverFunc type is an adapter to allow the use of ordinary
// function as OncallUserShiftCover mutator.
type OncallUserShiftCoverFunc func(context.Context, *ent.OncallUserShiftCoverMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallUserShiftCoverFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallUserShiftCoverMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallUserShiftCoverMutation", m)
}

// The OncallUserShiftHandoverFunc type is an adapter to allow the use of ordinary
// function as OncallUserShiftHandover mutator.
type OncallUserShiftHandoverFunc func(context.Context, *ent.OncallUserShiftHandoverMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f OncallUserShiftHandoverFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.OncallUserShiftHandoverMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.OncallUserShiftHandoverMutation", m)
}

// The ProviderConfigFunc type is an adapter to allow the use of ordinary
// function as ProviderConfig mutator.
type ProviderConfigFunc func(context.Context, *ent.ProviderConfigMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f ProviderConfigFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.ProviderConfigMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.ProviderConfigMutation", m)
}

// The ProviderSyncHistoryFunc type is an adapter to allow the use of ordinary
// function as ProviderSyncHistory mutator.
type ProviderSyncHistoryFunc func(context.Context, *ent.ProviderSyncHistoryMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f ProviderSyncHistoryFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.ProviderSyncHistoryMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.ProviderSyncHistoryMutation", m)
}

// The RetrospectiveFunc type is an adapter to allow the use of ordinary
// function as Retrospective mutator.
type RetrospectiveFunc func(context.Context, *ent.RetrospectiveMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f RetrospectiveFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.RetrospectiveMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.RetrospectiveMutation", m)
}

// The RetrospectiveDiscussionFunc type is an adapter to allow the use of ordinary
// function as RetrospectiveDiscussion mutator.
type RetrospectiveDiscussionFunc func(context.Context, *ent.RetrospectiveDiscussionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f RetrospectiveDiscussionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.RetrospectiveDiscussionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.RetrospectiveDiscussionMutation", m)
}

// The RetrospectiveDiscussionReplyFunc type is an adapter to allow the use of ordinary
// function as RetrospectiveDiscussionReply mutator.
type RetrospectiveDiscussionReplyFunc func(context.Context, *ent.RetrospectiveDiscussionReplyMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f RetrospectiveDiscussionReplyFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.RetrospectiveDiscussionReplyMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.RetrospectiveDiscussionReplyMutation", m)
}

// The RetrospectiveReviewFunc type is an adapter to allow the use of ordinary
// function as RetrospectiveReview mutator.
type RetrospectiveReviewFunc func(context.Context, *ent.RetrospectiveReviewMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f RetrospectiveReviewFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.RetrospectiveReviewMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.RetrospectiveReviewMutation", m)
}

// The SystemComponentFunc type is an adapter to allow the use of ordinary
// function as SystemComponent mutator.
type SystemComponentFunc func(context.Context, *ent.SystemComponentMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SystemComponentFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SystemComponentMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SystemComponentMutation", m)
}

// The SystemComponentControlRelationshipFunc type is an adapter to allow the use of ordinary
// function as SystemComponentControlRelationship mutator.
type SystemComponentControlRelationshipFunc func(context.Context, *ent.SystemComponentControlRelationshipMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SystemComponentControlRelationshipFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SystemComponentControlRelationshipMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SystemComponentControlRelationshipMutation", m)
}

// The SystemComponentFeedbackRelationshipFunc type is an adapter to allow the use of ordinary
// function as SystemComponentFeedbackRelationship mutator.
type SystemComponentFeedbackRelationshipFunc func(context.Context, *ent.SystemComponentFeedbackRelationshipMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SystemComponentFeedbackRelationshipFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SystemComponentFeedbackRelationshipMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SystemComponentFeedbackRelationshipMutation", m)
}

// The TaskFunc type is an adapter to allow the use of ordinary
// function as Task mutator.
type TaskFunc func(context.Context, *ent.TaskMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f TaskFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.TaskMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.TaskMutation", m)
}

// The TeamFunc type is an adapter to allow the use of ordinary
// function as Team mutator.
type TeamFunc func(context.Context, *ent.TeamMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f TeamFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.TeamMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.TeamMutation", m)
}

// The UserFunc type is an adapter to allow the use of ordinary
// function as User mutator.
type UserFunc func(context.Context, *ent.UserMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f UserFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.UserMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.UserMutation", m)
}

// Condition is a hook condition function.
type Condition func(context.Context, ent.Mutation) bool

// And groups conditions with the AND operator.
func And(first, second Condition, rest ...Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		if !first(ctx, m) || !second(ctx, m) {
			return false
		}
		for _, cond := range rest {
			if !cond(ctx, m) {
				return false
			}
		}
		return true
	}
}

// Or groups conditions with the OR operator.
func Or(first, second Condition, rest ...Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		if first(ctx, m) || second(ctx, m) {
			return true
		}
		for _, cond := range rest {
			if cond(ctx, m) {
				return true
			}
		}
		return false
	}
}

// Not negates a given condition.
func Not(cond Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		return !cond(ctx, m)
	}
}

// HasOp is a condition testing mutation operation.
func HasOp(op ent.Op) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		return m.Op().Is(op)
	}
}

// HasAddedFields is a condition validating `.AddedField` on fields.
func HasAddedFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if _, exists := m.AddedField(field); !exists {
			return false
		}
		for _, field := range fields {
			if _, exists := m.AddedField(field); !exists {
				return false
			}
		}
		return true
	}
}

// HasClearedFields is a condition validating `.FieldCleared` on fields.
func HasClearedFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if exists := m.FieldCleared(field); !exists {
			return false
		}
		for _, field := range fields {
			if exists := m.FieldCleared(field); !exists {
				return false
			}
		}
		return true
	}
}

// HasFields is a condition validating `.Field` on fields.
func HasFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if _, exists := m.Field(field); !exists {
			return false
		}
		for _, field := range fields {
			if _, exists := m.Field(field); !exists {
				return false
			}
		}
		return true
	}
}

// If executes the given hook under condition.
//
//	hook.If(ComputeAverage, And(HasFields(...), HasAddedFields(...)))
func If(hk ent.Hook, cond Condition) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if cond(ctx, m) {
				return hk(next).Mutate(ctx, m)
			}
			return next.Mutate(ctx, m)
		})
	}
}

// On executes the given hook only for the given operation.
//
//	hook.On(Log, ent.Delete|ent.Create)
func On(hk ent.Hook, op ent.Op) ent.Hook {
	return If(hk, HasOp(op))
}

// Unless skips the given hook only for the given operation.
//
//	hook.Unless(Log, ent.Update|ent.UpdateOne)
func Unless(hk ent.Hook, op ent.Op) ent.Hook {
	return If(hk, Not(HasOp(op)))
}

// FixedError is a hook returning a fixed error.
func FixedError(err error) ent.Hook {
	return func(ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
			return nil, err
		})
	}
}

// Reject returns a hook that rejects all operations that match op.
//
//	func (T) Hooks() []ent.Hook {
//		return []ent.Hook{
//			Reject(ent.Delete|ent.Update),
//		}
//	}
func Reject(op ent.Op) ent.Hook {
	hk := FixedError(fmt.Errorf("%s operation is not allowed", op))
	return On(hk, op)
}

// Chain acts as a list of hooks and is effectively immutable.
// Once created, it will always hold the same set of hooks in the same order.
type Chain struct {
	hooks []ent.Hook
}

// NewChain creates a new chain of hooks.
func NewChain(hooks ...ent.Hook) Chain {
	return Chain{append([]ent.Hook(nil), hooks...)}
}

// Hook chains the list of hooks and returns the final hook.
func (c Chain) Hook() ent.Hook {
	return func(mutator ent.Mutator) ent.Mutator {
		for i := len(c.hooks) - 1; i >= 0; i-- {
			mutator = c.hooks[i](mutator)
		}
		return mutator
	}
}

// Append extends a chain, adding the specified hook
// as the last ones in the mutation flow.
func (c Chain) Append(hooks ...ent.Hook) Chain {
	newHooks := make([]ent.Hook, 0, len(c.hooks)+len(hooks))
	newHooks = append(newHooks, c.hooks...)
	newHooks = append(newHooks, hooks...)
	return Chain{newHooks}
}

// Extend extends a chain, adding the specified chain
// as the last ones in the mutation flow.
func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.hooks...)
}
