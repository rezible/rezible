// Code generated by ent, DO NOT EDIT.

package ent

import "entgo.io/ent/dialect"

func (c *EnvironmentClient) Debug() *EnvironmentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &EnvironmentClient{config: cfg}
}

func (c *FunctionalityClient) Debug() *FunctionalityClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &FunctionalityClient{config: cfg}
}

func (c *IncidentClient) Debug() *IncidentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentClient{config: cfg}
}

func (c *IncidentDebriefClient) Debug() *IncidentDebriefClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentDebriefClient{config: cfg}
}

func (c *IncidentDebriefMessageClient) Debug() *IncidentDebriefMessageClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentDebriefMessageClient{config: cfg}
}

func (c *IncidentDebriefQuestionClient) Debug() *IncidentDebriefQuestionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentDebriefQuestionClient{config: cfg}
}

func (c *IncidentDebriefSuggestionClient) Debug() *IncidentDebriefSuggestionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentDebriefSuggestionClient{config: cfg}
}

func (c *IncidentEventClient) Debug() *IncidentEventClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentEventClient{config: cfg}
}

func (c *IncidentEventContextClient) Debug() *IncidentEventContextClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentEventContextClient{config: cfg}
}

func (c *IncidentEventContributingFactorClient) Debug() *IncidentEventContributingFactorClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentEventContributingFactorClient{config: cfg}
}

func (c *IncidentEventEvidenceClient) Debug() *IncidentEventEvidenceClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentEventEvidenceClient{config: cfg}
}

func (c *IncidentEventSystemComponentClient) Debug() *IncidentEventSystemComponentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentEventSystemComponentClient{config: cfg}
}

func (c *IncidentFieldClient) Debug() *IncidentFieldClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentFieldClient{config: cfg}
}

func (c *IncidentFieldOptionClient) Debug() *IncidentFieldOptionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentFieldOptionClient{config: cfg}
}

func (c *IncidentLinkClient) Debug() *IncidentLinkClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentLinkClient{config: cfg}
}

func (c *IncidentMilestoneClient) Debug() *IncidentMilestoneClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentMilestoneClient{config: cfg}
}

func (c *IncidentRoleClient) Debug() *IncidentRoleClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentRoleClient{config: cfg}
}

func (c *IncidentRoleAssignmentClient) Debug() *IncidentRoleAssignmentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentRoleAssignmentClient{config: cfg}
}

func (c *IncidentSeverityClient) Debug() *IncidentSeverityClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentSeverityClient{config: cfg}
}

func (c *IncidentTagClient) Debug() *IncidentTagClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentTagClient{config: cfg}
}

func (c *IncidentTeamAssignmentClient) Debug() *IncidentTeamAssignmentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentTeamAssignmentClient{config: cfg}
}

func (c *IncidentTypeClient) Debug() *IncidentTypeClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &IncidentTypeClient{config: cfg}
}

func (c *MeetingScheduleClient) Debug() *MeetingScheduleClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &MeetingScheduleClient{config: cfg}
}

func (c *MeetingSessionClient) Debug() *MeetingSessionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &MeetingSessionClient{config: cfg}
}

func (c *OncallAlertClient) Debug() *OncallAlertClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallAlertClient{config: cfg}
}

func (c *OncallAlertInstanceClient) Debug() *OncallAlertInstanceClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallAlertInstanceClient{config: cfg}
}

func (c *OncallHandoverTemplateClient) Debug() *OncallHandoverTemplateClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallHandoverTemplateClient{config: cfg}
}

func (c *OncallRosterClient) Debug() *OncallRosterClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallRosterClient{config: cfg}
}

func (c *OncallScheduleClient) Debug() *OncallScheduleClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallScheduleClient{config: cfg}
}

func (c *OncallScheduleParticipantClient) Debug() *OncallScheduleParticipantClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallScheduleParticipantClient{config: cfg}
}

func (c *OncallUserShiftClient) Debug() *OncallUserShiftClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallUserShiftClient{config: cfg}
}

func (c *OncallUserShiftAnnotationClient) Debug() *OncallUserShiftAnnotationClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallUserShiftAnnotationClient{config: cfg}
}

func (c *OncallUserShiftCoverClient) Debug() *OncallUserShiftCoverClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallUserShiftCoverClient{config: cfg}
}

func (c *OncallUserShiftHandoverClient) Debug() *OncallUserShiftHandoverClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &OncallUserShiftHandoverClient{config: cfg}
}

func (c *ProviderConfigClient) Debug() *ProviderConfigClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &ProviderConfigClient{config: cfg}
}

func (c *ProviderSyncHistoryClient) Debug() *ProviderSyncHistoryClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &ProviderSyncHistoryClient{config: cfg}
}

func (c *RetrospectiveClient) Debug() *RetrospectiveClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &RetrospectiveClient{config: cfg}
}

func (c *RetrospectiveDiscussionClient) Debug() *RetrospectiveDiscussionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &RetrospectiveDiscussionClient{config: cfg}
}

func (c *RetrospectiveDiscussionReplyClient) Debug() *RetrospectiveDiscussionReplyClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &RetrospectiveDiscussionReplyClient{config: cfg}
}

func (c *RetrospectiveReviewClient) Debug() *RetrospectiveReviewClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &RetrospectiveReviewClient{config: cfg}
}

func (c *SystemAnalysisClient) Debug() *SystemAnalysisClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemAnalysisClient{config: cfg}
}

func (c *SystemAnalysisComponentClient) Debug() *SystemAnalysisComponentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemAnalysisComponentClient{config: cfg}
}

func (c *SystemAnalysisRelationshipClient) Debug() *SystemAnalysisRelationshipClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemAnalysisRelationshipClient{config: cfg}
}

func (c *SystemComponentClient) Debug() *SystemComponentClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentClient{config: cfg}
}

func (c *SystemComponentConstraintClient) Debug() *SystemComponentConstraintClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentConstraintClient{config: cfg}
}

func (c *SystemComponentControlClient) Debug() *SystemComponentControlClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentControlClient{config: cfg}
}

func (c *SystemComponentKindClient) Debug() *SystemComponentKindClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentKindClient{config: cfg}
}

func (c *SystemComponentRelationshipClient) Debug() *SystemComponentRelationshipClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentRelationshipClient{config: cfg}
}

func (c *SystemComponentSignalClient) Debug() *SystemComponentSignalClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemComponentSignalClient{config: cfg}
}

func (c *SystemRelationshipControlActionClient) Debug() *SystemRelationshipControlActionClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemRelationshipControlActionClient{config: cfg}
}

func (c *SystemRelationshipFeedbackSignalClient) Debug() *SystemRelationshipFeedbackSignalClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &SystemRelationshipFeedbackSignalClient{config: cfg}
}

func (c *TaskClient) Debug() *TaskClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &TaskClient{config: cfg}
}

func (c *TeamClient) Debug() *TeamClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &TeamClient{config: cfg}
}

func (c *UserClient) Debug() *UserClient {
	if c.debug {
		return c
	}
	cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
	return &UserClient{config: cfg}
}
