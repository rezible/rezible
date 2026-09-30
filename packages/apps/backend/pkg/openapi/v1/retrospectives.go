package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
)

type RetrospectivesHandler interface {
	ListRetrospectives(context.Context, *ListRetrospectivesRequest) (*ListRetrospectivesResponse, error)
	GetRetrospective(context.Context, *GetRetrospectiveRequest) (*GetRetrospectiveResponse, error)
	UpdateRetrospective(context.Context, *UpdateRetrospectiveRequest) (*UpdateRetrospectiveResponse, error)

	RequestRetrospectiveReview(context.Context, *RequestRetrospectiveReviewRequest) (*RequestRetrospectiveReviewResponse, error)

	GetRetrospectiveReportComposition(context.Context, *GetRetrospectiveReportCompositionRequest) (*GetRetrospectiveReportCompositionResponse, error)
	SetRetrospectiveReportFindingSelection(context.Context, *SetRetrospectiveReportFindingSelectionRequest) (*SetRetrospectiveReportFindingSelectionResponse, error)
}

func (o operations) RegisterRetrospectives(api huma.API) {
	huma.Register(api, ListRetrospectives, o.ListRetrospectives)
	huma.Register(api, GetRetrospective, o.GetRetrospective)
	huma.Register(api, UpdateRetrospective, o.UpdateRetrospective)

	huma.Register(api, RequestRetrospectiveReview, o.RequestRetrospectiveReview)

	huma.Register(api, GetRetrospectiveReportComposition, o.GetRetrospectiveReportComposition)
	huma.Register(api, SetRetrospectiveReportFindingSelection, o.SetRetrospectiveReportFindingSelection)
}

func (o operations) RegisterRetrospectiveEnums(api huma.API) {
	registerEnumAlias[retrospective.State, retrospectiveStateSchema](api)
}

type retrospectiveStateSchema retrospective.State

func (retrospectiveStateSchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(retrospective.StateValues)
}

type (
	Retrospective struct {
		Id         uuid.UUID               `json:"id"`
		Attributes RetrospectiveAttributes `json:"attributes"`
	}

	RetrospectiveAttributes struct {
		DocumentId       uuid.UUID                    `json:"documentId"`
		SystemAnalysisId uuid.UUID                    `json:"systemAnalysisId"`
		State            retrospective.State          `json:"state"`
		ReportSections   []RetrospectiveReportSection `json:"reportSections"`
	}

	RetrospectiveReportComposition struct {
		Id         uuid.UUID                                `json:"id"`
		Attributes RetrospectiveReportCompositionAttributes `json:"attributes"`
	}

	RetrospectiveReportCompositionAttributes struct {
		IncidentId       uuid.UUID                                   `json:"incidentId"`
		DocumentId       uuid.UUID                                   `json:"documentId"`
		SystemAnalysisId uuid.UUID                                   `json:"systemAnalysisId"`
		Sections         []RetrospectiveReportCompositionSection     `json:"sections"`
		Findings         []Expandable[SystemAnalysisEntryAttributes] `json:"findings"`
		Tasks            []Task                                      `json:"tasks"`
	}

	RetrospectiveReportCompositionSection struct {
		Key         string  `json:"key" enum:"summary,customer-impact,background,findings,lessons,follow-up-actions,sources"`
		Title       string  `json:"title"`
		Kind        string  `json:"kind" enum:"narrative,findings,tasks,sources"`
		FragmentKey *string `json:"fragmentKey,omitempty"`
	}

	RetrospectiveReportSection struct {
		Kind        string `json:"kind" enum:"field"`
		Title       string `json:"title"`
		Field       string `json:"field"`
		Description string `json:"description"`
	}
)

// TODO: load this somewhere
var defaultReportSections = []RetrospectiveReportSection{
	{
		Kind:        "field",
		Title:       "Summary",
		Field:       "summary",
		Description: "",
	},
	{
		Kind:        "field",
		Title:       "Customer impact",
		Field:       "customer-impact",
		Description: "",
	},
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

func RetrospectiveFromEnt(r *ent.Retrospective) Retrospective {
	attrs := RetrospectiveAttributes{
		DocumentId:       r.DocumentID,
		SystemAnalysisId: r.SystemAnalysisID,
		State:            r.State,
		ReportSections:   defaultReportSections,
	}
	return Retrospective{Id: r.ID, Attributes: attrs}
}

func RetrospectiveReportCompositionFromRez(r *rez.RetrospectiveReportComposition) RetrospectiveReportComposition {
	return RetrospectiveReportComposition{}
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

var RequestRetrospectiveReview = huma.Operation{
	OperationID: "request-retrospective-review",
	Method:      http.MethodPost,
	Path:        "/retrospectives/{id}/review-request",
	Summary:     "Request Retrospective Review",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type RequestRetrospectiveReviewAttributes struct {
	ReviewerId    uuid.UUID `json:"reviewerId"`
	ReviewMessage *string   `json:"reviewMessage,omitempty"`
}
type RequestRetrospectiveReviewRequest IdRequestWithBody[RequestRetrospectiveReviewAttributes]
type RequestRetrospectiveReviewResponse ItemResponse[Review]

var GetRetrospectiveReportComposition = huma.Operation{
	OperationID: "get-retrospective-report-composition",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}/report",
	Summary:     "Get Incident Report Composition",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(),
}

type GetRetrospectiveReportCompositionRequest IdRequest
type GetRetrospectiveReportCompositionResponse ItemResponse[RetrospectiveReportComposition]

var SetRetrospectiveReportFindingSelection = huma.Operation{
	OperationID: "set-retrospective-report-findings",
	Method:      http.MethodPut,
	Path:        "/retrospectives/{id}/report/findings",
	Summary:     "Set Incident Report Selected Findings",
	Tags:        retrospectivesTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type SetRetrospectiveReportFindingSelectionRequestAttributes struct {
	AnalysisEntryIDs []uuid.UUID `json:"analysisEntryIDs"`
}

type SetRetrospectiveReportFindingSelectionRequest IdRequestWithBody[SetRetrospectiveReportFindingSelectionRequestAttributes]
type SetRetrospectiveReportFindingSelectionResponse CollectionResponse[SystemAnalysisEntry]
