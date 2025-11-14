package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
	"time"
)

type IncidentDebriefsHandler interface {
	GetIncidentUserDebrief(context.Context, *GetIncidentUserDebriefRequest) (*GetIncidentUserDebriefResponse, error)
	GetIncidentDebrief(context.Context, *GetIncidentDebriefRequest) (*GetIncidentDebriefResponse, error)
	UpdateIncidentDebrief(context.Context, *UpdateIncidentDebriefRequest) (*UpdateIncidentDebriefResponse, error)

	ListIncidentDebriefMessages(context.Context, *ListIncidentDebriefMessagesRequest) (*ListIncidentDebriefMessagesResponse, error)
	AddIncidentDebriefUserMessage(context.Context, *AddIncidentDebriefUserMessageRequest) (*AddIncidentDebriefUserMessageResponse, error)

	ListIncidentDebriefSuggestions(context.Context, *ListIncidentDebriefSuggestionsRequest) (*ListIncidentDebriefSuggestionsResponse, error)

	ListIncidentDebriefQuestions(context.Context, *ListIncidentDebriefQuestionsRequest) (*ListIncidentDebriefQuestionsResponse, error)
	CreateIncidentDebriefQuestion(context.Context, *CreateIncidentDebriefQuestionRequest) (*CreateIncidentDebriefQuestionResponse, error)
	GetIncidentDebriefQuestion(context.Context, *GetIncidentDebriefQuestionRequest) (*GetIncidentDebriefQuestionResponse, error)
	UpdateIncidentDebriefQuestion(context.Context, *UpdateIncidentDebriefQuestionRequest) (*UpdateIncidentDebriefQuestionResponse, error)
	ArchiveIncidentDebriefQuestion(context.Context, *ArchiveIncidentDebriefQuestionRequest) (*ArchiveIncidentDebriefQuestionResponse, error)
}

func (o operations) RegisterIncidentDebriefs(api huma.API) {
	huma.Register(api, GetIncidentUserDebrief, o.GetIncidentUserDebrief)
	huma.Register(api, GetIncidentDebrief, o.GetIncidentDebrief)
	huma.Register(api, UpdateIncidentDebrief, o.UpdateIncidentDebrief)

	huma.Register(api, ListIncidentDebriefMessages, o.ListIncidentDebriefMessages)
	huma.Register(api, AddIncidentDebriefUserMessage, o.AddIncidentDebriefUserMessage)

	huma.Register(api, ListIncidentDebriefSuggestions, o.ListIncidentDebriefSuggestions)

	huma.Register(api, ListIncidentDebriefQuestions, o.ListIncidentDebriefQuestions)
	huma.Register(api, CreateIncidentDebriefQuestion, o.CreateIncidentDebriefQuestion)
	huma.Register(api, GetIncidentDebriefQuestion, o.GetIncidentDebriefQuestion)
	huma.Register(api, UpdateIncidentDebriefQuestion, o.UpdateIncidentDebriefQuestion)
	huma.Register(api, ArchiveIncidentDebriefQuestion, o.ArchiveIncidentDebriefQuestion)
}

type (
	IncidentDebrief struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes IncidentDebriefAttributes `json:"attributes"`
	}

	IncidentDebriefAttributes struct {
		IncidentId uuid.UUID `json:"incidentId"`
		UserId     uuid.UUID `json:"userId"`
		Required   bool      `json:"required"`
		Started    bool      `json:"started"`
	}

	IncidentDebriefSuggestion struct {
		Id         uuid.UUID                           `json:"id"`
		Attributes IncidentDebriefSuggestionAttributes `json:"attributes"`
	}

	IncidentDebriefSuggestionAttributes struct {
		Ignored bool   `json:"ignored"`
		Content string `json:"content"`
	}

	IncidentDebriefMessage struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes IncidentDebriefMessageAttributes `json:"attributes"`
	}

	IncidentDebriefMessageAttributes struct {
		CreatedAt time.Time `json:"createdAt"`
		Type      string    `json:"type" enum:"user,assistant,question"`
		Body      string    `json:"body"`
	}

	IncidentDebriefQuestion struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes IncidentDebriefQuestionAttributes `json:"attributes"`
	}

	IncidentDebriefQuestionAttributes struct {
		Content string `json:"content"`
	}
)

func IncidentDebriefFromEnt(debrief *ent.IncidentDebrief) IncidentDebrief {
	ret := IncidentDebrief{
		Id: debrief.ID,
		Attributes: IncidentDebriefAttributes{
			IncidentId: debrief.IncidentID,
			UserId:     debrief.UserID,
			Required:   debrief.Required,
			Started:    debrief.Started,
		},
	}
	return ret
}

func IncidentDebriefMessageFromEnt(msg *ent.IncidentDebriefMessage) IncidentDebriefMessage {
	return IncidentDebriefMessage{
		Id: msg.ID,
		Attributes: IncidentDebriefMessageAttributes{
			CreatedAt: msg.CreatedAt,
			Type:      msg.Type.String(),
			Body:      msg.Body,
		},
	}
}

func IncidentDebriefSuggestionFromEnt(sug *ent.IncidentDebriefSuggestion) IncidentDebriefSuggestion {
	return IncidentDebriefSuggestion{
		Id: sug.ID,
		Attributes: IncidentDebriefSuggestionAttributes{
			Ignored: false,
			Content: sug.Content,
		},
	}
}

func IncidentDebriefQuestionFromEnt(q *ent.IncidentDebriefQuestion) IncidentDebriefQuestion {
	return IncidentDebriefQuestion{
		Id: q.ID,
		Attributes: IncidentDebriefQuestionAttributes{
			Content: q.Content,
		},
	}
}

var incidentDebriefsTags = []string{"Incident Debriefs"}

var GetIncidentUserDebrief = huma.Operation{
	OperationID: "get-incident-user-debrief",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/debrief",
	Summary:     "Get Debrief For Incident",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type GetIncidentUserDebriefRequest = GetIdRequest
type GetIncidentUserDebriefResponse ItemResponse[IncidentDebrief]

var GetIncidentDebrief = huma.Operation{
	OperationID: "get-incident-debrief",
	Method:      http.MethodGet,
	Path:        "/incident_debriefs/{id}",
	Summary:     "Get Incident Debrief",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type GetIncidentDebriefRequest = GetIdRequest
type GetIncidentDebriefResponse ItemResponse[IncidentDebrief]

var UpdateIncidentDebrief = huma.Operation{
	OperationID: "update-incident-debrief",
	Method:      http.MethodPatch,
	Path:        "/incident_debriefs/{id}",
	Summary:     "Update Incident Debrief",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentDebriefAttributes struct {
	Status string `json:"status" enum:"started,completed"`
}
type UpdateIncidentDebriefRequest UpdateIdRequest[UpdateIncidentDebriefAttributes]
type UpdateIncidentDebriefResponse ItemResponse[IncidentDebrief]

var ListIncidentDebriefMessages = huma.Operation{
	OperationID: "list-debrief-messages",
	Method:      http.MethodGet,
	Path:        "/incident_debriefs/{id}/messages",
	Summary:     "List Incident Debrief Messages",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type ListIncidentDebriefMessagesRequest ListIdRequest
type ListIncidentDebriefMessagesResponse ItemResponse[[]IncidentDebriefMessage]

var AddIncidentDebriefUserMessage = huma.Operation{
	OperationID: "add-incident-debrief-user-message",
	Method:      http.MethodPost,
	Path:        "/incident_debriefs/{id}/messages",
	Summary:     "Add an Incident Debrief message",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type AddIncidentDebriefUserMessageAttributes struct {
	MessageContent string `json:"messageContent"`
}
type AddIncidentDebriefUserMessageRequest CreateIdRequest[AddIncidentDebriefUserMessageAttributes]
type AddIncidentDebriefUserMessageResponse ItemResponse[IncidentDebriefMessage]

var ListIncidentDebriefSuggestions = huma.Operation{
	OperationID: "list-debrief-suggestions",
	Method:      http.MethodGet,
	Path:        "/incident_debriefs/{id}/suggestions",
	Summary:     "List Incident Debrief Suggestions",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type ListIncidentDebriefSuggestionsRequest ListIdRequest
type ListIncidentDebriefSuggestionsResponse ItemResponse[[]IncidentDebriefSuggestion]

var ListIncidentDebriefQuestions = huma.Operation{
	OperationID: "list-debrief-questions",
	Method:      http.MethodGet,
	Path:        "/debrief_questions",
	Summary:     "List Incident Debrief Questions",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type ListIncidentDebriefQuestionsRequest ListRequest
type ListIncidentDebriefQuestionsResponse PaginatedResponse[IncidentDebriefQuestion]

var GetIncidentDebriefQuestion = huma.Operation{
	OperationID: "get-debrief-question",
	Method:      http.MethodGet,
	Path:        "/debrief_questions/{id}",
	Summary:     "Get an Incident Debrief Question",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type GetIncidentDebriefQuestionRequest GetIdRequest
type GetIncidentDebriefQuestionResponse ItemResponse[IncidentDebriefQuestion]

var CreateIncidentDebriefQuestion = huma.Operation{
	OperationID: "create-debrief-question",
	Method:      http.MethodPost,
	Path:        "/debrief_questions",
	Summary:     "Create an Incident Debrief Question",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type CreateIncidentDebriefQuestionAttributes struct {
	Content string `json:"content"`
}
type CreateIncidentDebriefQuestionRequest RequestWithBodyAttributes[CreateIncidentDebriefQuestionAttributes]
type CreateIncidentDebriefQuestionResponse ItemResponse[IncidentDebriefQuestion]

var UpdateIncidentDebriefQuestion = huma.Operation{
	OperationID: "update-debrief-question",
	Method:      http.MethodPatch,
	Path:        "/debrief_questions/{id}",
	Summary:     "Update an Incident Debrief Question",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentDebriefQuestionAttributes struct {
	Content *string `json:"content,omitempty"`
}
type UpdateIncidentDebriefQuestionRequest UpdateIdRequest[UpdateIncidentDebriefQuestionAttributes]
type UpdateIncidentDebriefQuestionResponse ItemResponse[IncidentDebriefQuestion]

var ArchiveIncidentDebriefQuestion = huma.Operation{
	OperationID: "archive-debrief-question",
	Method:      http.MethodDelete,
	Path:        "/debrief_questions/{id}",
	Summary:     "Archive an Incident Debrief Question",
	Tags:        incidentDebriefsTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentDebriefQuestionRequest ArchiveIdRequest
type ArchiveIncidentDebriefQuestionResponse EmptyResponse
