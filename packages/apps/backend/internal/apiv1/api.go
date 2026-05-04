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
	*tasksHandler
	*incidentsHandler
	*incidentMetadataHandler
	*incidentDebriefsHandler
	*incidentEventsHandler
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
	*systemComponentsHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(svcs *rez.Services, db *ent.Client) *Handler {
	return &Handler{
		alertsHandler:             newAlertsHandler(svcs.Alerts),
		authSessionsHandler:       newAuthSessionsHandler(svcs.Organizations, svcs.Users),
		documentsHandler:          newDocumentsHandler(svcs.Documents),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db.IncidentDebriefQuestion, svcs.Users, svcs.Debriefs),
		incidentEventsHandler:     newIncidentEventsHandler(db),
		incidentMetadataHandler:   newIncidentMetadataHandler(db, svcs.Incidents),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		tasksHandler:              newTasksHandler(db),
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
		systemAnalysisHandler:     newSystemAnalysisHandler(db, svcs.Components),
		systemComponentsHandler:   newSystemComponentsHandler(db, svcs.Components),
		teamsHandler:              newTeamsHandler(db.User, db.Team, db.TeamMembership),
		usersHandler:              newUsersHandler(svcs.Users),
	}
}
