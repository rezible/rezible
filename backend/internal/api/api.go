package api

import (
	"context"
	"regexp"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	*middlewareHandler

	*alertsHandler
	*authSessionsHandler
	*documentsHandler
	*environmentsHandler
	*functionalitiesHandler
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
	*oncallEventsHandler
	*oncallMetricsHandler
	*oncallRostersHandler
	*oncallShiftsHandler
	*playbooksHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*systemComponentsHandler
	*subscriptionsHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(
	db *ent.Client,
	auth rez.AuthSessionService,
	users rez.UserService,
	incidents rez.IncidentService,
	debriefs rez.DebriefService,
	oncall rez.OncallService,
	oncallEvents rez.OncallEventsService,
	documents rez.DocumentsService,
	retros rez.RetrospectiveService,
	components rez.SystemComponentsService,
	alerts rez.AlertService,
	playbooks rez.PlaybookService,
) *Handler {
	return &Handler{
		middlewareHandler: newMiddlewareHandler(auth),

		alertsHandler:             newAlertsHandler(alerts),
		oncallMetricsHandler:      newOncallMetricsHandler(),
		authSessionsHandler:       newAuthSessionsHandler(auth, users),
		documentsHandler:          newDocumentsHandler(documents, auth, users),
		environmentsHandler:       newEnvironmentsHandler(db.Environment),
		functionalitiesHandler:    newFunctionalitiesHandler(),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db.IncidentDebriefQuestion, auth, users, debriefs),
		incidentEventsHandler:     newIncidentEventsHandler(db, auth),
		incidentFieldsHandler:     newIncidentFieldsHandler(db),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db),
		incidentRolesHandler:      newincidentRolesHandler(db.IncidentRole),
		incidentSeverityHandler:   newIncidentSeverityHandler(db.IncidentSeverity),
		incidentTagsHandler:       newIncidentTagsHandler(db.IncidentTag),
		tasksHandler:              newTasksHandler(db),
		incidentTypesHandler:      newIncidentTypesHandler(db.IncidentType),
		incidentsHandler:          newIncidentsHandler(db),
		integrationsHandler:       newIntegrationsHandler(),
		meetingsHandler:           newMeetingsHandler(),
		oncallEventsHandler:       newOncallEventsHandler(auth, users, oncall, incidents, oncallEvents),
		oncallRostersHandler:      newOncallRostersHandler(auth, users, incidents, oncall),
		oncallShiftsHandler:       newOncallShiftsHandler(auth, users, incidents, oncall),
		playbooksHandler:          newPlaybooksHandler(playbooks),
		retrospectivesHandler:     newRetrospectivesHandler(auth, users, incidents, retros, documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(db, components),
		systemComponentsHandler:   newSystemComponentsHandler(db, components),
		subscriptionsHandler:      newSubscriptionsHandler(),
		teamsHandler:              newTeamsHandler(db.Team),
		usersHandler:              newUsersHandler(users),
	}
}

func mustGetAuthSession(ctx context.Context, auth rez.AuthSessionService) *rez.AuthSession {
	sess, sessErr := auth.GetSession(ctx)
	if sessErr != nil {
		panic("mustGetAuthSession: " + sessErr.Error())
	}
	return sess
}

var (
	uniqueErrFieldRe         = regexp.MustCompile("unique constraint \".*_(.*)_key\"")
	enumValidationErrFieldRe = regexp.MustCompile("invalid enum value for")

	commonConstraints = map[string]string{
		"name":  "Name already exists",
		"value": "Value already exists",
	}
)

func detailError(msg string, err error) error {
	if isClientError, clientErr := checkIsClientError(err); isClientError {
		return clientErr
	}
	log.Error().Err(err).Msg(msg)
	return oapi.ErrorInternal(msg, err)
}

func checkIsClientError(err error) (bool, error) {
	if oapi.IsClientError(err) {
		return true, err
	}

	if ent.IsNotFound(err) {
		return true, oapi.ErrorNotFound("Not found")
	}

	if enumValidationErrFieldRe.MatchString(err.Error()) {
		return true, err
	}

	if ent.IsConstraintError(err) {
		match := uniqueErrFieldRe.FindStringSubmatch(err.Error())
		if match == nil || len(match) < 2 {
			return true, oapi.ErrorBadRequest("Constraint failed")
		}

		field := match[1]
		msg, found := commonConstraints[field]
		if found {
			detail := oapi.NewErrorDetail(msg, field, nil)
			return true, oapi.ErrorBadRequest("", detail)
		}
		return true, oapi.ErrorBadRequest("Value is not unique")
	}

	return false, nil
}
