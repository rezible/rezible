package rez

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/google/uuid"
	"github.com/texm/prosemirror-go"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
)

var (
	ErrNoAuthSession          = errors.New("no auth session")
	ErrAuthSessionExpired     = errors.New("auth session expired")
	ErrAuthSessionUserMissing = errors.New("missing auth session user")

	ErrUnauthorized = errors.New("unauthorized")

	BackendUrl  = "http://localhost:8888"
	FrontendUrl = "http://localhost:5173"
	DebugMode   = true
)

type (
	ListParams = ent.ListParams
)

type (
	Webhooks = map[string]http.Handler

	Server interface {
		RegisterWebhooks(...Webhooks)
		Start(ctx context.Context) error
		Stop(ctx context.Context) error
	}
)

type (
	ProviderLoader interface {
		WebhookHandler() http.Handler

		GetLanguageModelProvider(context.Context) (LanguageModelProvider, error)
		GetChatProvider(context.Context) (ChatProvider, error)
		GetAuthSessionProvider(context.Context) (AuthSessionProvider, error)
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
	AuthSessionCreatedFn = func(*ent.User, time.Time, string)

	AuthSessionProvider interface {
		Name() string
		GetUserMapping() *ent.User
		StartAuthFlow(w http.ResponseWriter, r *http.Request)
		HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreatedFn AuthSessionCreatedFn) (handled bool)
		ClearSession(w http.ResponseWriter, r *http.Request)
	}

	UserAuthSession struct {
		UserId    uuid.UUID
		ExpiresAt time.Time
	}

	AuthSessionService interface {
		ProviderName(context.Context) (string, error)

		AuthHandler() http.Handler
		FrontendMiddleware() func(http.Handler) http.Handler
		MCPServerMiddleware() func(http.Handler) http.Handler

		CreateUserAuthContext(context.Context, *UserAuthSession) (context.Context, error)
		GetUserAuthSession(context.Context) (*UserAuthSession, error)
		IssueUserAuthSessionToken(*UserAuthSession) (string, error)
		VerifyUserAuthSessionToken(tokenStr string) (*UserAuthSession, error)
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

		ListSystemComponents(context.Context, ListSystemComponentsParams) ([]*ent.SystemComponent, int, error)

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
	UserDataProvider interface {
		UserDataMapping() *ent.User
		PullUsers(ctx context.Context) iter.Seq2[*ent.User, error]
	}
	ListUsersParams = struct {
		ListParams
		TeamID uuid.UUID
	}

	UserService interface {
		Create(context.Context, ent.User) (*ent.User, error)
		ListUsers(context.Context, ListUsersParams) ([]*ent.User, error)
		GetById(context.Context, uuid.UUID) (*ent.User, error)
		GetByEmail(context.Context, string) (*ent.User, error)
		GetByChatId(context.Context, string) (*ent.User, error)
	}
)

type (
	ContentNode        = prosemirror.Node
	DocumentSchemaSpec = prosemirror.SchemaSpec

	DocumentsService interface {
		GetWebsocketAddress() string
		CheckUserDocumentAccess(ctx context.Context, userId uuid.UUID, documentName string) (readOnly bool, err error)
		GetDocumentSchemaSpec(ctx context.Context, schemaName string) (*DocumentSchemaSpec, error)

		CreateOncallShiftHandoverMessage(sections []OncallShiftHandoverSection, annotations []*ent.OncallAnnotation, roster *ent.OncallRoster, endingShift *ent.OncallUserShift, startingShift *ent.OncallUserShift) (*ContentNode, error)
	}
)

type (
	ChatMessageContextProvider struct {
		AnnotateMessageFn        func(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error)
		LookupChatUserFn         func(ctx context.Context, chatId string) (*ent.User, error)
		LookupChatMessageEventFn func(ctx context.Context, msgId string) (*ent.OncallEvent, error)
	}

	ChatEventHandler interface {
		HandleMentionEvent(chatId, threadId, userId, msgText string)
	}

	ChatProvider interface {
		GetWebhooks() Webhooks

		SetMessageContextProvider(ChatMessageContextProvider)
		SetEventHandler(ChatEventHandler)

		SendMessage(ctx context.Context, id string, msg *ContentNode) error
		SendReply(ctx context.Context, channelId string, threadId string, text string) error
		SendTextMessage(ctx context.Context, id string, text string) error

		// TODO: this should just be converted to *ContentNode by ChatService
		SendOncallHandover(ctx context.Context, params SendOncallHandoverParams) error
	}

	SendOncallHandoverParams struct {
		Content           []OncallShiftHandoverSection
		EndingShift       *ent.OncallUserShift
		StartingShift     *ent.OncallUserShift
		PinnedAnnotations []*ent.OncallAnnotation
	}

	ChatService interface {
		ChatEventHandler
		SendOncallHandoverReminder(context.Context, *ent.OncallUserShift) error
		SendOncallHandover(ctx context.Context, params SendOncallHandoverParams) error
	}
)

type (
	AiLanguageModel = model.ToolCallingChatModel

	LanguageModelProvider interface {
		Model() AiLanguageModel
	}

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
		GetWebhooks() Webhooks
		PullAlerts(context.Context) iter.Seq2[*ent.Alert, error]
		PullAlertEventsBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallEvent, error]
	}

	ListAlertsParams struct {
		ListParams
	}

	AlertService interface {
		ListAlerts(context.Context, ListAlertsParams) ([]*ent.Alert, int, error)
		GetAlert(context.Context, uuid.UUID) (*ent.Alert, error)
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
		GetWebhooks() Webhooks

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
		ListIncidents(context.Context, ListIncidentsParams) ([]*ent.Incident, error)
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
	ListRetrospectiveDiscussionsParams struct {
		ListParams
		RetrospectiveID uuid.UUID
		WithReplies     bool
	}

	CreateRetrospectiveDiscussionParams struct {
		RetrospectiveID uuid.UUID
		UserID          uuid.UUID
		Content         []byte
	}

	AddRetrospectiveDiscussionReplyParams struct {
		DiscussionId uuid.UUID
		UserID       uuid.UUID
		ParentID     *uuid.UUID
		Content      []byte
	}

	RetrospectiveService interface {
		Create(context.Context, ent.Retrospective) (*ent.Retrospective, error)
		GetById(context.Context, uuid.UUID) (*ent.Retrospective, error)
		GetForIncident(context.Context, *ent.Incident) (*ent.Retrospective, error)

		CreateDiscussion(context.Context, CreateRetrospectiveDiscussionParams) (*ent.RetrospectiveDiscussion, error)
		ListDiscussions(context.Context, ListRetrospectiveDiscussionsParams) ([]*ent.RetrospectiveDiscussion, error)
		GetDiscussionByID(context.Context, uuid.UUID) (*ent.RetrospectiveDiscussion, error)
		// TODO: just pass a *ent.RetrospectiveDiscussionReply
		AddDiscussionReply(context.Context, AddRetrospectiveDiscussionReplyParams) (*ent.RetrospectiveDiscussionReply, error)
	}
)

type (
	ListOncallEventsParams struct {
		ListParams
		From            time.Time
		To              time.Time
		RosterID        uuid.UUID
		WithAnnotations bool
	}

	ListOncallAnnotationsParams struct {
		ListParams
		From              time.Time
		To                time.Time
		RosterID          uuid.UUID
		Shift             *ent.OncallUserShift
		WithCreator       bool
		WithRoster        bool
		WithAlertFeedback bool
		WithEvent         bool
	}

	OncallEventsService interface {
		GetEvent(ctx context.Context, id uuid.UUID) (*ent.OncallEvent, error)
		ListEvents(ctx context.Context, params ListOncallEventsParams) ([]*ent.OncallEvent, int, error)
		GetProviderEvent(ctx context.Context, providerId string) (*ent.OncallEvent, error)

		ListAnnotations(ctx context.Context, params ListOncallAnnotationsParams) ([]*ent.OncallAnnotation, int, error)

		GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallAnnotation, error)
		UpdateAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error)
		DeleteAnnotation(ctx context.Context, id uuid.UUID) error
	}
)

type (
	OncallDataProvider interface {
		GetWebhooks() Webhooks

		RosterDataMapping() *ent.OncallRoster
		UserShiftDataMapping() *ent.OncallUserShift

		PullRosters(context.Context) iter.Seq2[*ent.OncallRoster, error]
		PullShiftsForRoster(ctx context.Context, rosterId string, from, to time.Time) iter.Seq2[*ent.OncallUserShift, error]
		FetchOncallersForRoster(ctx context.Context, rosterId string) ([]*ent.User, error)
	}

	ListUserOncallParams = struct {
		ListParams
		UserID uuid.UUID
	}
	ListOncallRostersParams   = ListUserOncallParams
	ListOncallSchedulesParams = ListUserOncallParams

	ListOncallShiftsParams struct {
		ListParams
		UserID uuid.UUID
		Anchor time.Time
		Window time.Duration
	}

	OncallShiftHandoverSection struct {
		Header  string            `json:"header"`
		Kind    string            `json:"kind"`
		Content *prosemirror.Node `json:"jsonContent,omitempty"`
	}

	OncallService interface {
		MakeScanShiftsPeriodicJob(context.Context) (*jobs.PeriodicJob, error)
		HandlePeriodicScanShifts(context.Context, jobs.ScanOncallShifts) error

		ListRosters(context.Context, ListOncallRostersParams) ([]*ent.OncallRoster, error)
		GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error)
		GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error)
		GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error)

		ListSchedules(ctx context.Context, params ListOncallSchedulesParams) ([]*ent.OncallSchedule, error)

		ListShifts(ctx context.Context, params ListOncallShiftsParams) ([]*ent.OncallUserShift, error)
		GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error)
		GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, *ent.OncallUserShift, error)

		GetHandoverForShift(ctx context.Context, shiftId uuid.UUID, create bool) (*ent.OncallUserShiftHandover, error)
		GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftHandover, error)
		UpdateShiftHandover(ctx context.Context, handover *ent.OncallUserShiftHandover) (*ent.OncallUserShiftHandover, error)
		SendShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftHandover, error)

		HandleEnsureShiftHandoverSent(context.Context, jobs.EnsureShiftHandoverSent) error
		HandleEnsureShiftHandoverReminderSent(context.Context, jobs.EnsureShiftHandoverReminderSent) error
		HandleGenerateShiftMetrics(context.Context, jobs.GenerateShiftMetrics) error
	}
)
