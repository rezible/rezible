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
	*oncallHandler
	*retrospectivesHandler
	*systemAnalysisHandler
	*systemComponentsHandler
	*sessionsHandler
	*subscriptionsHandler
	*teamsHandler
	*usersHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(
	db *ent.Client,
	auth rez.AuthService,
	users rez.UserService,
	incidents rez.IncidentService,
	debriefs rez.DebriefService,
	oncall rez.OncallService,
	alerts rez.AlertsService,
	documents rez.DocumentsService,
	retros rez.RetrospectiveService,
) *Handler {
	return &Handler{
		middlewareHandler: newMiddlewareHandler(auth),

		documentsHandler:          newDocumentsHandler(documents, auth, users),
		environmentsHandler:       newEnvironmentsHandler(db.Environment),
		functionalitiesHandler:    newFunctionalitiesHandler(),
		incidentDebriefsHandler:   newIncidentDebriefsHandler(db.IncidentDebriefQuestion, auth, users, debriefs),
		incidentEventsHandler:     newIncidentEventsHandler(db),
		incidentFieldsHandler:     newIncidentFieldsHandler(db),
		incidentMilestonesHandler: newIncidentMilestonesHandler(db.IncidentMilestone),
		incidentRolesHandler:      newincidentRolesHandler(db.IncidentRole),
		incidentSeverityHandler:   newIncidentSeverityHandler(db.IncidentSeverity),
		incidentTagsHandler:       newIncidentTagsHandler(db.IncidentTag),
		tasksHandler:              newTasksHandler(db),
		incidentTypesHandler:      newIncidentTypesHandler(db.IncidentType),
		incidentsHandler:          newIncidentsHandler(db),
		integrationsHandler:       newIntegrationsHandler(),
		meetingsHandler:           newMeetingsHandler(),
		oncallHandler:             newOncallHandler(auth, users, incidents, oncall, alerts),
		retrospectivesHandler:     newRetrospectivesHandler(auth, users, incidents, retros, documents),
		systemAnalysisHandler:     newSystemAnalysisHandler(),
		systemComponentsHandler:   newSystemComponentsHandler(),
		subscriptionsHandler:      newSubscriptionsHandler(),
		teamsHandler:              newTeamsHandler(db.Team),
		usersHandler:              newUsersHandler(users),
		sessionsHandler:           newSessionsHandler(auth, users),
	}
}

func (h *Handler) MakeAdapter() oapi.Adapter {
	return oapi.MakeDefaultApi(h).Adapter()
}

func mustGetAuthSession(ctx context.Context, auth rez.AuthService) *rez.AuthSession {
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
