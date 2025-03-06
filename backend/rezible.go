package rez

import (
	"context"
	"errors"
	"github.com/rezible/rezible/jobs"
	"iter"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/texm/prosemirror-go"
	"github.com/tmc/langchaingo/llms"
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
	BackgroundJobService interface {
		RegisterWorkers(UserService, IncidentService, OncallService, AlertsService, DebriefService, SystemComponentsService) error
		Start(ctx context.Context) error
		Stop(ctx context.Context) error

		Insert(ctx context.Context, args jobs.JobArgs, opts *jobs.InsertOpts) error
		InsertMany(ctx context.Context, params []jobs.InsertManyParams) error
		InsertTx(ctx context.Context, tx *ent.Tx, args jobs.JobArgs, opts *jobs.InsertOpts) error
	}
)

type (
	ProviderLoader interface {
		HandleWebhookRequest(w http.ResponseWriter, r *http.Request)

		LoadAiModelProvider(context.Context) (AiModelProvider, error)
		LoadChatProvider(context.Context) (ChatProvider, error)
		LoadOncallDataProvider(context.Context) (OncallDataProvider, error)
		LoadAlertsDataProvider(context.Context) (AlertsDataProvider, error)
		LoadIncidentDataProvider(context.Context) (IncidentDataProvider, error)
		LoadAuthSessionProvider(context.Context) (AuthSessionProvider, error)
		LoadUserDataProvider(context.Context) (UserDataProvider, error)
		LoadSystemComponentsDataProvider(context.Context) (SystemComponentsDataProvider, error)
	}
	DataProviderResourceUpdatedCallback = func(providerID string, updatedAt time.Time)
)

type (
	AuthSession struct {
		ExpiresAt time.Time
		UserId    uuid.UUID
	}
	AuthSessionCreatedFn = func(*ent.User, time.Time, string)

	AuthSessionProvider interface {
		GetUserMapping() *ent.User
		StartAuthFlow(w http.ResponseWriter, r *http.Request)
		HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated AuthSessionCreatedFn) (handled bool)
		ClearSession(w http.ResponseWriter, r *http.Request)
	}

	AuthService interface {
		MakeAuthHandler() http.Handler
		MakeRequireAuthMiddleware(redirectStartFlow bool) func(http.Handler) http.Handler
		GetSession(context.Context) (*AuthSession, error)
		IssueSessionToken(*AuthSession) (string, error)
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

	SystemComponentsService interface {
		SyncData(context.Context) error

		Create(context.Context, ent.SystemComponent) (*ent.SystemComponent, error)

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
		Create(context.Context, ent.Team) (*ent.Team, error)
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
		SyncData(context.Context) error

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
	}
)

type (
	SendOncallHandoverParams struct {
		Content []OncallShiftHandoverSection

		EndingShift   *ent.OncallUserShift
		StartingShift *ent.OncallUserShift
		Incidents     []*ent.Incident
		Annotations   []*ent.OncallUserShiftAnnotation
	}

	LookupProviderUserFunc              = func(context.Context, string) (*ent.User, error)
	ChatInteractionFuncCreateAnnotation = func(ctx context.Context, shiftId uuid.UUID, msgId string, setFn func(*ent.OncallUserShiftAnnotation)) error

	ChatProvider interface {
		GetWebhooks() Webhooks

		// SetCreateAnnotationLookupUserFunc(ChatInteractionFuncLookupUser)
		SetLookupUserFunc(LookupProviderUserFunc)

		SetCreateAnnotationFunc(ChatInteractionFuncCreateAnnotation)

		// TODO: just use a generic SendMessage(rez.ContentNode), and convert in chat client
		SendOncallHandover(context.Context, SendOncallHandoverParams) error

		// SendUserMessage(ctx context.Context, user *ent.User, msg *ContentNode) error

		SendUserMessage(ctx context.Context, id string, msgText string) error
		SendUserLinkMessage(ctx context.Context, id string, msgText string, linkUrl string, linkText string) error
	}

	ChatService interface {
		SetCreateAnnotationFunc(ChatInteractionFuncCreateAnnotation)

		SendOncallHandover(context.Context, SendOncallHandoverParams) error

		// SendUserMessage(ctx context.Context, user *ent.User, msg *ContentNode) error

		SendUserMessage(ctx context.Context, user *ent.User, msgText string) error
		SendUserLinkMessage(ctx context.Context, user *ent.User, msgText string, linkUrl string, linkText string) error
	}
)

type (
	AiModel = llms.Model

	AiModelProvider interface {
		GetModel() AiModel
	}

	AiService interface {
		GenerateDebriefResponse(context.Context, *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error)
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
		SyncData(context.Context) error

		GetByID(context.Context, uuid.UUID) (*ent.Incident, error)
		GetIdForSlug(context.Context, string) (uuid.UUID, error)
		GetBySlug(context.Context, string) (*ent.Incident, error)
		ListIncidents(context.Context, ListIncidentsParams) ([]*ent.Incident, error)
	}
)

type (
	DebriefService interface {
		SendUserDebriefRequests(ctx context.Context, incidentId uuid.UUID) error

		CreateDebrief(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (*ent.IncidentDebrief, error)
		GetDebrief(ctx context.Context, id uuid.UUID) (*ent.IncidentDebrief, error)
		GetUserDebrief(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (*ent.IncidentDebrief, error)
		AddUserDebriefMessage(ctx context.Context, debriefID uuid.UUID, text string) (*ent.IncidentDebriefMessage, error)
		GenerateResponse(ctx context.Context, debriefId uuid.UUID) error
		StartDebrief(ctx context.Context, debriefID uuid.UUID) (*ent.IncidentDebrief, error)
		CompleteDebrief(ctx context.Context, debriefID uuid.UUID) (*ent.IncidentDebrief, error)
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
		GetByIncidentId(context.Context, uuid.UUID) (*ent.Retrospective, error)

		CreateDiscussion(context.Context, CreateRetrospectiveDiscussionParams) (*ent.RetrospectiveDiscussion, error)
		ListDiscussions(context.Context, ListRetrospectiveDiscussionsParams) ([]*ent.RetrospectiveDiscussion, error)
		GetDiscussionByID(context.Context, uuid.UUID) (*ent.RetrospectiveDiscussion, error)
		AddDiscussionReply(context.Context, AddRetrospectiveDiscussionReplyParams) (*ent.RetrospectiveDiscussionReply, error)
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

	listForShiftIdParams struct {
		ListParams
		ShiftID uuid.UUID
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
	ListOncallShiftEventsParams      = listForShiftIdParams
	ListOncallShiftAnnotationsParams struct {
		ListParams
		ShiftID uuid.UUID
		Pinned  *bool
	}

	OncallShiftHandoverSection struct {
		Header  string            `json:"header"`
		Kind    string            `json:"kind"`
		Content *prosemirror.Node `json:"jsonContent,omitempty"`
	}

	OncallService interface {
		SyncData(context.Context) error
		ScanForShiftsNeedingHandover(context.Context) error

		ListRosters(context.Context, ListOncallRostersParams) ([]*ent.OncallRoster, error)
		GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error)
		GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error)
		GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error)

		ListSchedules(ctx context.Context, params ListOncallSchedulesParams) ([]*ent.OncallSchedule, error)

		ListShifts(ctx context.Context, params ListOncallShiftsParams) ([]*ent.OncallUserShift, error)
		GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error)
		GetNextShift(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error)

		ListShiftAnnotations(ctx context.Context, params ListOncallShiftAnnotationsParams) ([]*ent.OncallUserShiftAnnotation, error)
		GetShiftAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftAnnotation, error)
		CreateShiftAnnotation(ctx context.Context, anno *ent.OncallUserShiftAnnotation) (*ent.OncallUserShiftAnnotation, error)
		ArchiveShiftAnnotation(ctx context.Context, id uuid.UUID) error

		GetRosterHandoverTemplate(ctx context.Context, rosterId uuid.UUID) (*ent.OncallHandoverTemplate, error)
		GetShiftHandover(ctx context.Context, shiftId uuid.UUID) (*ent.OncallUserShiftHandover, error)
		EnsureShiftHandover(ctx context.Context, shiftId uuid.UUID) error
		SendShiftHandover(ctx context.Context, id uuid.UUID, contents []OncallShiftHandoverSection) (*ent.OncallUserShiftHandover, error)
	}
)

type (
	AlertsDataProvider interface {
		GetWebhooks() Webhooks

		SetOnAlertInstanceUpdatedCallback(DataProviderResourceUpdatedCallback)
		PullAlertInstancesBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallAlertInstance, error]
	}

	ListAlertsParams struct {
		ListParams
		Start time.Time
		End   time.Time
	}

	AlertsService interface {
		SyncData(ctx context.Context) error
		ListAlerts(ctx context.Context, params ListAlertsParams) ([]*ent.OncallAlert, error)
	}
)
