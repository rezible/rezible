package api

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type Handler struct {
	*alertsHandler
	*authSessionsHandler
	*documentsHandler
	*incidentDebriefsHandler
	*incidentEventsHandler
	*incidentFieldsHandler
	*incidentMilestonesHandler
	*incidentRolesHandler
	*incidentSeverityHandler
	*incidentTagsHandler
	*tasksHandler
	*incidentTypesHandler
	*incidentsHandler
	*integrationsHandler
	*meetingsHandler
	*eventsHandler
	*eventAnnotationsHandler
	*oncallMetricsHandler
	*oncallRostersHandler
	*oncallShiftsHandler
	*playbooksHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*systemComponentsHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(
	db *ent.Client,
	auth rez.AuthService,
	orgs rez.OrganizationService,
	configs rez.ProviderConfigService,
	users rez.UserService,
	incidents rez.IncidentService,
	debriefs rez.DebriefService,
	rosters rez.OncallRostersService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	events rez.EventsService,
	annos rez.EventAnnotationsService,
	documents rez.DocumentsService,
	retros rez.RetrospectiveService,
	components rez.SystemComponentsService,
	alerts rez.AlertService,
	playbooks rez.PlaybookService,
) *Handler {
	return &Handler{
		alertsHandler:             newAlertsHandler(alerts),
		authSessionsHandler:       newAuthSessionsHandler(auth, orgs, users),
		documentsHandler:          newDocumentsHandler(documents, auth, users),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db.IncidentDebriefQuestion, auth, users, debriefs),
		incidentEventsHandler:     newIncidentEventsHandler(db, auth),
		incidentFieldsHandler:     newIncidentFieldsHandler(db),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		incidentRolesHandler:      newincidentRolesHandler(db.IncidentRole),
		incidentSeverityHandler:   newIncidentSeverityHandler(db.IncidentSeverity),
		incidentTagsHandler:       newIncidentTagsHandler(db.IncidentTag),
		tasksHandler:              newTasksHandler(db),
		incidentTypesHandler:      newIncidentTypesHandler(db.IncidentType),
		incidentsHandler:          newIncidentsHandler(db, incidents),
		integrationsHandler:       newIntegrationsHandler(orgs, configs),
		meetingsHandler:           newMeetingsHandler(),
		eventsHandler:             newEventsHandler(auth, events),
		eventAnnotationsHandler:   newEventAnnotationsHandler(auth, annos),
		oncallRostersHandler:      newOncallRostersHandler(auth, users, incidents, rosters, shifts),
		oncallShiftsHandler:       newOncallShiftsHandler(auth, users, incidents, shifts),
		oncallMetricsHandler:      newOncallMetricsHandler(oncallMetrics),
		playbooksHandler:          newPlaybooksHandler(playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(auth, users, incidents, retros, documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(db, components),
		systemComponentsHandler:   newSystemComponentsHandler(db, components),
		teamsHandler:              newTeamsHandler(db.Team),
		usersHandler:              newUsersHandler(users),
	}
}
