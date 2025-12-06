package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type RetrospectivesHandler interface {
	ListRetrospectives(context.Context, *ListRetrospectivesRequest) (*ListRetrospectivesResponse, error)
	GetRetrospective(context.Context, *GetRetrospectiveRequest) (*GetRetrospectiveResponse, error)
	CreateRetrospective(context.Context, *CreateRetrospectiveRequest) (*CreateRetrospectiveResponse, error)
	GetRetrospectiveForIncident(context.Context, *GetRetrospectiveForIncidentRequest) (*GetRetrospectiveForIncidentResponse, error)

	ListRetrospectiveReviews(context.Context, *ListRetrospectiveReviewsRequest) (*ListRetrospectiveReviewsResponse, error)
	CreateRetrospectiveReview(context.Context, *CreateRetrospectiveReviewRequest) (*CreateRetrospectiveReviewResponse, error)
	UpdateRetrospectiveReview(context.Context, *UpdateRetrospectiveReviewRequest) (*UpdateRetrospectiveReviewResponse, error)
	ArchiveRetrospectiveReview(context.Context, *ArchiveRetrospectiveReviewRequest) (*ArchiveRetrospectiveReviewResponse, error)

	ListRetrospectiveComments(context.Context, *ListRetrospectiveCommentsRequest) (*ListRetrospectiveCommentsResponse, error)
	CreateRetrospectiveComment(context.Context, *CreateRetrospectiveCommentRequest) (*CreateRetrospectiveCommentResponse, error)
	GetRetrospectiveComment(context.Context, *GetRetrospectiveCommentRequest) (*GetRetrospectiveCommentResponse, error)
	UpdateRetrospectiveComment(context.Context, *UpdateRetrospectiveCommentRequest) (*UpdateRetrospectiveCommentResponse, error)
}

func (o operations) RegisterRetrospectives(api huma.API) {
	huma.Register(api, ListRetrospectives, o.ListRetrospectives)
	huma.Register(api, GetRetrospective, o.GetRetrospective)
	huma.Register(api, CreateRetrospective, o.CreateRetrospective)
	huma.Register(api, GetRetrospectiveForIncident, o.GetRetrospectiveForIncident)

	huma.Register(api, ListRetrospectiveReviews, o.ListRetrospectiveReviews)
	huma.Register(api, CreateRetrospectiveReview, o.CreateRetrospectiveReview)
	huma.Register(api, UpdateRetrospectiveReview, o.UpdateRetrospectiveReview)
	huma.Register(api, ArchiveRetrospectiveReview, o.ArchiveRetrospectiveReview)

	huma.Register(api, ListRetrospectiveComments, o.ListRetrospectiveComments)
	huma.Register(api, CreateRetrospectiveComment, o.CreateRetrospectiveComment)
	huma.Register(api, GetRetrospectiveComment, o.GetRetrospectiveComment)
	huma.Register(api, UpdateRetrospectiveComment, o.UpdateRetrospectiveComment)
}

type (
	Retrospective struct {
		Id         uuid.UUID               `json:"id"`
		Attributes RetrospectiveAttributes `json:"attributes"`
	}

	RetrospectiveAttributes struct {
		DocumentId       uuid.UUID                    `json:"documentId"`
		SystemAnalysisId *uuid.UUID                   `json:"systemAnalysisId,omitempty"`
		Type             RetrospectiveType            `json:"type" enum:"simple,full"`
		State            RetrospectiveState           `json:"state" enum:"draft,in_review,meeting_scheduled,completed"`
		ReportSections   []RetrospectiveReportSection `json:"reportSections"`
	}

	RetrospectiveType  string
	RetrospectiveState string

	RetrospectiveReview struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes RetrospectiveReviewAttributes `json:"attributes"`
	}

	RetrospectiveReviewAttributes struct {
		Requester Expandable[User]                           `json:"requester"`
		Reviewer  Expandable[User]                           `json:"reviewer"`
		Comment   Expandable[RetrospectiveCommentAttributes] `json:"comment"`
	}

	RetrospectiveReportSection struct {
		Type        string `json:"type" enum:"field"`
		Title       string `json:"title"`
		Field       string `json:"field"`
		Description string `json:"description"`
	}

	RetrospectiveComment struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes RetrospectiveCommentAttributes `json:"attributes"`
	}

	RetrospectiveCommentAttributes struct {
		User    User                   `json:"user"`
		Content string                 `json:"content"`
		Replies []RetrospectiveComment `json:"replies"`
	}
)

func RetrospectiveFromEnt(r *ent.Retrospective) Retrospective {
	attr := RetrospectiveAttributes{
		DocumentId: r.DocumentID,
		Type:       RetrospectiveType(r.Type.String()),
		State:      RetrospectiveState(r.State.String()),
	}
	if r.SystemAnalysisID != uuid.Nil {
		attr.SystemAnalysisId = &r.SystemAnalysisID
	}

	// TODO: fetch this
	attr.ReportSections = []RetrospectiveReportSection{
		{
			Type:        "field",
			Title:       "Background",
			Field:       "background",
			Description: "",
		},
		{
			Type:        "field",
			Title:       "Lessons Learned",
			Field:       "lessons",
			Description: "",
		},
	}

	return Retrospective{Id: r.ID, Attributes: attr}
}

func RetrospectiveCommentFromEnt(r *ent.RetrospectiveComment) RetrospectiveComment {
	replies := make([]RetrospectiveComment, len(r.Edges.Replies))
	for i, rr := range r.Edges.Replies {
		replies[i] = RetrospectiveCommentFromEnt(rr)
	}

	return RetrospectiveComment{
		Id: r.ID,
		Attributes: RetrospectiveCommentAttributes{
			Content: string(r.Content),
			Replies: replies,
		},
	}
}

// Operations

var retrospectivesTags = []string{"Retrospectives"}

var ListRetrospectives = huma.Operation{
	OperationID: "list-retrospectives",
	Method:      http.MethodGet,
	Path:        "/retrospectives",
	Summary:     "List Retrospectives",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ListRetrospectivesRequest ListRequest
type ListRetrospectivesResponse PaginatedResponse[Retrospective]

var GetRetrospective = huma.Operation{
	OperationID: "get-retrospective",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}",
	Summary:     "Get a Retrospective",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type GetRetrospectiveRequest GetIdRequest
type GetRetrospectiveResponse ItemResponse[Retrospective]

var GetRetrospectiveForIncident = huma.Operation{
	OperationID: "get-retrospective-for-incident",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/retrospective",
	Summary:     "Get a Retrospective for an Incident",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type GetRetrospectiveForIncidentRequest = GetFlexibleIdRequest
type GetRetrospectiveForIncidentResponse ItemResponse[Retrospective]

var CreateRetrospective = huma.Operation{
	OperationID: "create-retrospective",
	Method:      http.MethodPost,
	Path:        "/retrospectives",
	Summary:     "Create an Incident Retrospective",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type CreateRetrospectiveAttributes struct {
	IncidentId     uuid.UUID `json:"incidentId" required:"true"`
	SystemAnalysis bool      `json:"systemAnalysis"`
}
type CreateRetrospectiveRequest RequestWithBodyAttributes[CreateRetrospectiveAttributes]
type CreateRetrospectiveResponse ItemResponse[Retrospective]

// Reviews

var ListRetrospectiveReviews = huma.Operation{
	OperationID: "list-retrospective-reviews",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}/reviews",
	Summary:     "List Retrospective Reviews",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ListRetrospectiveReviewsRequest ListRequest
type ListRetrospectiveReviewsResponse PaginatedResponse[RetrospectiveReview]

var CreateRetrospectiveReview = huma.Operation{
	OperationID: "create-retrospective-review",
	Method:      http.MethodPost,
	Path:        "/retrospectives/{id}/reviews",
	Summary:     "Create a Retrospective Review",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type CreateRetrospectiveReviewRequest CreateIdRequest[RetrospectiveReviewAttributes]
type CreateRetrospectiveReviewResponse ItemResponse[RetrospectiveReview]

var UpdateRetrospectiveReview = huma.Operation{
	OperationID: "update-retrospective-review",
	Method:      http.MethodPatch,
	Path:        "/retrospective_reviews/{id}",
	Summary:     "Update a Retrospective Review",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type UpdateRetrospectiveReviewRequestAttributes struct {
}
type UpdateRetrospectiveReviewRequest UpdateIdRequest[UpdateRetrospectiveReviewRequestAttributes]
type UpdateRetrospectiveReviewResponse ItemResponse[RetrospectiveReview]

var ArchiveRetrospectiveReview = huma.Operation{
	OperationID: "archive-retrospective-review",
	Method:      http.MethodDelete,
	Path:        "/retrospective_reviews/{id}",
	Summary:     "Archive a Retrospective Review",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ArchiveRetrospectiveReviewRequest ArchiveIdRequest
type ArchiveRetrospectiveReviewResponse EmptyResponse

var ListRetrospectiveComments = huma.Operation{
	OperationID: "list-retrospective-comments",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}/comments",
	Summary:     "List Comments For a Retrospective",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ListRetrospectiveCommentsRequest ListIdRequest
type ListRetrospectiveCommentsResponse PaginatedResponse[RetrospectiveComment]

var CreateRetrospectiveComment = huma.Operation{
	OperationID: "create-retrospective-comment",
	Method:      http.MethodPost,
	Path:        "/retrospectives/{id}/comments",
	Summary:     "Create a Retrospective Comment",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type CreateRetrospectiveCommentAttributes struct {
	Content json.RawMessage `json:"content"`
}
type CreateRetrospectiveCommentRequest CreateIdRequest[CreateRetrospectiveCommentAttributes]
type CreateRetrospectiveCommentResponse ItemResponse[RetrospectiveComment]

var GetRetrospectiveComment = huma.Operation{
	OperationID: "get-retrospective-comment",
	Method:      http.MethodGet,
	Path:        "/retrospective_comments/{id}",
	Summary:     "Get a Retrospective Comment",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type GetRetrospectiveCommentRequest ListIdRequest
type GetRetrospectiveCommentResponse ItemResponse[RetrospectiveComment]

var UpdateRetrospectiveComment = huma.Operation{
	OperationID: "update-retrospective-comment",
	Method:      http.MethodPatch,
	Path:        "/retrospective_comments/{id}",
	Summary:     "Update a Retrospective Comment",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type UpdateRetrospectiveCommentAttributes struct {
	Resolved *bool   `json:"resolved,omitempty"`
	Content  *string `json:"content,omitempty"`
}
type UpdateRetrospectiveCommentRequest UpdateIdRequest[UpdateRetrospectiveCommentAttributes]
type UpdateRetrospectiveCommentResponse ItemResponse[RetrospectiveComment]
