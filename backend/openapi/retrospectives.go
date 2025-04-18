package openapi

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

	//ListRetrospectiveTemplates(context.Context, *ListRetrospectiveTemplatesRequest) (*ListRetrospectiveTemplatesResponse, error)
	//CreateRetrospectiveTemplate(context.Context, *CreateRetrospectiveTemplateRequest) (*CreateRetrospectiveTemplateResponse, error)
	//UpdateRetrospectiveTemplate(context.Context, *UpdateRetrospectiveTemplateRequest) (*UpdateRetrospectiveTemplateResponse, error)
	//ArchiveRetrospectiveTemplate(context.Context, *ArchiveRetrospectiveTemplateRequest) (*ArchiveRetrospectiveTemplateResponse, error)

	ListRetrospectiveDiscussions(context.Context, *ListRetrospectiveDiscussionsRequest) (*ListRetrospectiveDiscussionsResponse, error)
	CreateRetrospectiveDiscussion(context.Context, *CreateRetrospectiveDiscussionRequest) (*CreateRetrospectiveDiscussionResponse, error)
	GetRetrospectiveDiscussion(context.Context, *GetRetrospectiveDiscussionRequest) (*GetRetrospectiveDiscussionResponse, error)
	UpdateRetrospectiveDiscussion(context.Context, *UpdateRetrospectiveDiscussionRequest) (*UpdateRetrospectiveDiscussionResponse, error)
	AddRetrospectiveDiscussionReply(context.Context, *AddRetrospectiveDiscussionReplyRequest) (*AddRetrospectiveDiscussionReplyResponse, error)
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

	//huma.Register(api, ListRetrospectiveTemplates, o.ListRetrospectiveTemplates)
	//huma.Register(api, CreateRetrospectiveTemplate, o.CreateRetrospectiveTemplate)
	//huma.Register(api, UpdateRetrospectiveTemplate, o.UpdateRetrospectiveTemplate)
	//huma.Register(api, ArchiveRetrospectiveTemplate, o.ArchiveRetrospectiveTemplate)

	huma.Register(api, ListRetrospectiveDiscussions, o.ListRetrospectiveDiscussions)
	huma.Register(api, CreateRetrospectiveDiscussion, o.CreateRetrospectiveDiscussion)
	huma.Register(api, GetRetrospectiveDiscussion, o.GetRetrospectiveDiscussion)
	huma.Register(api, UpdateRetrospectiveDiscussion, o.UpdateRetrospectiveDiscussion)
	huma.Register(api, AddRetrospectiveDiscussionReply, o.AddRetrospectiveDiscussionReply)
}

type (
	Retrospective struct {
		Id         uuid.UUID               `json:"id"`
		Attributes RetrospectiveAttributes `json:"attributes"`
	}

	RetrospectiveAttributes struct {
		Type             RetrospectiveType            `json:"type" enum:"simple,full"`
		State            RetrospectiveState           `json:"state" enum:"draft,in_review,meeting_scheduled,completed"`
		ReportSections   []RetrospectiveReportSection `json:"reportSections"`
		SystemAnalysisId *uuid.UUID                   `json:"systemAnalysisId,omitempty"`
	}

	RetrospectiveType  string
	RetrospectiveState string

	RetrospectiveReview struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes RetrospectiveReviewAttributes `json:"attributes"`
	}

	RetrospectiveReviewAttributes struct {
	}

	RetrospectiveReportTemplate struct {
		Id         uuid.UUID                             `json:"id"`
		Attributes RetrospectiveReportTemplateAttributes `json:"attributes"`
	}

	RetrospectiveReportTemplateAttributes struct {
		Sections []RetrospectiveReportSection `json:"sections"`
	}

	RetrospectiveReportSection struct {
		Type        string `json:"type" enum:"field"`
		Title       string `json:"title"`
		Field       string `json:"field"`
		Description string `json:"description"`
	}

	RetrospectiveDiscussion struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes RetrospectiveDiscussionAttributes `json:"attributes"`
	}

	RetrospectiveDiscussionAttributes struct {
		DocumentAnnotationId *uuid.UUID                     `json:"annotationId,omitempty"`
		Resolved             bool                           `json:"resolved"`
		Content              string                         `json:"content"`
		Replies              []RetrospectiveDiscussionReply `json:"replies"`
	}

	RetrospectiveDiscussionReply struct {
		Id         uuid.UUID                              `json:"id"`
		Attributes RetrospectiveDiscussionReplyAttributes `json:"attributes"`
	}

	RetrospectiveDiscussionReplyAttributes struct {
		Content string                         `json:"content"`
		Replies []RetrospectiveDiscussionReply `json:"replies"`
	}
)

func RetrospectiveFromEnt(r *ent.Retrospective) Retrospective {
	attr := RetrospectiveAttributes{
		Type:  RetrospectiveType(r.Type.String()),
		State: RetrospectiveState(r.State.String()),
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

func RetrospectiveDiscussionFromEnt(d *ent.RetrospectiveDiscussion) RetrospectiveDiscussion {
	replies := make([]RetrospectiveDiscussionReply, len(d.Edges.Replies))
	for i, r := range d.Edges.Replies {
		replies[i] = RetrospectiveDiscussionReplyFromEnt(r)
	}

	return RetrospectiveDiscussion{
		Id: d.ID,
		Attributes: RetrospectiveDiscussionAttributes{
			Resolved: false,
			Content:  string(d.Content),
			Replies:  replies,
		},
	}
}

func RetrospectiveDiscussionReplyFromEnt(r *ent.RetrospectiveDiscussionReply) RetrospectiveDiscussionReply {
	replies := make([]RetrospectiveDiscussionReply, len(r.Edges.Replies))
	for i, rr := range r.Edges.Replies {
		replies[i] = RetrospectiveDiscussionReplyFromEnt(rr)
	}

	return RetrospectiveDiscussionReply{
		Id: r.ID,
		Attributes: RetrospectiveDiscussionReplyAttributes{
			Content: string(r.Content),
			Replies: replies,
		},
	}
}

// Operations

var retrospectivesTags = []string{"Retrospectives"}
var retrospectiveDiscussionTags = []string{"Retrospective Discussions"}

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

/*
var ListRetrospectiveTemplates = huma.Operation{
	OperationID: "list-retrospective-templates",
	Method:      http.MethodGet,
	Path:        "/retrospective_templates",
	Summary:     "Get a Retrospective Template",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ListRetrospectiveTemplatesRequest ListRequest
type ListRetrospectiveTemplatesResponse PaginatedResponse[RetrospectiveTemplate]

var CreateRetrospectiveTemplate = huma.Operation{
	OperationID: "create-retrospective-template",
	Method:      http.MethodPost,
	Path:        "/retrospective_templates",
	Summary:     "Create a Retrospective Template",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type CreateRetrospectiveTemplateAttributes struct {
	Sections *[]RetrospectiveSection `json:"sections,omitempty"`
}
type CreateRetrospectiveTemplateRequest RequestWithBodyAttributes[CreateRetrospectiveTemplateAttributes]
type CreateRetrospectiveTemplateResponse ItemResponse[RetrospectiveTemplate]

var UpdateRetrospectiveTemplate = huma.Operation{
	OperationID: "update-retrospective-template",
	Method:      http.MethodPatch,
	Path:        "/retrospective_templates/{id}",
	Summary:     "Update a Retrospective Template",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type UpdateRetrospectiveTemplateAttributes struct {
	Sections *[]RetrospectiveSection `json:"sections,omitempty"`
}
type UpdateRetrospectiveTemplateRequest UpdateIdRequest[UpdateRetrospectiveTemplateAttributes]
type UpdateRetrospectiveTemplateResponse ItemResponse[RetrospectiveTemplate]

var ArchiveRetrospectiveTemplate = huma.Operation{
	OperationID: "archive-retrospective-template",
	Method:      http.MethodDelete,
	Path:        "/retrospective_templates/{id}",
	Summary:     "Archive a Retrospective Template",
	Tags:        retrospectivesTags,
	Errors:      errorCodes(),
}

type ArchiveRetrospectiveTemplateRequest ArchiveIdRequest
type ArchiveRetrospectiveTemplateResponse EmptyResponse
*/

var ListRetrospectiveDiscussions = huma.Operation{
	OperationID: "list-retrospective-discussions",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}/discussions",
	Summary:     "List Discussions For a Retrospective",
	Tags:        retrospectiveDiscussionTags,
	Errors:      errorCodes(),
}

type ListRetrospectiveDiscussionsRequest ListIdRequest
type ListRetrospectiveDiscussionsResponse PaginatedResponse[RetrospectiveDiscussion]

type retrospectiveDiscussionRequest struct {
	RetrospectiveId uuid.UUID `path:"id"`
	DiscussionId    uuid.UUID `path:"discussionId"`
}

var GetRetrospectiveDiscussion = huma.Operation{
	OperationID: "get-retrospective-discussion",
	Method:      http.MethodGet,
	Path:        "/retrospectives/{id}/discussions/{discussionId}",
	Summary:     "Get a Retrospective Discussion",
	Tags:        retrospectiveDiscussionTags,
	Errors:      errorCodes(),
}

type GetRetrospectiveDiscussionRequest retrospectiveDiscussionRequest
type GetRetrospectiveDiscussionResponse ItemResponse[RetrospectiveDiscussion]

var CreateRetrospectiveDiscussion = huma.Operation{
	OperationID: "create-retrospective-discussion",
	Method:      http.MethodPost,
	Path:        "/retrospectives/{id}/discussions",
	Summary:     "Create a Retrospective Discussion",
	Tags:        retrospectiveDiscussionTags,
	Errors:      errorCodes(),
}

type CreateRetrospectiveDiscussionAttributes struct {
	Content json.RawMessage `json:"content"`
}
type CreateRetrospectiveDiscussionRequest CreateIdRequest[CreateRetrospectiveDiscussionAttributes]
type CreateRetrospectiveDiscussionResponse ItemResponse[RetrospectiveDiscussion]

var UpdateRetrospectiveDiscussion = huma.Operation{
	OperationID: "update-retrospective-discussion",
	Method:      http.MethodPatch,
	Path:        "/retrospectives/{id}/discussions/{discussionId}",
	Summary:     "Update a Retrospective Discussion",
	Tags:        retrospectiveDiscussionTags,
	Errors:      errorCodes(),
}

type UpdateRetrospectiveDiscussionAttributes struct {
	Resolved *bool `json:"resolved,omitempty"`
}
type UpdateRetrospectiveDiscussionRequest struct {
	retrospectiveDiscussionRequest
	RequestWithBodyAttributes[UpdateRetrospectiveDiscussionAttributes]
}
type UpdateRetrospectiveDiscussionResponse ItemResponse[RetrospectiveDiscussion]

var AddRetrospectiveDiscussionReply = huma.Operation{
	OperationID: "add-retrospective-discussion-reply",
	Method:      http.MethodPost,
	Path:        "/retrospectives/{id}/discussions/{discussionId}",
	Summary:     "Add a Reply to a Retrospective Discussion",
	Tags:        retrospectiveDiscussionTags,
	Errors:      errorCodes(),
}

type AddRetrospectiveDiscussionReplyRequestAttributes struct {
	ParentReplyId *uuid.UUID      `json:"parentReplyId,omitempty"`
	Content       json.RawMessage `json:"content"`
}
type AddRetrospectiveDiscussionReplyRequest struct {
	retrospectiveDiscussionRequest
	RequestWithBodyAttributes[AddRetrospectiveDiscussionReplyRequestAttributes]
}
type AddRetrospectiveDiscussionReplyResponse ItemResponse[RetrospectiveDiscussion]
