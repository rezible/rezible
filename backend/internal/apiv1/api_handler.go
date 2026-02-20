package apiv1

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
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
	*organizationsHandler
	*playbooksHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*systemComponentsHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(svcs *rez.Services, db *ent.Client) *Handler {
	return &Handler{
		alertsHandler:             newAlertsHandler(svcs.Alerts),
		authSessionsHandler:       newAuthSessionsHandler(svcs.Auth, svcs.Organizations, svcs.Users),
		documentsHandler:          newDocumentsHandler(svcs.Documents, svcs.Auth),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db.IncidentDebriefQuestion, svcs.Auth, svcs.Users, svcs.Debriefs),
		incidentEventsHandler:     newIncidentEventsHandler(db, svcs.Auth),
		incidentFieldsHandler:     newIncidentFieldsHandler(db),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		incidentRolesHandler:      newincidentRolesHandler(db.IncidentRole),
		incidentSeverityHandler:   newIncidentSeverityHandler(db.IncidentSeverity),
		incidentTagsHandler:       newIncidentTagsHandler(db.IncidentTag),
		tasksHandler:              newTasksHandler(db),
		incidentTypesHandler:      newIncidentTypesHandler(db.IncidentType),
		incidentsHandler:          newIncidentsHandler(db, svcs.Incidents),
		integrationsHandler:       newIntegrationsHandler(svcs.Integrations),
		meetingsHandler:           newMeetingsHandler(),
		eventsHandler:             newEventsHandler(svcs.Auth, svcs.Events),
		eventAnnotationsHandler:   newEventAnnotationsHandler(svcs.Auth, svcs.EventAnnotations),
		oncallRostersHandler:      newOncallRostersHandler(svcs.Auth, svcs.Users, svcs.Incidents, svcs.OncallRosters, svcs.OncallShifts),
		oncallShiftsHandler:       newOncallShiftsHandler(svcs.Auth, svcs.Users, svcs.Incidents, svcs.OncallShifts),
		oncallMetricsHandler:      newOncallMetricsHandler(svcs.OncallMetrics),
		organizationsHandler:      newOrganizationsHandler(svcs.Organizations),
		playbooksHandler:          newPlaybooksHandler(svcs.Playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(svcs.Auth, svcs.Users, svcs.Incidents, svcs.Retros, svcs.Documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(db, svcs.Components),
		systemComponentsHandler:   newSystemComponentsHandler(db, svcs.Components),
		teamsHandler:              newTeamsHandler(svcs.Auth, db.User, db.Team, db.TeamMembership),
		usersHandler:              newUsersHandler(svcs.Users),
	}
}
