package rez

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"github.com/texm/prosemirror-go"
)

var (
	ErrNoAuthSession            = errors.New("no auth session")
	ErrAuthSessionExpired       = errors.New("auth session expired")
	ErrAuthSessionInvalidScope  = errors.New("invalid session token scope")
	ErrInvalidUser              = errors.New("user does not exist")
	ErrInvalidTenant            = errors.New("tenant does not exist")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrNoConfiguredIntegrations = errors.New("no configured integrations")
)

type ConfigLoader interface {
	GetString(key string) string
	GetStringOr(key string, orDefault string) string
	GetStrings(key string) []string
	GetBool(key string) bool
	GetBoolOr(key string, orDefault bool) bool
	GetDuration(key string) time.Duration
	GetDurationOr(key string, orDefault time.Duration) time.Duration

	SingleTenantMode() bool
	DebugMode() bool
	DataSyncMode() bool

	DatabaseUrl() string
	DocumentsServerAddress() string
	AppUrl() string

	ApiRouteBase() string
	AuthRouteBase() string

	AllowTenantCreation() bool
	AllowUserCreation() bool
}

var Config ConfigLoader

type EventListener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Database interface {
	Client() *ent.Client
	Close() error
}

type Services struct {
	Jobs             JobsService
	Messages         MessageService
	Auth             AuthService
	Organizations    OrganizationService
	Integrations     IntegrationsService
	Users            UserService
	Teams            TeamService
	Incidents        IncidentService
	Debriefs         DebriefService
	OncallRosters    OncallRostersService
	OncallShifts     OncallShiftsService
	OncallMetrics    OncallMetricsService
	Events           EventsService
	EventAnnotations EventAnnotationsService
	Documents        DocumentsService
	Retros           RetrospectiveService
	Components       SystemComponentsService
	Alerts           AlertService
	Playbooks        PlaybookService
}

type (
	SetupPackageFunc = func(context.Context, *Services) (IntegrationPackage, error)

	IntegrationPackage interface {
		Name() string
		IsAvailable() (bool, error)
		SupportedDataKinds() []string
		OAuthConfigRequired() bool

		GetConfiguredIntegration(*ent.Integration) ConfiguredIntegration
	}

	ConfiguredIntegration interface {
		Name() string
		RawConfig() json.RawMessage
		UserPreferences() map[string]any
		EnabledDataKinds() []string
		GetSanitizedConfig() (json.RawMessage, error)
	}

	ListIntegrationsParams struct {
		Names        []string
		ConfigValues map[string]any
		Filter       func(*ent.IntegrationQuery)
	}

	IntegrationsService interface {
		LookupByConfigValues(ctx context.Context, name string, configValues map[string]any) (*ent.Integration, error)

		ListConfigured(ctx context.Context, params ListIntegrationsParams) ([]ConfiguredIntegration, error)
		Get(ctx context.Context, name string) (*ent.Integration, error)
		GetConfigured(ctx context.Context, name string) (ConfiguredIntegration, error)
		SetIntegration(ctx context.Context, name string, setFn func(*ent.IntegrationMutation)) (ConfiguredIntegration, error)
		DeleteConfigured(ctx context.Context, name string) error

		StartOAuth2Flow(ctx context.Context, name string) (string, error)
		CompleteOAuth2Flow(ctx context.Context, name, state, code string) (ConfiguredIntegration, error)

		GetChatIntegration(ctx context.Context) (ChatService, error)
		GetVideoConferenceIntegration(ctx context.Context) (VideoConferenceIntegration, error)
	}
)

type (
	MessageService interface {
		AddCommandHandlers(handlers ...cqrs.CommandHandler) error
		SendCommand(ctx context.Context, cmd any) error

		AddEventHandlers(handlers ...cqrs.EventHandler) error
		PublishEvent(ctx context.Context, event any) error
	}
)

func NewCommandHandler[T any](name string, handleFn func(context.Context, *T) error) cqrs.CommandHandler {
	return cqrs.NewCommandHandler[T](name, handleFn)
}

func NewEventHandler[T any](name string, handleFn func(context.Context, *T) error) cqrs.EventHandler {
	return cqrs.NewEventHandler[T](name, handleFn)
}

type (
	JobsService interface {
		Start(context.Context) error
		Stop(context.Context) error

		Insert(context.Context, jobs.JobArgs, *jobs.InsertOpts) error
		InsertTx(context.Context, *ent.Tx, jobs.JobArgs, *jobs.InsertOpts) error
		InsertMany(context.Context, []jobs.InsertManyParams) error
	}
)

type (
	OrganizationService interface {
		GetById(context.Context, uuid.UUID) (*ent.Organization, error)
		GetCurrent(context.Context) (*ent.Organization, error)
		CompleteSetup(context.Context, *ent.Organization) error
		FindOrCreateFromProvider(context.Context, ent.Organization) (*ent.Organization, error)
	}
)

type (
	UserDataProvider interface {
		UserDataMapping() *ent.User
		PullUsers(ctx context.Context) iter.Seq2[*ent.User, error]
	}
	ListUsersParams = struct {
		ent.ListParams
		TeamID uuid.UUID
	}

	UserService interface {
		CreateUserContext(ctx context.Context, userId uuid.UUID) (context.Context, error)
		GetUserContext(ctx context.Context) *ent.User

		FindOrCreateAuthProviderUser(context.Context, ent.User) (*ent.User, error)

		ListUsers(context.Context, ListUsersParams) ([]*ent.User, error)

		GetById(context.Context, uuid.UUID) (*ent.User, error)
		GetByEmail(context.Context, string) (*ent.User, error)
		GetByChatId(context.Context, string) (*ent.User, error)
	}
)

type (
	AuthProviderSession struct {
		Organization ent.Organization
		User         ent.User
		ExpiresAt    time.Time
		RedirectUrl  string
	}

	AuthSessionProvider interface {
		Id() string
		DisplayName() string
		UserMapping() *ent.User
		HandleStartAuthFlow(w http.ResponseWriter, r *http.Request)
		HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(AuthProviderSession)) bool
		SessionExists(r *http.Request) bool
		ClearSession(w http.ResponseWriter, r *http.Request) error
	}

	AuthSessionScopes map[string][]string
	AuthSession       struct {
		UserId    uuid.UUID
		ExpiresAt time.Time
		Scopes    AuthSessionScopes
	}

	AuthService interface {
		Providers() []AuthSessionProvider
		GetProviderStartFlowPath(prov AuthSessionProvider) string

		AuthRouteHandler() http.Handler
		MCPServerMiddleware() func(http.Handler) http.Handler

		IssueAuthSessionToken(sess *AuthSession) (string, error)
		VerifyAuthSessionToken(token string, scopes AuthSessionScopes) (*AuthSession, error)

		CreateAuthContext(context.Context, *AuthSession) (context.Context, error)
		GetAuthSession(context.Context) (*AuthSession, error)
	}
)

type (
	IntegrationWithChatService interface {
		ChatService() ChatService
	}

	ChatService interface {
		SendMessage(ctx context.Context, id string, msg *ContentNode) (string, error)
		SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error)
		SendTextMessage(ctx context.Context, id string, text string) (string, error)
	}
)

type (
	IntegrationWithVideoConference interface {
		VideoConferenceIntegration(ctx context.Context) (VideoConferenceIntegration, error)
	}

	VideoConferenceIntegration interface {
		CreateVideoConference(ctx context.Context) (string, error)
	}
)

type (
	SystemComponentsDataProvider interface {
		SystemComponentDataMapping() *ent.SystemComponent
		PullSystemComponents(context.Context) iter.Seq2[*ent.SystemComponent, error]
	}

	ComponentTraitReference struct {
		Id          uuid.UUID
		Description string
	}

	CreateSystemAnalysisRelationshipParams struct {
		AnalysisId      uuid.UUID
		SourceId        uuid.UUID
		TargetId        uuid.UUID
		Description     string
		FeedbackSignals []ComponentTraitReference
		ControlActions  []ComponentTraitReference
	}

	ListSystemComponentsParams struct {
		ent.ListParams
	}

	SystemComponentsService interface {
		Create(context.Context, ent.SystemComponent) (*ent.SystemComponent, error)

		ListSystemComponents(context.Context, ListSystemComponentsParams) (*ent.ListResult[*ent.SystemComponent], error)

		GetRelationship(context.Context, uuid.UUID, uuid.UUID) (*ent.SystemComponentRelationship, error)
		CreateRelationship(context.Context, ent.SystemComponentRelationship) (*ent.SystemComponentRelationship, error)

		GetSystemAnalysis(context.Context, uuid.UUID) (*ent.SystemAnalysis, error)
		CreateSystemAnalysisRelationship(context.Context, CreateSystemAnalysisRelationshipParams) (*ent.SystemAnalysisRelationship, error)
	}
)

type (
	TeamDataProvider interface {
		TeamDataMapping() *ent.Team
		PullTeams(context.Context) iter.Seq2[*ent.Team, error]
	}

	TeamService interface {
		GetById(context.Context, uuid.UUID) (*ent.Team, error)
	}
)

type (
	ContentNode = prosemirror.Node

	DocumentsService interface {
		GetUserDocumentAccess(ctx context.Context, userId uuid.UUID, documentId uuid.UUID) (bool, error)

		GetDocument(context.Context, uuid.UUID) (*ent.Document, error)
		SetDocument(context.Context, uuid.UUID, func(*ent.DocumentMutation)) (*ent.Document, error)
	}
)

type (
	AiAgentService interface {
		GenerateDebriefResponse(context.Context, *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error)
	}
)

type (
	TicketDataProvider interface {
		PullTickets(context.Context) iter.Seq2[*ent.Ticket, error]
	}
)

type (
	AlertDataProvider interface {
		PullAlerts(context.Context) iter.Seq2[*ent.Alert, error]
		PullAlertInstancesBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.AlertInstance, error]
	}

	ListAlertsParams struct {
		ent.ListParams
	}

	GetAlertMetricsParams struct {
		AlertId  uuid.UUID
		RosterId uuid.UUID
		From     time.Time
		To       time.Time
	}

	AlertService interface {
		ListAlerts(context.Context, ListAlertsParams) ([]*ent.Alert, int, error)
		GetAlert(context.Context, uuid.UUID) (*ent.Alert, error)
		GetAlertMetrics(context.Context, GetAlertMetricsParams) (*ent.AlertMetrics, error)
		GetActiveAlertsForComponents(context.Context, []uuid.UUID) ([]*ent.Alert, error)
	}
)

type (
	PlaybookDataProvider interface {
		PullPlaybooks(context.Context) iter.Seq2[*ent.Playbook, error]
	}

	ListPlaybooksParams struct {
		ent.ListParams
	}

	PlaybookService interface {
		ListPlaybooks(context.Context, ListPlaybooksParams) ([]*ent.Playbook, int, error)
		GetPlaybook(context.Context, uuid.UUID) (*ent.Playbook, error)
		SetPlaybook(context.Context, *ent.Playbook) (*ent.Playbook, error)
	}
)

type (
	IncidentDataProvider interface {
		IncidentDataMapping() *ent.Incident
		IncidentRoleDataMapping() *ent.IncidentRole

		PullIncidents(context.Context) iter.Seq2[*ent.Incident, error]
		GetIncidentByID(context.Context, string) (*ent.Incident, error)
		ListIncidentRoles(context.Context) ([]*ent.IncidentRole, error)
	}

	IncidentMetadata struct {
		Roles      ent.IncidentRoles
		Types      ent.IncidentTypes
		Fields     ent.IncidentFields
		Severities ent.IncidentSeverities
	}

	ListIncidentsParams struct {
		ent.ListParams
		UserId       uuid.UUID
		OpenedAfter  time.Time
		OpenedBefore time.Time
	}

	IncidentService interface {
		ListIncidents(context.Context, ListIncidentsParams) (*ent.ListResult[*ent.Incident], error)
		Get(context.Context, uuid.UUID) (*ent.Incident, error)
		Set(context.Context, uuid.UUID, func(*ent.IncidentMutation) []ent.Mutation) (*ent.Incident, error)
		GetBySlug(context.Context, string) (*ent.Incident, error)
		GetByChatChannelID(context.Context, string) (*ent.Incident, error)

		ListIncidentRoles(context.Context) ([]*ent.IncidentRole, error)

		GetIncidentMilestone(context.Context, uuid.UUID) (*ent.IncidentMilestone, error)
		SetIncidentMilestone(context.Context, uuid.UUID, func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error)

		ListIncidentSeverities(context.Context) ([]*ent.IncidentSeverity, error)
		GetIncidentSeverity(context.Context, uuid.UUID) (*ent.IncidentSeverity, error)

		ListIncidentTypes(context.Context) ([]*ent.IncidentType, error)

		GetIncidentMetadata(context.Context) (*IncidentMetadata, error)
	}

	EventOnIncidentUpdated struct {
		Created    bool
		IncidentId uuid.UUID
	}

	EventOnIncidentMilestoneUpdated struct {
		Created     bool
		IncidentId  uuid.UUID
		MilestoneId uuid.UUID
	}
)

type (
	DebriefService interface {
		CreateDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error)
		GetDebrief(ctx context.Context, id uuid.UUID) (*ent.IncidentDebrief, error)
		GetUserDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error)
		AddDebriefMessage(ctx context.Context, debriefId uuid.UUID, text string) (*ent.IncidentDebriefMessage, error)

		StartDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error)
		CompleteDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error)
	}
)

type (
	ListRetrospectiveCommentsParams struct {
		ent.ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	ListRetrospectiveReviewsParams struct {
		ent.ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	RetrospectiveService interface {
		Create(context.Context, ent.Retrospective) (*ent.Retrospective, error)
		GetById(context.Context, uuid.UUID) (*ent.Retrospective, error)
		GetForIncident(context.Context, *ent.Incident) (*ent.Retrospective, error)

		//ListReviews(context.Context, ListRetrospectiveReviewsParams) ([]*ent.RetrospectiveReview, error)
		//GetReview(context.Context, uuid.UUID) (*ent.RetrospectiveReview, error)
		//SetReview(context.Context, *ent.RetrospectiveReview) (*ent.RetrospectiveReview, error)

		ListComments(context.Context, ListRetrospectiveCommentsParams) ([]*ent.RetrospectiveComment, error)
		GetComment(context.Context, uuid.UUID) (*ent.RetrospectiveComment, error)
		SetComment(context.Context, *ent.RetrospectiveComment) (*ent.RetrospectiveComment, error)
	}
)

type (
	ListEventsParams struct {
		ent.ListParams
		From            time.Time
		To              time.Time
		WithAnnotations bool
	}

	EventsService interface {
		GetEvent(ctx context.Context, id uuid.UUID) (*ent.Event, error)
		ListEvents(ctx context.Context, params ListEventsParams) (*ent.ListResult[*ent.Event], error)
	}

	ExpandAnnotationsParams struct {
		WithCreator       bool
		WithRoster        bool
		WithAlertFeedback bool
		WithEvent         bool
	}

	ListAnnotationsParams struct {
		ent.ListParams
		From     time.Time
		To       time.Time
		UserIds  []uuid.UUID
		EventIds []uuid.UUID
		Expand   ExpandAnnotationsParams
	}

	EventAnnotationsService interface {
		ListAnnotations(ctx context.Context, params ListAnnotationsParams) (*ent.ListResult[*ent.EventAnnotation], error)

		LookupByUserEvent(ctx context.Context, userId uuid.UUID, event *ent.Event) (*ent.EventAnnotation, error)

		GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error)
		SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error)
		DeleteAnnotation(ctx context.Context, id uuid.UUID) error
	}
)

type (
	OncallDataProvider interface {
		RosterDataMapping() *ent.OncallRoster
		UserShiftDataMapping() *ent.OncallShift

		PullRosters(context.Context) iter.Seq2[*ent.OncallRoster, error]
		PullShiftsForRoster(ctx context.Context, rosterId string, from, to time.Time) iter.Seq2[*ent.OncallShift, error]
		FetchOncallersForRoster(ctx context.Context, rosterId string) ([]*ent.User, error)
	}

	ListOncallRostersParams = struct {
		ent.ListParams
		UserID uuid.UUID
	}

	ListOncallSchedulesParams = struct {
		ent.ListParams
		UserID uuid.UUID
	}

	OncallRostersService interface {
		ListRosters(context.Context, ListOncallRostersParams) (*ent.ListResult[*ent.OncallRoster], error)
		GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error)
		GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error)
		GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error)

		ListSchedules(ctx context.Context, params ListOncallSchedulesParams) (*ent.ListResult[*ent.OncallSchedule], error)

		GetCurrentOncallForComponent(context.Context, uuid.UUID) ([]*ent.User, error)
	}

	ListOncallShiftsParams struct {
		ent.ListParams
		UserID uuid.UUID
		Anchor time.Time
		Window time.Duration
	}

	OncallShiftHandoverSection struct {
		Header  string            `json:"header"`
		Kind    string            `json:"kind"`
		Content *prosemirror.Node `json:"jsonContent,omitempty"`
	}

	OncallShiftsService interface {
		ListShifts(ctx context.Context, params ListOncallShiftsParams) (*ent.ListResult[*ent.OncallShift], error)
		GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallShift, error)
		GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallShift, *ent.OncallShift, error)

		GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
		GetHandoverForShift(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error)
		UpdateShiftHandover(ctx context.Context, handover *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error)
		SendShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
	}

	OncallMetricsService interface {
		GetShiftMetrics(ctx context.Context, id uuid.UUID) (*ent.OncallShiftMetrics, error)
		GetComparisonShiftMetrics(ctx context.Context, from, to time.Time) (*ent.OncallShiftMetrics, error)
	}
)
