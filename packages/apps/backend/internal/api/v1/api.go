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

func NewHandler(
	db rez.Database,
	alerts rez.AlertService,
	orgs rez.OrganizationService,
	users rez.UserService,
	documents rez.DocumentsService,
	debriefs rez.DebriefService,
	incidents rez.IncidentService,
	integrations rez.IntegrationsService,
	provEvents rez.ProviderEventService,
	eventAnnotations rez.EventAnnotationsService,
	rosters rez.OncallRostersService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	playbooks rez.PlaybookService,
	retros rez.RetrospectiveService,
	topology rez.SystemTopologyService,
) *Handler {
	return &Handler{
		alertsHandler:       newAlertsHandler(alerts),
		authSessionsHandler: newAuthSessionsHandler(orgs, users),
		documentsHandler:    newDocumentsHandler(documents),
		incidentDebriefsHandler: newIncidentDebriefsHandler(
			db, users, debriefs),
		incidentTimelineHandler:   newIncidentTimelineHandler(db),
		incidentMetadataHandler:   newIncidentMetadataHandler(db, incidents),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		tasksHandler:              newTasksHandler(db),
		incidentsHandler:          newIncidentsHandler(incidents),
		integrationsHandler:       newIntegrationsHandler(integrations),
		meetingsHandler:           newMeetingsHandler(),
		eventsHandler:             newEventsHandler(provEvents),
		eventAnnotationsHandler:   newEventAnnotationsHandler(eventAnnotations),
		oncallRostersHandler:      newOncallRostersHandler(users, incidents, rosters, shifts),
		oncallShiftsHandler:       newOncallShiftsHandler(users, incidents, shifts),
		oncallMetricsHandler:      newOncallMetricsHandler(oncallMetrics),
		organizationsHandler:      newOrganizationsHandler(orgs),
		playbooksHandler:          newPlaybooksHandler(playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(users, incidents, retros, documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(db),
		systemTopologyHandler:     newSystemTopologyHandler(topology),
		teamsHandler:              newTeamsHandler(db),
		usersHandler:              newUsersHandler(users),
	}
}
