package rez

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/texm/prosemirror-go"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
)

var (
	BackendUrl  = "http://localhost:8888"
	FrontendUrl = "http://localhost:5173"
	DebugMode   = true

	ErrNoAuthSession           = errors.New("no auth session")
	ErrAuthSessionExpired      = errors.New("auth session expired")
	ErrAuthSessionUserMissing  = errors.New("missing auth session user")
	ErrAuthSessionInvalidScope = errors.New("invalid session token scope")
	ErrUnauthorized            = errors.New("unauthorized")
)

type (
	ListParams = ent.ListParams
)

type (
	ProviderLoader interface {
		GetIncidentDataProvider(context.Context) (IncidentDataProvider, error)
		GetOncallDataProvider(context.Context) (OncallDataProvider, error)
		GetSystemComponentsDataProvider(context.Context) (SystemComponentsDataProvider, error)
		GetTeamDataProvider(context.Context) (TeamDataProvider, error)
		GetUserDataProvider(context.Context) (UserDataProvider, error)
		GetTicketDataProvider(context.Context) (TicketDataProvider, error)
		GetAlertDataProvider(context.Context) (AlertDataProvider, error)
		GetPlaybookDataProvider(context.Context) (PlaybookDataProvider, error)
	}

	DataProviderResourceUpdatedCallback = func(providerID string, updatedAt time.Time)

	ProviderSyncService interface {
		MakeSyncProviderDataPeriodicJob() jobs.PeriodicJob
		SyncProviderData(context.Context, jobs.SyncProviderData) error
	}
)

type (
	JobsService interface {
		Start(context.Context) error
		Stop(context.Context) error

		Insert(ctx context.Context, params jobs.InsertJobParams) error
		InsertTx(ctx context.Context, tx *ent.Tx, params jobs.InsertJobParams) error
		InsertMany(ctx context.Context, params []jobs.InsertJobParams) error
	}
)

type (
	UserDataProvider interface {
		UserDataMapping() *ent.User
		PullUsers(ctx context.Context) iter.Seq2[*ent.User, error]
	}
	ListUsersParams = struct {
		ListParams
		TeamID uuid.UUID
	}

	UserService interface {
		CreateUserContext(ctx context.Context, userId uuid.UUID) (context.Context, error)
		GetUserContext(ctx context.Context) *ent.User

		FindOrCreateAuthProviderUser(ctx context.Context, user *ent.User, tenantId string) (*ent.User, error)

		ListUsers(context.Context, ListUsersParams) ([]*ent.User, error)

		GetById(context.Context, uuid.UUID) (*ent.User, error)
		GetByEmail(context.Context, string) (*ent.User, error)
		GetByChatId(context.Context, string) (*ent.User, error)

		LookupProviderUser(ctx context.Context, provUser *ent.User) (*ent.User, error)
	}
)

type (
	AuthProviderSession struct {
		User        *ent.User
		TenantId    string
		ExpiresAt   time.Time
		RedirectUrl string
	}

	AuthSessionProvider interface {
		Name() string
		GetUserMapping() *ent.User
		StartAuthFlow(w http.ResponseWriter, r *http.Request)
		HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(AuthProviderSession)) bool
		ClearSession(w http.ResponseWriter, r *http.Request)
	}

	AuthSession struct {
		UserId    uuid.UUID
		ExpiresAt time.Time
	}

	AuthService interface {
		Providers() []AuthSessionProvider

		UserAuthHandler() http.Handler
		MCPServerMiddleware() func(http.Handler) http.Handler

		CreateVerifiedApiAuthContext(ctx context.Context, token string, requiredScopes []string) (context.Context, error)

		IssueAuthSessionToken(sess *AuthSession, scope map[string]string) (string, error)
		VerifyAuthSessionToken(token string, scope map[string]string) (*AuthSession, error)

		SetAuthSession(context.Context, *AuthSession) context.Context
		GetAuthSession(context.Context) (*AuthSession, error)
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
		ListParams
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
	ContentNode        = prosemirror.Node
	DocumentSchemaSpec = prosemirror.SchemaSpec

	OncallShiftHandoverSection struct {
		Header  string            `json:"header"`
		Kind    string            `json:"kind"`
		Content *prosemirror.Node `json:"jsonContent,omitempty"`
	}

	DocumentsService interface {
		GetServerWebsocketAddress() string
		Handler() http.Handler
		CreateEditorSessionToken(sess *AuthSession, docId uuid.UUID) (string, error)
		CreateOncallShiftHandoverMessage(sections []OncallShiftHandoverSection, annotations []*ent.OncallAnnotation, roster *ent.OncallRoster, endingShift *ent.OncallShift, startingShift *ent.OncallShift) (*ContentNode, error)
	}
)

type (
	SendOncallHandoverParams struct {
		Content           []OncallShiftHandoverSection
		EndingShift       *ent.OncallShift
		StartingShift     *ent.OncallShift
		PinnedAnnotations []*ent.OncallAnnotation
	}

	ChatService interface {
		GetWebhooksHandler() http.Handler

		SetOncallEventsService(OncallEventsService)

		SendMessage(ctx context.Context, id string, msg *ContentNode) error
		SendReply(ctx context.Context, channelId string, threadId string, text string) error
		SendTextMessage(ctx context.Context, id string, text string) error

		// TODO: this should just be converted to *ContentNode by DocumentService
		SendOncallHandover(ctx context.Context, params SendOncallHandoverParams) error
		SendOncallHandoverReminder(context.Context, *ent.OncallShift) error
	}
)

type (
	LanguageModelService interface {
		GenerateDebriefResponse(context.Context, *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error)
	}

	AiAgentService interface {
	}
)

type (
	TicketDataProvider interface {
		PullTickets(context.Context) iter.Seq2[*ent.Ticket, error]
	}

	ListTicketsParams struct {
		ListParams
	}

	TicketService interface {
		ListTickets(context.Context, ListTicketsParams) ([]*ent.Ticket, int, error)
	}
)

type (
	AlertDataProvider interface {
		PullAlerts(context.Context) iter.Seq2[*ent.Alert, error]
		PullAlertEventsBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallEvent, error]
	}

	ListAlertsParams struct {
		ListParams
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
	}
)

type (
	PlaybookDataProvider interface {
		PullPlaybooks(context.Context) iter.Seq2[*ent.Playbook, error]
	}

	ListPlaybooksParams struct {
		ListParams
	}

	PlaybookService interface {
		ListPlaybooks(context.Context, ListPlaybooksParams) ([]*ent.Playbook, int, error)
		GetPlaybook(context.Context, uuid.UUID) (*ent.Playbook, error)
		UpdatePlaybook(context.Context, *ent.Playbook) (*ent.Playbook, error)
	}
)

type (
	IncidentDataProvider interface {
		IncidentDataMapping() *ent.Incident
		IncidentRoleDataMapping() *ent.IncidentRole

		SetOnIncidentUpdatedCallback(DataProviderResourceUpdatedCallback)
		PullIncidents(context.Context) iter.Seq2[*ent.Incident, error]
		GetIncidentByID(context.Context, string) (*ent.Incident, error)
		GetRoles(context.Context) ([]*ent.IncidentRole, error)
	}

	ListIncidentsParams struct {
		ListParams
		UserId       uuid.UUID
		OpenedAfter  time.Time
		OpenedBefore time.Time
	}

	IncidentService interface {
		GetByID(context.Context, uuid.UUID) (*ent.Incident, error)
		GetIdForSlug(context.Context, string) (uuid.UUID, error)
		GetBySlug(context.Context, string) (*ent.Incident, error)
		ListIncidents(context.Context, ListIncidentsParams) (*ent.ListResult[*ent.Incident], error)
	}
)

type (
	DebriefService interface {
		HandleSendDebriefRequests(context.Context, jobs.SendIncidentDebriefRequests) error
		HandleGenerateDebriefResponse(context.Context, jobs.GenerateIncidentDebriefResponse) error
		HandleGenerateSuggestions(context.Context, jobs.GenerateIncidentDebriefSuggestions) error

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
		ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	ListRetrospectiveReviewsParams struct {
		ListParams
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
	ExpandAnnotationsParams struct {
		WithCreator       bool
		WithRoster        bool
		WithAlertFeedback bool
		WithEvent         bool
	}

	ListOncallEventsParams struct {
		ListParams
		From               time.Time
		To                 time.Time
		RosterID           uuid.UUID
		AnnotationRosterID uuid.UUID
		AlertID            uuid.UUID
		WithAnnotations    bool
	}

	ListOncallAnnotationsParams struct {
		ListParams
		From     time.Time
		To       time.Time
		RosterID uuid.UUID
		Shift    *ent.OncallShift
		Expand   ExpandAnnotationsParams
	}

	LookupOncallProviderEventFn func(ctx context.Context, id string) (*ent.OncallEvent, error)

	OncallEventsService interface {
		GetEvent(ctx context.Context, id uuid.UUID) (*ent.OncallEvent, error)
		ListEvents(ctx context.Context, params ListOncallEventsParams) (*ent.ListResult[*ent.OncallEvent], error)
		GetProviderEvent(ctx context.Context, providerId string) (*ent.OncallEvent, error)

		ListAnnotations(ctx context.Context, params ListOncallAnnotationsParams) (*ent.ListResult[*ent.OncallAnnotation], error)

		GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallAnnotation, error)
		UpdateAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error)
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
		ListParams
		UserID uuid.UUID
	}

	ListOncallSchedulesParams = struct {
		ListParams
		UserID uuid.UUID
	}

	ListOncallShiftsParams struct {
		ListParams
		UserID uuid.UUID
		Anchor time.Time
		Window time.Duration
	}

	OncallService interface {
		MakeScanShiftsPeriodicJob() jobs.PeriodicJob
		HandlePeriodicScanShifts(context.Context, jobs.ScanOncallShifts) error
		HandleEnsureShiftHandoverSent(context.Context, jobs.EnsureShiftHandoverSent) error
		HandleEnsureShiftHandoverReminderSent(context.Context, jobs.EnsureShiftHandoverReminderSent) error
		HandleGenerateShiftMetrics(context.Context, jobs.GenerateShiftMetrics) error

		ListRosters(context.Context, ListOncallRostersParams) (*ent.ListResult[*ent.OncallRoster], error)
		GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error)
		GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error)
		GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error)

		ListSchedules(ctx context.Context, params ListOncallSchedulesParams) (*ent.ListResult[*ent.OncallSchedule], error)

		ListShifts(ctx context.Context, params ListOncallShiftsParams) (*ent.ListResult[*ent.OncallShift], error)
		GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallShift, error)
		GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallShift, *ent.OncallShift, error)
		GetShiftMetrics(ctx context.Context, id uuid.UUID) (*ent.OncallShiftMetrics, error)
		GetComparisonShiftMetrics(ctx context.Context, from, to time.Time) (*ent.OncallShiftMetrics, error)

		GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
		GetHandoverForShift(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error)
		UpdateShiftHandover(ctx context.Context, handover *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error)
		SendShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error)
	}
)
