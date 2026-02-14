package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type MeetingsHandler interface {
	ListMeetingSchedules(context.Context, *ListMeetingSchedulesRequest) (*ListMeetingSchedulesResponse, error)
	GetMeetingSchedule(context.Context, *GetMeetingScheduleRequest) (*GetMeetingScheduleResponse, error)
	CreateMeetingSchedule(context.Context, *CreateMeetingScheduleRequest) (*CreateMeetingScheduleResponse, error)
	UpdateMeetingSchedule(context.Context, *UpdateMeetingScheduleRequest) (*UpdateMeetingScheduleResponse, error)
	ArchiveMeetingSchedule(context.Context, *ArchiveMeetingScheduleRequest) (*ArchiveMeetingScheduleResponse, error)

	ListMeetingSessions(context.Context, *ListMeetingSessionsRequest) (*ListMeetingSessionsResponse, error)
	CreateMeetingSession(context.Context, *CreateMeetingSessionRequest) (*CreateMeetingSessionResponse, error)
	GetMeetingSession(context.Context, *GetMeetingSessionRequest) (*GetMeetingSessionResponse, error)
	UpdateMeetingSession(context.Context, *UpdateMeetingSessionRequest) (*UpdateMeetingSessionResponse, error)
	ArchiveMeetingSession(context.Context, *ArchiveMeetingSessionRequest) (*ArchiveMeetingSessionResponse, error)
}

func (o operations) RegisterMeetings(api huma.API) {
	huma.Register(api, ListMeetingSchedules, o.ListMeetingSchedules)
	huma.Register(api, GetMeetingSchedule, o.GetMeetingSchedule)
	huma.Register(api, CreateMeetingSchedule, o.CreateMeetingSchedule)
	huma.Register(api, UpdateMeetingSchedule, o.UpdateMeetingSchedule)
	huma.Register(api, ArchiveMeetingSchedule, o.ArchiveMeetingSchedule)

	huma.Register(api, ListMeetingSessions, o.ListMeetingSessions)
	huma.Register(api, CreateMeetingSession, o.CreateMeetingSession)
	huma.Register(api, GetMeetingSession, o.GetMeetingSession)
	huma.Register(api, UpdateMeetingSession, o.UpdateMeetingSession)
	huma.Register(api, ArchiveMeetingSession, o.ArchiveMeetingSession)
}

type (
	VideoConference struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes VideoConferenceAttributes `json:"attributes"`
	}

	VideoConferenceAttributes struct {
		Provider             string                 `json:"provider"`
		ExternalId           string                 `json:"externalId,omitempty"`
		JoinUrl              string                 `json:"joinUrl"`
		HostUrl              string                 `json:"hostUrl,omitempty"`
		DialIn               string                 `json:"dialIn,omitempty"`
		Passcode             string                 `json:"passcode,omitempty"`
		Status               string                 `json:"status"`
		CreatedByIntegration string                 `json:"createdByIntegration,omitempty"`
		Metadata             map[string]interface{} `json:"metadata,omitempty"`
	}

	MeetingSchedule struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes MeetingScheduleAttributes `json:"attributes"`
	}

	MeetingScheduleAttributes struct {
		Name               string                `json:"name"`
		SessionTitle       string                `json:"sessionTitle"`
		Attendees          MeetingAttendees      `json:"attendees"`
		HostTeamId         uuid.UUID             `json:"hostTeamId"`
		DocumentTemplateId uuid.UUID             `json:"documentTemplateId"`
		Timing             MeetingScheduleTiming `json:"timing"`
	}

	MeetingAttendees struct {
		Private bool     `json:"private"`
		Users   []string `json:"users"`
		Teams   []string `json:"teams"`
	}

	MeetingScheduleTiming struct {
		StartAt             time.Time `json:"startsAt"`
		DurationMinutes     int       `json:"durationMinutes"`
		Repeat              string    `json:"repeat" enum:"daily,weekly,monthly"`
		RepeatStep          int       `json:"repeatStep"`
		RepeatMonthlyOn     *string   `json:"repeatMonthlyOn" enum:"same_day,same_weekday"`
		Indefinite          bool      `json:"indefinite"`
		UntilDate           *string   `json:"untilDate,omitempty"`
		UntilNumRepetitions *int      `json:"untilNumRepetitions,omitempty"`
	}

	MeetingDocumentTemplate struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes MeetingDocumentTemplateAttributes `json:"attributes"`
	}

	MeetingDocumentTemplateAttributes struct {
		Version string `json:"version"`
		Content string `json:"content"`
	}

	MeetingSession struct {
		Id         uuid.UUID                `json:"id"`
		Attributes MeetingSessionAttributes `json:"attributes"`
	}

	MeetingSessionAttributes struct {
		MeetingScheduleId uuid.UUID        `json:"meetingScheduleId"`
		StartsAt          time.Time        `json:"startsAt"`
		Title             string           `json:"title"`
		HostTeamId        uuid.UUID        `json:"hostTeamId"`
		Attendees         MeetingAttendees `json:"attendees"`
		DocumentName      string           `json:"documentName"`
	}
)

func VideoConferenceFromEnt(c *ent.VideoConference) VideoConference {
	attrs := VideoConferenceAttributes{
		Provider:             c.Provider,
		ExternalId:           c.ExternalID,
		JoinUrl:              c.JoinURL,
		HostUrl:              c.HostURL,
		DialIn:               c.DialIn,
		Passcode:             c.Passcode,
		Status:               string(c.Status),
		CreatedByIntegration: c.CreatedByIntegration,
		Metadata:             map[string]any{},
	}

	if len(c.Metadata) > 0 {
		_ = json.Unmarshal(c.Metadata, &attrs.Metadata)
	}

	return VideoConference{Id: c.ID, Attributes: attrs}
}

// Operations

var meetingsTags = []string{"Meetings"}

var ListMeetingSchedules = huma.Operation{
	OperationID: "list-meeting-schedules",
	Method:      http.MethodGet,
	Path:        "/meeting_schedules",
	Summary:     "List Meeting Schedules",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type ListMeetingSchedulesRequest struct {
	ListRequest
}
type ListMeetingSchedulesResponse PaginatedResponse[MeetingSchedule]

var GetMeetingSchedule = huma.Operation{
	OperationID: "get-meeting-schedule",
	Method:      http.MethodGet,
	Path:        "/meeting_schedules/{id}",
	Summary:     "Get a Meeting Schedule",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type GetMeetingScheduleRequest GetIdRequest
type GetMeetingScheduleResponse ItemResponse[MeetingSchedule]

var CreateMeetingSchedule = huma.Operation{
	OperationID: "create-meeting-schedule",
	Method:      http.MethodPost,
	Path:        "/meeting_schedules",
	Summary:     "Create a Meeting Schedule",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type CreateMeetingScheduleAttributes struct {
	Name            string           `json:"name"`
	SessionTitle    string           `json:"sessionTitle"`
	Description     *string          `json:"description,omitempty"`
	StartsAt        time.Time        `json:"startsAt"`
	DurationMinutes int              `json:"durationMinutes"`
	Attendees       MeetingAttendees `json:"attendees"`
	Repeats         string           `json:"repeats" enum:"daily,weekly,monthly"`
	RepetitionStep  int              `json:"repetitionStep"`
	// Weekdays        []string                    `json:"weekdays"`
	RepeatMonthlyOn string `json:"repeatMonthlyOn,omitempty" enum:"same_day,same_weekday"`
	UntilDate       string `json:"untilDate,omitempty" format:"date"`
	NumRepetitions  int    `json:"numRepetitions,omitempty"`
}
type CreateMeetingScheduleRequest RequestWithBodyAttributes[CreateMeetingScheduleAttributes]
type CreateMeetingScheduleResponse ItemResponse[MeetingSchedule]

var UpdateMeetingSchedule = huma.Operation{
	OperationID: "update-meeting-schedule",
	Method:      http.MethodPatch,
	Path:        "/meeting_schedules/{id}",
	Summary:     "Update a Meeting Schedule",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type UpdateMeetingScheduleAttributes struct {
}
type UpdateMeetingScheduleRequest UpdateIdRequest[UpdateMeetingScheduleAttributes]
type UpdateMeetingScheduleResponse ItemResponse[MeetingSchedule]

var ArchiveMeetingSchedule = huma.Operation{
	OperationID: "archive-meeting-schedule",
	Method:      http.MethodDelete,
	Path:        "/meeting_schedules/{id}",
	Summary:     "Archive a Meeting Schedule",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type ArchiveMeetingScheduleRequest ArchiveIdRequest
type ArchiveMeetingScheduleResponse EmptyResponse

var ListMeetingSessions = huma.Operation{
	OperationID: "list-meeting-sessions",
	Method:      http.MethodGet,
	Path:        "/meeting_sessions",
	Summary:     "List Sessions",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type ListMeetingSessionsRequest struct {
	ListRequest
	MeetingScheduleId uuid.UUID `query:"meetingScheduleId" required:"false"`
	UserId            uuid.UUID `query:"userId" required:"false"`
	TeamId            uuid.UUID `query:"teamId" required:"false"`
	From              string    `query:"from" required:"false"`
	To                string    `query:"to" required:"false"`
}
type ListMeetingSessionsResponse PaginatedResponse[MeetingSession]

var GetMeetingSession = huma.Operation{
	OperationID: "get-meeting-session",
	Method:      http.MethodGet,
	Path:        "/meeting_sessions/{id}",
	Summary:     "Get a Meeting Session",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type GetMeetingSessionRequest GetIdRequest
type GetMeetingSessionResponse ItemResponse[MeetingSession]

var CreateMeetingSession = huma.Operation{
	OperationID: "create-meeting-session",
	Method:      http.MethodPost,
	Path:        "/meeting_sessions",
	Summary:     "Create a Meeting Session",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type CreateMeetingSessionAttributes struct {
	Title              string           `json:"title"`
	Description        *string          `json:"description,omitempty"`
	Attendees          MeetingAttendees `json:"attendees"`
	StartsAt           time.Time        `json:"startsAt"`
	DurationMinutes    int              `json:"durationMinutes"`
	DocumentTemplateId *uuid.UUID       `json:"documentTemplateId,omitempty"`
}
type CreateMeetingSessionRequest RequestWithBodyAttributes[CreateMeetingSessionAttributes]
type CreateMeetingSessionResponse ItemResponse[MeetingSession]

var UpdateMeetingSession = huma.Operation{
	OperationID: "update-meeting-session",
	Method:      http.MethodPatch,
	Path:        "/meeting_sessions/{id}",
	Summary:     "Update a Meeting Session",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type UpdateMeetingSessionAttributes struct {
}
type UpdateMeetingSessionRequest UpdateIdRequest[UpdateMeetingSessionAttributes]
type UpdateMeetingSessionResponse ItemResponse[MeetingSession]

var ArchiveMeetingSession = huma.Operation{
	OperationID: "archive-meeting-session",
	Method:      http.MethodDelete,
	Path:        "/meeting_sessions/{id}",
	Summary:     "Archive a Meeting Session",
	Tags:        meetingsTags,
	Errors:      errorCodes(),
}

type ArchiveMeetingSessionRequest ArchiveIdRequest
type ArchiveMeetingSessionResponse EmptyResponse
