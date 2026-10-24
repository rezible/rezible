package v1

import (
	"context"
	"github.com/rezible/rezible/ent/review"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
)

type RetrospectivesHandler interface {
	SkipRetrospective(context.Context, *SkipRetrospectiveRequest) (*SkipRetrospectiveResponse, error)
	CloseRetrospective(context.Context, *CloseRetrospectiveRequest) (*CloseRetrospectiveResponse, error)
	SetRetrospectiveReviewDecision(context.Context, *SetRetrospectiveReviewDecisionRequest) (*SetRetrospectiveReviewDecisionResponse, error)
	RemoveRetrospectiveReviewer(context.Context, *RemoveRetrospectiveReviewerRequest) (*RemoveRetrospectiveReviewerResponse, error)
	AddRetrospectiveReviewer(context.Context, *AddRetrospectiveReviewerRequest) (*AddRetrospectiveReviewerResponse, error)
	RequestRetrospectiveReview(context.Context, *RequestRetrospectiveReviewRequest) (*RequestRetrospectiveReviewResponse, error)
	SetRetrospectiveOwner(context.Context, *SetRetrospectiveOwnerRequest) (*SetRetrospectiveOwnerResponse, error)
	GetRetrospective(context.Context, *GetRetrospectiveRequest) (*GetRetrospectiveResponse, error)

	GetRetrospectiveReportComposition(context.Context, *GetRetrospectiveReportCompositionRequest) (*GetRetrospectiveReportCompositionResponse, error)
	SetRetrospectiveReportFindingSelection(context.Context, *SetRetrospectiveReportFindingSelectionRequest) (*SetRetrospectiveReportFindingSelectionResponse, error)
}

func (o operations) RegisterRetrospectives(api huma.API) {
	huma.Register(api, SkipRetrospective, o.SkipRetrospective)
	huma.Register(api, CloseRetrospective, o.CloseRetrospective)
	huma.Register(api, SetRetrospectiveReviewDecision, o.SetRetrospectiveReviewDecision)
	huma.Register(api, RemoveRetrospectiveReviewer, o.RemoveRetrospectiveReviewer)
	huma.Register(api, AddRetrospectiveReviewer, o.AddRetrospectiveReviewer)
	huma.Register(api, RequestRetrospectiveReview, o.RequestRetrospectiveReview)
	huma.Register(api, SetRetrospectiveOwner, o.SetRetrospectiveOwner)
	huma.Register(api, GetRetrospective, o.GetRetrospective)

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
		Owner            *Expandable[UserAttributes]  `json:"owner,omitempty"`
		ClosedAt         *time.Time                   `json:"closedAt,omitempty"`
		ClosedBy         *Expandable[UserAttributes]  `json:"closedBy,omitempty"`
		SkipReason       *string                      `json:"skipReason,omitempty"`
		Outcome          string                       `json:"outcome,omitempty" enum:"reviewed,closed_without_review,skipped"`
		Reviews          []RetrospectiveReview        `json:"reviews"`
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

type RetrospectiveReview struct {
	Reviewer      Expandable[UserAttributes] `json:"reviewer"`
	State         review.State               `json:"state" enum:"waiting,request_changes,approved"`
	Feedback      *string                    `json:"feedback,omitempty"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	MessageFailed bool                       `json:"messageFailed"`
}

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
		ClosedAt:         r.ClosedAt, SkipReason: r.SkipReason,
		Reviews: make([]RetrospectiveReview, 0, len(r.Edges.Reviews)),
	}
	if r.OwnerID != nil {
		attrs.Owner = &Expandable[UserAttributes]{Id: *r.OwnerID}
		if r.Edges.Owner != nil {
			attrs.Owner.Attributes = new(UserFromEnt(r.Edges.Owner).Attributes)
		}
	}
	if r.ClosedByID != nil {
		attrs.ClosedBy = &Expandable[UserAttributes]{Id: *r.ClosedByID}
		if r.Edges.ClosedBy != nil {
			attrs.ClosedBy.Attributes = new(UserFromEnt(r.Edges.ClosedBy).Attributes)
		}
	}
	for _, row := range r.Edges.Reviews {
		reviewer := Expandable[UserAttributes]{Id: row.ReviewerID}
		if row.Edges.Reviewer != nil {
			reviewer.Attributes = new(UserFromEnt(row.Edges.Reviewer).Attributes)
		}
		attrs.Reviews = append(attrs.Reviews, RetrospectiveReview{Reviewer: reviewer, State: row.State, Feedback: row.Feedback, UpdatedAt: row.UpdatedAt, MessageFailed: row.MessageFailedAt != nil})
	}
	if r.State == retrospective.StateClosed {
		switch {
		case r.SkipReason != nil:
			attrs.Outcome = "skipped"
		case len(r.Edges.Reviews) > 0:
			attrs.Outcome = "reviewed"
		default:
			attrs.Outcome = "closed_without_review"
		}
	}
	return Retrospective{Id: r.ID, Attributes: attrs}
}

func RetrospectiveReportCompositionFromRez(r *rez.RetrospectiveReportComposition) RetrospectiveReportComposition {
	return RetrospectiveReportComposition{}
}

// Operations

var retrospectivesTags = []string{"Retrospectives"}

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

var SetRetrospectiveOwner = huma.Operation{
	OperationID: "set-retrospective-owner", Method: "PUT", Path: "/retrospectives/{id}/owner",
	Summary: "SetRetrospectiveOwner", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type SetRetrospectiveOwnerAttributes struct {
	UserId uuid.UUID `json:"userId"`
}
type SetRetrospectiveOwnerRequest IdRequestWithBody[SetRetrospectiveOwnerAttributes]
type SetRetrospectiveOwnerResponse ItemResponse[Retrospective]

var RequestRetrospectiveReview = huma.Operation{
	OperationID: "request-retrospective-review", Method: "POST", Path: "/retrospectives/{id}/review-request",
	Summary: "RequestRetrospectiveReview", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type RequestRetrospectiveReviewAttributes struct {
	ReviewerIds []uuid.UUID `json:"reviewerIds"`
}
type RequestRetrospectiveReviewRequest IdRequestWithBody[RequestRetrospectiveReviewAttributes]
type RequestRetrospectiveReviewResponse ItemResponse[Retrospective]

var AddRetrospectiveReviewer = huma.Operation{
	OperationID: "add-retrospective-reviewer", Method: "POST", Path: "/retrospectives/{id}/reviewers",
	Summary: "AddRetrospectiveReviewer", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type AddRetrospectiveReviewerAttributes struct {
	UserId uuid.UUID `json:"userId"`
}
type AddRetrospectiveReviewerRequest IdRequestWithBody[AddRetrospectiveReviewerAttributes]
type AddRetrospectiveReviewerResponse ItemResponse[Retrospective]

var RemoveRetrospectiveReviewer = huma.Operation{
	OperationID: "remove-retrospective-reviewer", Method: "DELETE", Path: "/retrospectives/{id}/reviewers/{userId}",
	Summary: "RemoveRetrospectiveReviewer", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type RemoveRetrospectiveReviewerRequest struct {
	Id     uuid.UUID `path:"id"`
	UserId uuid.UUID `path:"userId"`
}
type RemoveRetrospectiveReviewerResponse ItemResponse[Retrospective]

var SetRetrospectiveReviewDecision = huma.Operation{
	OperationID: "set-retrospective-review-decision", Method: "PUT", Path: "/retrospectives/{id}/review-decision",
	Summary: "SetRetrospectiveReviewDecision", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type SetRetrospectiveReviewDecisionAttributes struct {
	Decision review.State `json:"decision" enum:"approved,request_changes"`
	Feedback string       `json:"feedback,omitempty" maxLength:"4000"`
}
type SetRetrospectiveReviewDecisionRequest IdRequestWithBody[SetRetrospectiveReviewDecisionAttributes]
type SetRetrospectiveReviewDecisionResponse ItemResponse[Retrospective]

var CloseRetrospective = huma.Operation{
	OperationID: "close-retrospective", Method: "POST", Path: "/retrospectives/{id}/close",
	Summary: "CloseRetrospective", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type CloseRetrospectiveRequest IdRequest
type CloseRetrospectiveResponse ItemResponse[Retrospective]

var SkipRetrospective = huma.Operation{
	OperationID: "skip-retrospective", Method: "POST", Path: "/retrospectives/{id}/skip",
	Summary: "SkipRetrospective", Tags: retrospectivesTags, Errors: ErrorCodes(http.StatusConflict, http.StatusForbidden),
}

type SkipRetrospectiveAttributes struct {
	Reason string `json:"reason" maxLength:"4000"`
}
type SkipRetrospectiveRequest IdRequestWithBody[SkipRetrospectiveAttributes]
type SkipRetrospectiveResponse ItemResponse[Retrospective]
