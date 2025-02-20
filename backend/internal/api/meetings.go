package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
	"time"
)

type meetingsHandler struct{}

func newMeetingsHandler() *meetingsHandler {
	return &meetingsHandler{}
}

func (h *meetingsHandler) ListMeetingSchedules(ctx context.Context, request *oapi.ListMeetingSchedulesRequest) (*oapi.ListMeetingSchedulesResponse, error) {
	var resp oapi.ListMeetingSchedulesResponse

	sched1 := oapi.MeetingSchedule{
		Id: uuid.New(),
		Attributes: oapi.MeetingScheduleAttributes{
			Name: "Weekly Example Meeting",
			Attendees: oapi.MeetingAttendees{
				Private: false,
				Users:   make([]string, 0),
				Teams:   make([]string, 0),
			},
			HostTeamId:         uuid.New(),
			DocumentTemplateId: uuid.New(),
			Timing: oapi.MeetingScheduleTiming{
				StartAt:         time.Now().Add(time.Hour),
				DurationMinutes: 60,
				Repeat:          "weekly",
				RepeatStep:      1,
				Indefinite:      true,
			},
		},
	}
	resp.Body.Data = []oapi.MeetingSchedule{sched1}

	return &resp, nil
}

func (h *meetingsHandler) GetMeetingSchedule(ctx context.Context, request *oapi.GetMeetingScheduleRequest) (*oapi.GetMeetingScheduleResponse, error) {
	var resp oapi.GetMeetingScheduleResponse

	resp.Body.Data = oapi.MeetingSchedule{
		Id: uuid.New(),
		Attributes: oapi.MeetingScheduleAttributes{
			Name: "Weekly Example Meeting",
			Attendees: oapi.MeetingAttendees{
				Private: false,
				Users:   make([]string, 0),
				Teams:   make([]string, 0),
			},
			HostTeamId:         uuid.New(),
			DocumentTemplateId: uuid.New(),
			Timing: oapi.MeetingScheduleTiming{
				StartAt:         time.Now().Add(time.Hour),
				DurationMinutes: 60,
				Repeat:          "weekly",
				RepeatStep:      1,
				Indefinite:      true,
			},
		},
	}

	return &resp, nil
}

func (h *meetingsHandler) CreateMeetingSchedule(ctx context.Context, request *oapi.CreateMeetingScheduleRequest) (*oapi.CreateMeetingScheduleResponse, error) {
	var resp oapi.CreateMeetingScheduleResponse

	return &resp, nil
}

func (h *meetingsHandler) UpdateMeetingSchedule(ctx context.Context, request *oapi.UpdateMeetingScheduleRequest) (*oapi.UpdateMeetingScheduleResponse, error) {
	var resp oapi.UpdateMeetingScheduleResponse

	return &resp, nil
}

func (h *meetingsHandler) ArchiveMeetingSchedule(ctx context.Context, request *oapi.ArchiveMeetingScheduleRequest) (*oapi.ArchiveMeetingScheduleResponse, error) {
	var resp oapi.ArchiveMeetingScheduleResponse

	return &resp, nil
}

func (h *meetingsHandler) ListMeetingSessions(ctx context.Context, request *oapi.ListMeetingSessionsRequest) (*oapi.ListMeetingSessionsResponse, error) {
	var resp oapi.ListMeetingSessionsResponse

	return &resp, nil
}

func (h *meetingsHandler) CreateMeetingSession(ctx context.Context, request *oapi.CreateMeetingSessionRequest) (*oapi.CreateMeetingSessionResponse, error) {
	var resp oapi.CreateMeetingSessionResponse

	return &resp, nil
}

func (h *meetingsHandler) GetMeetingSession(ctx context.Context, request *oapi.GetMeetingSessionRequest) (*oapi.GetMeetingSessionResponse, error) {
	var resp oapi.GetMeetingSessionResponse

	return &resp, nil
}

func (h *meetingsHandler) UpdateMeetingSession(ctx context.Context, request *oapi.UpdateMeetingSessionRequest) (*oapi.UpdateMeetingSessionResponse, error) {
	var resp oapi.UpdateMeetingSessionResponse

	return &resp, nil
}

func (h *meetingsHandler) ArchiveMeetingSession(ctx context.Context, request *oapi.ArchiveMeetingSessionRequest) (*oapi.ArchiveMeetingSessionResponse, error) {
	var resp oapi.ArchiveMeetingSessionResponse

	return &resp, nil
}
