package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type ReviewsHandler interface {
	ListReviews(context.Context, *ListReviewsRequest) (*ListReviewsResponse, error)
	GetReview(context.Context, *GetReviewRequest) (*GetReviewResponse, error)
	CreateReview(context.Context, *CreateReviewRequest) (*CreateReviewResponse, error)
	UpdateReview(context.Context, *UpdateReviewRequest) (*UpdateReviewResponse, error)
	ArchiveReview(context.Context, *ArchiveReviewRequest) (*ArchiveReviewResponse, error)
}

func (o operations) RegisterReviews(api huma.API) {
	huma.Register(api, ListReviews, o.ListReviews)
	huma.Register(api, GetReview, o.GetReview)
	huma.Register(api, CreateReview, o.CreateReview)
	huma.Register(api, UpdateReview, o.UpdateReview)
	huma.Register(api, ArchiveReview, o.ArchiveReview)
}

type (
	Review struct {
		Id         uuid.UUID        `json:"id"`
		Attributes ReviewAttributes `json:"attributes"`
	}

	ReviewAttributes struct {
		CreatedAt       time.Time  `json:"createdAt"`
		UpdatedAt       time.Time  `json:"updatedAt"`
		RequesterId     uuid.UUID  `json:"requesterId"`
		ReviewerId      uuid.UUID  `json:"reviewerId"`
		CommentId       *uuid.UUID `json:"commentId,omitempty"`
		State           string     `json:"state" enum:"waiting,request_changes,approved"`
		RetrospectiveId *uuid.UUID `json:"retrospectiveId,omitempty"`
		AnalysisEntryId *uuid.UUID `json:"analysisEntryId,omitempty"`
	}
)

func ReviewFromEnt(v *ent.Review) Review {
	attr := ReviewAttributes{
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		RequesterId: v.RequesterID,
		ReviewerId:  v.ReviewerID,
		State:       v.State.String(),
	}
	if v.RetrospectiveID != nil {
		attr.RetrospectiveId = v.RetrospectiveID
	}
	if v.AnalysisEntryID != nil {
		attr.AnalysisEntryId = v.AnalysisEntryID
	}
	if v.CommentID != nil {
		attr.CommentId = v.CommentID
	}
	return Review{Id: v.ID, Attributes: attr}
}

var reviewsTags = []string{"Reviews"}

var ListReviews = huma.Operation{
	OperationID: "list-reviews",
	Method:      http.MethodGet,
	Path:        "/reviews",
	Summary:     "List Reviews",
	Tags:        reviewsTags,
	Errors:      ErrorCodes(),
}

type ListReviewsRequest struct {
	PaginationRequest
	RetrospectiveId uuid.UUID `query:"retrospectiveId"`
	AnalysisEntryId uuid.UUID `query:"analysisEntryId"`
}

type ListReviewsResponse PaginatedResponse[Review]

var GetReview = huma.Operation{
	OperationID: "get-review",
	Method:      http.MethodGet,
	Path:        "/reviews/{id}",
	Summary:     "Get Review",
	Tags:        reviewsTags,
	Errors:      ErrorCodes(),
}

type GetReviewRequest IdRequest
type GetReviewResponse ItemResponse[Review]

var CreateReview = huma.Operation{
	OperationID: "create-review",
	Method:      http.MethodPost,
	Path:        "/reviews",
	Summary:     "Create Review",
	Tags:        reviewsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type CreateReviewAttributes struct {
	RetrospectiveId *uuid.UUID `json:"retrospectiveId,omitempty"`
	AnalysisEntryId *uuid.UUID `json:"analysisEntryId,omitempty"`
	RequesterId     uuid.UUID  `json:"requesterId"`
	ReviewerId      uuid.UUID  `json:"reviewerId"`
	CommentId       *uuid.UUID `json:"commentId,omitempty"`
	State           string     `json:"state" enum:"waiting,request_changes,approved"`
}

type CreateReviewRequest RequestWithBodyAttributes[CreateReviewAttributes]
type CreateReviewResponse ItemResponse[Review]

var UpdateReview = huma.Operation{
	OperationID: "update-review",
	Method:      http.MethodPatch,
	Path:        "/reviews/{id}",
	Summary:     "Update Review",
	Tags:        reviewsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type UpdateReviewAttributes struct {
	State     OmittableNullable[string]    `json:"state,omitempty" enum:"waiting,request_changes,approved"`
	CommentId OmittableNullable[uuid.UUID] `json:"commentId,omitempty"`
}

type UpdateReviewRequest IdRequestWithBody[UpdateReviewAttributes]
type UpdateReviewResponse ItemResponse[Review]

var ArchiveReview = huma.Operation{
	OperationID: "archive-review",
	Method:      http.MethodDelete,
	Path:        "/reviews/{id}",
	Summary:     "Archive Review",
	Tags:        reviewsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type ArchiveReviewRequest IdRequest
type ArchiveReviewResponse EmptyResponse
