package apiv1

import (
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type Handler struct {
	*alertsHandler
	*authSessionsHandler
	*documentsHandler
	*tasksHandler
	*incidentsHandler
	*incidentMetadataHandler
	*incidentDebriefsHandler
	*incidentTimelineHandler
	*incidentMilestonesHandler
	*integrationsHandler
	*meetingsHandler
	*eventsHandler
	*eventAnnotationsHandler
	*oncallMetricsHandler
	*oncallRostersHandler
	*oncallShiftsHandler
	*organizationsHandler
	*playbooksHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*systemTopologyHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(svcs *rez.Services) *Handler {
	dbc := svcs.Database.Client()

	return &Handler{
		alertsHandler:             newAlertsHandler(svcs.Alerts),
		authSessionsHandler:       newAuthSessionsHandler(svcs.Organizations, svcs.Users),
		documentsHandler:          newDocumentsHandler(svcs.Documents),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(dbc.IncidentDebriefQuestion, svcs.Users, svcs.Debriefs),
		incidentTimelineHandler:   newIncidentTimelineHandler(dbc),
		incidentMetadataHandler:   newIncidentMetadataHandler(dbc, svcs.Incidents),
		incidentMilestonesHandler: newIncidentMilestonesHandler(dbc),
		tasksHandler:              newTasksHandler(dbc),
		incidentsHandler:          newIncidentsHandler(svcs.Incidents),
		integrationsHandler:       newIntegrationsHandler(svcs.Integrations),
		meetingsHandler:           newMeetingsHandler(),
		eventsHandler:             newEventsHandler(svcs.Events),
		eventAnnotationsHandler:   newEventAnnotationsHandler(svcs.EventAnnotations),
		oncallRostersHandler:      newOncallRostersHandler(svcs.Users, svcs.Incidents, svcs.OncallRosters, svcs.OncallShifts),
		oncallShiftsHandler:       newOncallShiftsHandler(svcs.Users, svcs.Incidents, svcs.OncallShifts),
		oncallMetricsHandler:      newOncallMetricsHandler(svcs.OncallMetrics),
		organizationsHandler:      newOrganizationsHandler(svcs.Organizations),
		playbooksHandler:          newPlaybooksHandler(svcs.Playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(svcs.Users, svcs.Incidents, svcs.Retros, svcs.Documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(dbc),
		systemTopologyHandler:     newSystemTopologyHandler(svcs.Topology),
		teamsHandler:              newTeamsHandler(dbc.User, dbc.Team, dbc.TeamMembership),
		usersHandler:              newUsersHandler(svcs.Users),
	}
}
