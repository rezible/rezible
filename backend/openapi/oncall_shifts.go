package openapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

type OncallShiftsHandler interface {
	ListOncallShifts(context.Context, *ListOncallShiftsRequest) (*ListOncallShiftsResponse, error)
	GetOncallShift(context.Context, *GetOncallShiftRequest) (*GetOncallShiftResponse, error)
	GetAdjacentOncallShifts(context.Context, *GetAdjacentOncallShiftsRequest) (*GetAdjacentOncallShiftsResponse, error)

	CreateOncallShiftHandoverTemplate(context.Context, *CreateOncallShiftHandoverTemplateRequest) (*CreateOncallShiftHandoverTemplateResponse, error)
	GetOncallShiftHandoverTemplate(context.Context, *GetOncallShiftHandoverTemplateRequest) (*GetOncallShiftHandoverTemplateResponse, error)
	UpdateOncallShiftHandoverTemplate(context.Context, *UpdateOncallShiftHandoverTemplateRequest) (*UpdateOncallShiftHandoverTemplateResponse, error)
	ArchiveOncallShiftHandoverTemplate(context.Context, *ArchiveOncallShiftHandoverTemplateRequest) (*ArchiveOncallShiftHandoverTemplateResponse, error)

	GetOncallShiftHandover(context.Context, *GetOncallShiftHandoverRequest) (*GetOncallShiftHandoverResponse, error)
	UpdateOncallShiftHandover(context.Context, *UpdateOncallShiftHandoverRequest) (*UpdateOncallShiftHandoverResponse, error)
	SendOncallShiftHandover(context.Context, *SendOncallShiftHandoverRequest) (*SendOncallShiftHandoverResponse, error)
}

func (o operations) RegisterOncallShifts(api huma.API) {
	huma.Register(api, ListOncallShifts, o.ListOncallShifts)

	huma.Register(api, GetOncallShift, o.GetOncallShift)
	huma.Register(api, GetAdjacentOncallShifts, o.GetAdjacentOncallShifts)

	huma.Register(api, CreateOncallShiftHandoverTemplate, o.CreateOncallShiftHandoverTemplate)
	huma.Register(api, GetOncallShiftHandoverTemplate, o.GetOncallShiftHandoverTemplate)
	huma.Register(api, UpdateOncallShiftHandoverTemplate, o.UpdateOncallShiftHandoverTemplate)
	huma.Register(api, ArchiveOncallShiftHandoverTemplate, o.ArchiveOncallShiftHandoverTemplate)

	huma.Register(api, GetOncallShiftHandover, o.GetOncallShiftHandover)
	huma.Register(api, UpdateOncallShiftHandover, o.UpdateOncallShiftHandover)
	huma.Register(api, SendOncallShiftHandover, o.SendOncallShiftHandover)
}

type (
	OncallShift struct {
		Id         uuid.UUID             `json:"id"`
		Attributes OncallShiftAttributes `json:"attributes"`
	}

	OncallShiftAttributes struct {
		User         User         `json:"user"`
		Roster       OncallRoster `json:"roster"`
		Role         string       `json:"role"`
		StartAt      time.Time    `json:"startAt"`
		EndAt        time.Time    `json:"endAt"`
		PrimaryShift *OncallShift `json:"primaryShift"`
	}

	OncallShiftsAdjacent struct {
		Previous *OncallShift `json:"previous,omitempty"`
		Next     *OncallShift `json:"next,omitempty"`
	}

	OncallShiftHandoverTemplate struct {
		Id         uuid.UUID                             `json:"id"`
		Attributes OncallShiftHandoverTemplateAttributes `json:"attributes"`
	}

	OncallShiftHandoverTemplateAttributes struct {
		Sections []OncallShiftHandoverSection `json:"sections"`
	}

	OncallShiftHandover struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes OncallShiftHandoverAttributes `json:"attributes"`
	}

	OncallShiftHandoverAttributes struct {
		ShiftId           uuid.UUID                    `json:"shiftId"`
		SentAt            time.Time                    `json:"sentAt"`
		Content           []OncallShiftHandoverSection `json:"content"`
		PinnedAnnotations []OncallAnnotation           `json:"pinnedAnnotations"`
	}

	OncallShiftHandoverSection struct {
		Header      string  `json:"header"`
		Kind        string  `json:"kind" enum:"regular,annotations"`
		JsonContent *string `json:"jsonContent,omitempty"`
	}
)

func OncallShiftFromEnt(shift *ent.OncallShift) OncallShift {
	attr := OncallShiftAttributes{
		Role:    shift.Role.String(),
		StartAt: shift.StartAt,
		EndAt:   shift.EndAt,
	}

	if shift.Edges.Roster != nil {
		attr.Roster = OncallRosterFromEnt(shift.Edges.Roster)
	} else if shift.RosterID != uuid.Nil {
		attr.Roster = OncallRoster{Id: shift.RosterID}
	}

	if shift.Edges.User != nil {
		attr.User = UserFromEnt(shift.Edges.User)
	} else if shift.UserID != uuid.Nil {
		attr.User = User{Id: shift.UserID}
	}

	if shift.Edges.PrimaryShift != nil {
		primary := OncallShiftFromEnt(shift.Edges.PrimaryShift)
		attr.PrimaryShift = &primary
	} else if shift.PrimaryShiftID != uuid.Nil {
		attr.PrimaryShift = &OncallShift{Id: shift.PrimaryShiftID}
	}

	return OncallShift{
		Id:         shift.ID,
		Attributes: attr,
	}
}

type unmarshalOncallShiftContentSection struct {
	Header      string          `json:"header"`
	Kind        string          `json:"kind" enum:"regular,annotations,incidents"`
	JsonContent json.RawMessage `json:"jsonContent,omitempty"`
}

func OncallShiftHandoverFromEnt(p *ent.OncallShiftHandover) OncallShiftHandover {
	var rawContents []unmarshalOncallShiftContentSection
	if jsonErr := json.Unmarshal(p.Contents, &rawContents); jsonErr != nil {
		// TODO: just return an error
		log.Error().Err(jsonErr).Msg("Error unmarshalling OncallShiftHandover contents")
	}
	content := make([]OncallShiftHandoverSection, len(rawContents))
	for i, rawContent := range rawContents {
		content[i] = OncallShiftHandoverSection{
			Header: rawContent.Header,
			Kind:   rawContent.Kind,
		}
		if rawContent.Kind == "regular" && rawContent.JsonContent != nil {
			str := string(rawContent.JsonContent)
			content[i].JsonContent = &str
		}
	}
	attr := OncallShiftHandoverAttributes{
		ShiftId:           p.ShiftID,
		Content:           content,
		SentAt:            p.SentAt,
		PinnedAnnotations: make([]OncallAnnotation, len(p.Edges.PinnedAnnotations)),
	}
	for i, anno := range p.Edges.PinnedAnnotations {
		attr.PinnedAnnotations[i] = OncallAnnotationFromEnt(anno)
	}

	return OncallShiftHandover{
		Id:         p.ID,
		Attributes: attr,
	}
}

// ops

var oncallShiftsTags = []string{"Oncall Shifts"}

var ListOncallShifts = huma.Operation{
	OperationID: "list-oncall-shifts",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts",
	Summary:     "List Oncall Shifts",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type ListOncallShiftsRequest struct {
	ListRequest
	UserId uuid.UUID `query:"userId" required:"false" nullable:"false"`
	Active bool      `query:"active" required:"false" nullable:"false"`
}
type ListOncallShiftsResponse PaginatedResponse[OncallShift]

var GetOncallShift = huma.Operation{
	OperationID: "get-oncall-shift",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}",
	Summary:     "Get an Oncall Shift",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftRequest GetIdRequest
type GetOncallShiftResponse ItemResponse[OncallShift]

var GetAdjacentOncallShifts = huma.Operation{
	OperationID: "get-adjacent-oncall-shifts",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/adjacent",
	Summary:     "Get shifts adjacent to a given shift",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type GetAdjacentOncallShiftsRequest GetIdRequest
type GetAdjacentOncallShiftsResponse ItemResponse[OncallShiftsAdjacent]

var CreateOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "create-oncall-handover-template",
	Method:      http.MethodPost,
	Path:        "/oncall/handover_templates",
	Summary:     "Create an Oncall Handover Template",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type CreateOncallShiftHandoverTemplateRequestAttributes struct {
	Sections []OncallShiftHandoverSection `json:"sections"`
}
type CreateOncallShiftHandoverTemplateRequest RequestWithBodyAttributes[CreateOncallShiftHandoverTemplateRequestAttributes]
type CreateOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var GetOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "get-oncall-shift-handover-template",
	Method:      http.MethodGet,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Get handover for a shift",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftHandoverTemplateRequest GetIdRequest
type GetOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var UpdateOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "update-oncall-handover-template",
	Method:      http.MethodPatch,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Update an Oncall Handover Template",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftHandoverTemplateRequestAttributes struct {
	Sections []OncallShiftHandoverSection `json:"sections"`
}
type UpdateOncallShiftHandoverTemplateRequest UpdateIdRequest[UpdateOncallShiftHandoverTemplateRequestAttributes]
type UpdateOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var ArchiveOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "archive-oncall-handover-template",
	Method:      http.MethodDelete,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Archive an Oncall Handover Template",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type ArchiveOncallShiftHandoverTemplateRequest ArchiveIdRequest
type ArchiveOncallShiftHandoverTemplateResponse EmptyResponse

var GetOncallShiftHandover = huma.Operation{
	OperationID: "get-oncall-shift-handover",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/handover",
	Summary:     "Get handover for a shift",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftHandoverRequest GetIdRequest
type GetOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var UpdateOncallShiftHandover = huma.Operation{
	OperationID: "update-oncall-shift-handover",
	Method:      http.MethodPatch,
	Path:        "/oncall/handovers/{id}",
	Summary:     "Update an Oncall Shift Handover",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftHandoverAttributes struct {
	Content             *[]OncallShiftHandoverSection `json:"content,omitempty"`
	PinnedAnnotationIds *[]uuid.UUID                  `json:"pinnedAnnotationIds,omitempty"`
}
type UpdateOncallShiftHandoverRequest UpdateIdRequest[UpdateOncallShiftHandoverAttributes]
type UpdateOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var SendOncallShiftHandover = huma.Operation{
	OperationID: "send-oncall-shift-handover",
	Method:      http.MethodPost,
	Path:        "/oncall/handovers/{id}/send",
	Summary:     "Send a Shift Handover",
	Tags:        oncallShiftsTags,
	Errors:      errorCodes(),
}

type SendOncallShiftHandoverAttributes struct {
}
type SendOncallShiftHandoverRequest CreateIdRequest[SendOncallShiftHandoverAttributes]
type SendOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]
