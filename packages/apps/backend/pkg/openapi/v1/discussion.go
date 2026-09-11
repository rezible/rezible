package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type DiscussionHandler interface {
	ListDiscussionThreads(context.Context, *ListDiscussionThreadsRequest) (*ListDiscussionThreadsResponse, error)
	GetDiscussionThread(context.Context, *GetDiscussionThreadRequest) (*GetDiscussionThreadResponse, error)
	CreateDiscussionThread(context.Context, *CreateDiscussionThreadRequest) (*CreateDiscussionThreadResponse, error)
	UpdateDiscussionThread(context.Context, *UpdateDiscussionThreadRequest) (*UpdateDiscussionThreadResponse, error)

	ListDiscussionComments(context.Context, *ListDiscussionCommentsRequest) (*ListDiscussionCommentsResponse, error)
	GetDiscussionComment(context.Context, *GetDiscussionCommentRequest) (*GetDiscussionCommentResponse, error)
	CreateDiscussionComment(context.Context, *CreateDiscussionCommentRequest) (*CreateDiscussionCommentResponse, error)
	UpdateDiscussionComment(context.Context, *UpdateDiscussionCommentRequest) (*UpdateDiscussionCommentResponse, error)
}

func (o operations) RegisterDiscussion(api huma.API) {
	huma.Register(api, ListDiscussionThreads, o.ListDiscussionThreads)
	huma.Register(api, GetDiscussionThread, o.GetDiscussionThread)
	huma.Register(api, CreateDiscussionThread, o.CreateDiscussionThread)
	huma.Register(api, UpdateDiscussionThread, o.UpdateDiscussionThread)

	huma.Register(api, ListDiscussionComments, o.ListDiscussionComments)
	huma.Register(api, GetDiscussionComment, o.GetDiscussionComment)
	huma.Register(api, CreateDiscussionComment, o.CreateDiscussionComment)
	huma.Register(api, UpdateDiscussionComment, o.UpdateDiscussionComment)
}

type (
	DiscussionThread struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes DiscussionThreadAttributes `json:"attributes"`
	}

	DiscussionThreadAttributes struct {
		AnalysisId      *uuid.UUID `json:"analysisId,omitempty"`
		RetrospectiveId *uuid.UUID `json:"retrospectiveId,omitempty"`
		UserId          uuid.UUID  `json:"userId"`
		Kind            string     `json:"kind" enum:"comment,question"`
		TargetKind      *string    `json:"targetKind,omitempty" enum:"finding,knowledge_entity,knowledge_relationship,normalized_event"`
		TargetId        *uuid.UUID `json:"targetId,omitempty"`
		ResolutionState *string    `json:"resolutionState,omitempty" enum:"open,resolved"`
		ResolvedById    *uuid.UUID `json:"resolvedById,omitempty"`
		ResolutionNote  *string    `json:"resolutionNote,omitempty"`
		CreatedAt       time.Time  `json:"createdAt"`
		UpdatedAt       time.Time  `json:"updatedAt"`
		ResolvedAt      *time.Time `json:"resolvedAt,omitempty"`
	}

	DiscussionComment struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes DiscussionCommentAttributes `json:"attributes"`
	}

	DiscussionCommentAttributes struct {
		ThreadId  uuid.UUID  `json:"threadId"`
		ParentId  *uuid.UUID `json:"parentId,omitempty"`
		UserId    uuid.UUID  `json:"userId"`
		Content   string     `json:"content"`
		CreatedAt time.Time  `json:"createdAt"`
		UpdatedAt time.Time  `json:"updatedAt"`
	}
)

func DiscussionThreadFromEnt(v *ent.DiscussionThread) DiscussionThread {
	return DiscussionThread{
		Id: v.ID,
		Attributes: DiscussionThreadAttributes{
			AnalysisId:      v.AnalysisID,
			RetrospectiveId: v.RetrospectiveID,
			UserId:          v.UserID,
			Kind:            v.Kind.String(),
			TargetKind:      discussionString(v.TargetKind),
			TargetId:        v.TargetID,
			ResolutionState: discussionString(v.ResolutionState),
			ResolvedById:    v.ResolvedByID,
			ResolutionNote:  v.ResolutionNote,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
			ResolvedAt:      v.ResolvedAt,
		},
	}
}

func DiscussionCommentFromEnt(v *ent.DiscussionComment) DiscussionComment {
	return DiscussionComment{
		Id: v.ID,
		Attributes: DiscussionCommentAttributes{
			ThreadId:  v.ThreadID,
			ParentId:  v.ParentID,
			UserId:    v.UserID,
			Content:   string(v.Content),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		},
	}
}

func discussionString[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

var discussionErrors = ErrorCodes(http.StatusNotImplemented)
var ListDiscussionThreads = huma.Operation{
	OperationID: "list-discussion-threads",
	Method:      http.MethodGet,
	Path:        "/discussion-threads",
	Summary:     "List discussion threads",
	Errors:      ErrorCodes(),
}

type ListDiscussionThreadsRequest struct {
	PaginationRequest
	DiscussionOwnerQuery
	TargetKind      string    `query:"targetKind,omitempty" enum:"finding,knowledge_entity,knowledge_relationship,normalized_event"`
	TargetId        uuid.UUID `query:"targetId,omitempty"`
	Kind            string    `query:"kind,omitempty" enum:"comment,question"`
	ResolutionState string    `query:"resolutionState,omitempty" enum:"open,resolved"`
}
type DiscussionOwnerQuery struct {
	AnalysisId      uuid.UUID `query:"analysisId,omitempty"`
	RetrospectiveId uuid.UUID `query:"retrospectiveId,omitempty"`
}
type ListDiscussionThreadsResponse PaginatedResponse[DiscussionThread]

var GetDiscussionThread = huma.Operation{
	OperationID: "get-discussion-thread",
	Method:      http.MethodGet,
	Path:        "/discussion-threads/{id}",
	Summary:     "Get discussion thread",
	Errors:      ErrorCodes(),
}

type GetDiscussionThreadRequest IdRequest
type GetDiscussionThreadResponse ItemResponse[DiscussionThread]

var CreateDiscussionThread = huma.Operation{
	OperationID: "create-discussion-thread",
	Method:      http.MethodPost,
	Path:        "/discussion-threads",
	Summary:     "Create discussion thread",
	Errors:      discussionErrors,
}

type CreateDiscussionThreadAttributes struct {
	AnalysisId      *uuid.UUID `json:"analysisId,omitempty"`
	RetrospectiveId *uuid.UUID `json:"retrospectiveId,omitempty"`
	Kind            string     `json:"kind" enum:"comment,question"`
	TargetKind      *string    `json:"targetKind,omitempty" enum:"finding,knowledge_entity,knowledge_relationship,normalized_event"`
	TargetId        *uuid.UUID `json:"targetId,omitempty"`
	InitialMessage  string     `json:"initialMessage"`
}
type CreateDiscussionThreadRequest RequestWithBodyAttributes[CreateDiscussionThreadAttributes]
type CreateDiscussionThreadResponse ItemResponse[DiscussionThread]

var UpdateDiscussionThread = huma.Operation{
	OperationID: "update-discussion-thread",
	Method:      http.MethodPatch,
	Path:        "/discussion-threads/{id}",
	Summary:     "Update discussion thread",
	Errors:      discussionErrors,
}

type UpdateDiscussionThreadAttributes struct {
	ResolutionState string  `json:"resolutionState" enum:"open,resolved"`
	ResolutionNote  *string `json:"resolutionNote,omitempty"`
}
type UpdateDiscussionThreadRequest IdRequestWithBody[UpdateDiscussionThreadAttributes]
type UpdateDiscussionThreadResponse ItemResponse[DiscussionThread]

var ListDiscussionComments = huma.Operation{
	OperationID: "list-discussion-comments",
	Method:      http.MethodGet,
	Path:        "/discussion-threads/{id}/comments",
	Summary:     "List discussion comments",
	Errors:      ErrorCodes(),
}

type ListDiscussionCommentsRequest struct {
	PaginationRequest
	IdRequest
	ParentId uuid.UUID `query:"parentId,omitempty"`
}
type ListDiscussionCommentsResponse PaginatedResponse[DiscussionComment]

var GetDiscussionComment = huma.Operation{
	OperationID: "get-discussion-comment",
	Method:      http.MethodGet,
	Path:        "/discussion-comments/{id}",
	Summary:     "Get discussion comment",
	Errors:      ErrorCodes(),
}

type GetDiscussionCommentRequest IdRequest
type GetDiscussionCommentResponse ItemResponse[DiscussionComment]

var CreateDiscussionComment = huma.Operation{
	OperationID: "create-discussion-comment",
	Method:      http.MethodPost,
	Path:        "/discussion-threads/{id}/comments",
	Summary:     "Create discussion comment",
	Errors:      discussionErrors,
}

type CreateDiscussionCommentAttributes struct {
	Content  string     `json:"content"`
	ParentId *uuid.UUID `json:"parentId,omitempty"`
}
type CreateDiscussionCommentRequest IdRequestWithBody[CreateDiscussionCommentAttributes]
type CreateDiscussionCommentResponse ItemResponse[DiscussionComment]

var UpdateDiscussionComment = huma.Operation{
	OperationID: "update-discussion-comment",
	Method:      http.MethodPatch,
	Path:        "/discussion-comments/{id}",
	Summary:     "Update discussion comment",
	Errors:      discussionErrors,
}

type UpdateDiscussionCommentAttributes struct {
	Content string `json:"content"`
}
type UpdateDiscussionCommentRequest IdRequestWithBody[UpdateDiscussionCommentAttributes]
type UpdateDiscussionCommentResponse ItemResponse[DiscussionComment]
