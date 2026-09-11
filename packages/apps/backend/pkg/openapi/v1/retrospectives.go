package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type RetrospectivesHandler interface {
	ListRetrospectives(context.Context, *ListRetrospectivesRequest) (*ListRetrospectivesResponse, error)
	GetRetrospective(context.Context, *GetRetrospectiveRequest) (*GetRetrospectiveResponse, error)
	UpdateRetrospective(context.Context, *UpdateRetrospectiveRequest) (*UpdateRetrospectiveResponse, error)
}

func (o operations) RegisterRetrospectives(api huma.API) {
	huma.Register(api, ListRetrospectives, o.ListRetrospectives)
	huma.Register(api, GetRetrospective, o.GetRetrospective)
	huma.Register(api, UpdateRetrospective, o.UpdateRetrospective)
}

type (
	Retrospective struct {
		Id         uuid.UUID               `json:"id"`
		Attributes RetrospectiveAttributes `json:"attributes"`
	}

	RetrospectiveAttributes struct {
		DocumentId       uuid.UUID                    `json:"documentId"`
		SystemAnalysisId uuid.UUID                    `json:"systemAnalysisId"`
		Kind             string                       `json:"type" enum:"simple,full"`
		State            string                       `json:"state" enum:"draft,in_review,meeting,closed"`
		Reviews          []Review                     `json:"reviews"`
		ReportSections   []RetrospectiveReportSection `json:"reportSections"`
	}

	RetrospectiveReportSection struct {
		Kind        string `json:"kind" enum:"field"`
		Title       string `json:"title"`
		Field       string `json:"field"`
		Description string `json:"description"`
	}
)

func RetrospectiveFromEnt(r *ent.Retrospective) Retrospective {
	attr := RetrospectiveAttributes{
		Reviews:          ConvertSlice(r.Edges.Reviews, ReviewFromEnt),
		DocumentId:       r.DocumentID,
		SystemAnalysisId: r.SystemAnalysisID,
		Kind:             r.Kind.String(),
		State:            r.State.String(),
	}
	// TODO: fetch this
	attr.ReportSections = []RetrospectiveReportSection{
		{
			Kind:        "field",
			Title:       "Background",
			Field:       "background",
			Description: "",
		},
		{
			Kind:        "field",
			Title:       "Lessons Learned",
			Field:       "lessons",
			Description: "",
		},
	}

	return Retrospective{Id: r.ID, Attributes: attr}
}

// Operations

var retrospectivesTags = []string{"Retrospectives"}

var ListRetrospectives = huma.Operation{
	OperationID: "list-retrospectives",
	Method:      http.MethodGet,
	Path:        "/retrospectives",
	Summary:     "List Retrospectives",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(),
}

type ListRetrospectivesRequest struct {
	PaginationRequest
}
type ListRetrospectivesResponse PaginatedResponse[Retrospective]

var GetRetrospective = huma.Operation{
	OperationID: "get-retrospective",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}",
	Summary:     "Get a Retrospective",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(),
}

type GetRetrospectiveRequest IdRequest
type GetRetrospectiveResponse ItemResponse[Retrospective]

var UpdateRetrospective = huma.Operation{
	OperationID: "update-retrospective",
	Method:      http.MethodPatch,
	Path:        "/retrospectives/{id}",
	Summary:     "Create an Incident Retrospective",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type UpdateRetrospectiveAttributes struct {
	Kind  *string `json:"kind,omitempty" enum:"simple,full"`
	State *string `json:"state,omitempty" enum:"draft,in_review,meeting,closed"`
}
type UpdateRetrospectiveRequest IdRequestWithBody[UpdateRetrospectiveAttributes]
type UpdateRetrospectiveResponse ItemResponse[Retrospective]
