package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type Handler struct {
	oapi.SecurityProvider

	*activityHandler
	*aiHandler
	*alertsHandler
	*discussionHandler
	*documentsHandler
	*tasksHandler
	*incidentsHandler
	*incidentMetadataHandler
	*incidentDebriefsHandler
	*investigationsHandler
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
	*situationsHandler
	*teamsHandler
	*usersHandler
	*userSessionsHandler
}

var _ oapi.Handler = (*Handler)(nil)

func NewHandler(
	securityProvider oapi.SecurityProvider,
	db rez.Database,
	agents rez.AiAgentCatalogue,
	agentSessions rez.AiAgentSessionService,
	messages rez.MessageQueue,
	alerts rez.AlertService,
	orgs rez.OrganizationService,
	users rez.UserService,
	documents rez.DocumentsService,
	debriefs rez.DebriefService,
	incidents rez.IncidentService,
	integrations rez.IntegrationService,
	investigations rez.InvestigationService,
	investigationOutputs rez.InvestigationOutputService,
	events rez.EventsService,
	rosters rez.OncallRostersService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	playbooks rez.PlaybookService,
	retros rez.RetrospectiveService,
	tasks rez.TaskService,
	discussions rez.DiscussionService,
	systemAnalysis rez.SystemAnalysisService,
	knowledge rez.KnowledgeGraphQueryService,
	situations rez.SituationService,
) *Handler {
	bh := &baseHandler{users: users}

	return &Handler{
		SecurityProvider: securityProvider,

		alertsHandler:           newAlertsHandler(alerts),
		aiHandler:               newAiHandler(agents, agentSessions, messages),
		userSessionsHandler:     newUserSessionsHandler(bh, orgs),
		documentsHandler:        newDocumentsHandler(bh, documents),
		incidentDebriefsHandler: newIncidentDebriefsHandler(bh, db, debriefs),
		incidentMetadataHandler: newIncidentMetadataHandler(db, incidents),
		tasksHandler:            newTasksHandler(bh, tasks),
		incidentsHandler:        newIncidentsHandler(bh, db, incidents),
		activityHandler:         newActivityHandler(),
		integrationsHandler:     newIntegrationsHandler(integrations),
		investigationsHandler:   newInvestigationsHandler(investigations, investigationOutputs),
		meetingsHandler:         newMeetingsHandler(),
		eventsHandler:           newEventsHandler(bh, events),
		oncallRostersHandler:    newOncallRostersHandler(bh, incidents, rosters, shifts),
		oncallShiftsHandler:     newOncallShiftsHandler(users, incidents, shifts),
		oncallMetricsHandler:    newOncallMetricsHandler(oncallMetrics),
		organizationsHandler:    newOrganizationsHandler(orgs),
		playbooksHandler:        newPlaybooksHandler(playbooks),
		retrospectivesHandler:   newRetrospectivesHandler(users, incidents, retros, documents),
		systemAnalysisHandler:   newSystemAnalysisHandler(systemAnalysis),
		knowledgeGraphHandler:   newKnowledgeGraphHandler(knowledge),
		situationsHandler:       newSituationsHandler(bh, situations),
		discussionHandler:       newDiscussionHandler(discussions),
		teamsHandler:            newTeamsHandler(db),
		usersHandler:            newUsersHandler(users),
	}
}

type baseHandler struct {
	users rez.UserService
}

// mustAuth reads the existing execution identity and panics when the
// protected-handler identity invariant is broken. Only a user actor with a
// present/nonzero user ID and a tenant ID satisfies the invariant; system or
// agent execution is never treated as a logged-in user just because identity
// fields exist.
func (h *baseHandler) mustAuth(ctx context.Context) execution.Auth {
	ec := execution.GetContext(ctx)
	if !ec.IsUser() {
		panic("protected handler requires an authenticated user execution context")
	}
	if ec.Auth.UserID == nil || *ec.Auth.UserID == uuid.Nil {
		panic("protected handler requires a present authenticated user id")
	}
	if ec.Auth.TenantID == nil {
		panic("protected handler requires a tenant id")
	}
	return ec.Auth
}

// mustUserID returns the authenticated user ID without a database call.
func (h *baseHandler) mustUserID(ctx context.Context) uuid.UUID {
	return *h.mustAuth(ctx).UserID
}

// currentUser loads the current user through the execution context's tenant.
// Lookup failures are returned normally for callers to wrap.
func (h *baseHandler) currentUser(ctx context.Context) (*ent.User, error) {
	return h.users.Get(ctx, user.ID(h.mustUserID(ctx)))
}
