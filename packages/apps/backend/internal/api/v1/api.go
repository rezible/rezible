package apiv1

import (
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type Handler struct {
	*alertsHandler
	*userSessionsHandler
	*documentsHandler
	*tasksHandler
	*incidentsHandler
	*incidentMetadataHandler
	*incidentDebriefsHandler
	*incidentMilestonesHandler
	*aiHandler
	*integrationsHandler
	*meetingsHandler
	*eventsHandler
	*oncallMetricsHandler
	*oncallRostersHandler
	*oncallShiftsHandler
	*organizationsHandler
	*playbooksHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*knowledgeGraphHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(
	db rez.Database,
	ai rez.AiService,
	agents rez.AgentSessionService,
	alerts rez.AlertService,
	orgs rez.OrganizationService,
	users rez.UserService,
	documents rez.DocumentsService,
	debriefs rez.DebriefService,
	incidents rez.IncidentService,
	integrations rez.IntegrationService,
	events rez.EventsService,
	rosters rez.OncallRostersService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	playbooks rez.PlaybookService,
	retros rez.RetrospectiveService,
	systemAnalysis rez.SystemAnalysisService,
	knowledge rez.KnowledgeGraphService,
) (*Handler, error) {
	h := &Handler{
		alertsHandler:             newAlertsHandler(alerts),
		aiHandler:                 newAiHandler(ai, agents),
		userSessionsHandler:       newUserSessionsHandler(orgs, users),
		documentsHandler:          newDocumentsHandler(documents, users),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db, users, debriefs),
		incidentMetadataHandler:   newIncidentMetadataHandler(db, incidents),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		tasksHandler:              newTasksHandler(db),
		incidentsHandler:          newIncidentsHandler(incidents),
		integrationsHandler:       newIntegrationsHandler(integrations),
		meetingsHandler:           newMeetingsHandler(),
		eventsHandler:             newEventsHandler(events),
		oncallRostersHandler:      newOncallRostersHandler(users, incidents, rosters, shifts),
		oncallShiftsHandler:       newOncallShiftsHandler(users, incidents, shifts),
		oncallMetricsHandler:      newOncallMetricsHandler(oncallMetrics),
		organizationsHandler:      newOrganizationsHandler(orgs),
		playbooksHandler:          newPlaybooksHandler(playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(users, incidents, retros, documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(systemAnalysis),
		knowledgeGraphHandler:     newKnowledgeGraphHandler(knowledge),
		teamsHandler:              newTeamsHandler(db),
		usersHandler:              newUsersHandler(users),
	}

	return h, nil
}
