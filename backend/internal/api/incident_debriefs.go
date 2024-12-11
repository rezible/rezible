package api

import (
	"context"
	"github.com/rs/zerolog/log"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/incidentdebriefmessage"
	"github.com/twohundreds/rezible/ent/schema"
	oapi "github.com/twohundreds/rezible/openapi"
)

type incidentDebriefsHandler struct {
	questions *ent.IncidentDebriefQuestionClient
	auth      rez.AuthService
	users     rez.UserService
	incidents rez.IncidentService
}

func newIncidentDebriefsHandler(questions *ent.IncidentDebriefQuestionClient, auth rez.AuthService, users rez.UserService, incidents rez.IncidentService) *incidentDebriefsHandler {
	return &incidentDebriefsHandler{questions, auth, users, incidents}
}

func (h *incidentDebriefsHandler) GetIncidentUserDebrief(ctx context.Context, request *oapi.GetIncidentUserDebriefRequest) (*oapi.GetIncidentUserDebriefResponse, error) {
	var resp oapi.GetIncidentUserDebriefResponse

	sess, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get auth session", sessErr)
	}

	debrief, debriefErr := h.incidents.GetUserDebrief(ctx, request.Id, sess.User.ID)
	if debriefErr != nil {
		if !ent.IsNotFound(debriefErr) {
			return nil, detailError("failed to get incident debrief", debriefErr)
		}
		created, createErr := h.incidents.CreateDebrief(ctx, request.Id, sess.User.ID)
		if createErr != nil {
			return nil, detailError("failed to create debrief", createErr)
		}
		debrief = created
	}
	resp.Body.Data = oapi.IncidentDebriefFromEnt(debrief)

	return &resp, nil
}

func (h *incidentDebriefsHandler) GetIncidentDebrief(ctx context.Context, request *oapi.GetIncidentDebriefRequest) (*oapi.GetIncidentDebriefResponse, error) {
	var resp oapi.GetIncidentDebriefResponse

	_, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get auth session", sessErr)
	}

	// TODO: ensure session user has access to debrief

	debrief, debriefErr := h.incidents.GetDebrief(ctx, request.Id)
	if debriefErr != nil {
		return nil, detailError("failed to get incident debrief", debriefErr)
	}
	resp.Body.Data = oapi.IncidentDebriefFromEnt(debrief)

	return &resp, nil
}

func (h *incidentDebriefsHandler) UpdateIncidentDebrief(ctx context.Context, request *oapi.UpdateIncidentDebriefRequest) (*oapi.UpdateIncidentDebriefResponse, error) {
	var resp oapi.UpdateIncidentDebriefResponse

	status := request.Body.Attributes.Status
	log.Debug().Str("status", status).Msg("update")

	var debrief *ent.IncidentDebrief
	var err error
	if status == "started" {
		debrief, err = h.incidents.StartDebrief(ctx, request.Id)
	} else if status == "completed" {
		debrief, err = h.incidents.CompleteDebrief(ctx, request.Id)
	}

	if debrief == nil || err != nil {
		return nil, detailError("update failed", err)
	}

	resp.Body.Data = oapi.IncidentDebriefFromEnt(debrief)

	return &resp, nil
}

func (h *incidentDebriefsHandler) ListIncidentDebriefMessages(ctx context.Context, request *oapi.ListIncidentDebriefMessagesRequest) (*oapi.ListIncidentDebriefMessagesResponse, error) {
	var resp oapi.ListIncidentDebriefMessagesResponse

	debrief, debriefErr := h.incidents.GetDebrief(ctx, request.Id)
	if debriefErr != nil {
		return nil, detailError("failed to get debrief", debriefErr)
	}

	msgs, msgsErr := debrief.QueryMessages().
		Order(incidentdebriefmessage.ByCreatedAt()).
		All(ctx)
	if msgsErr != nil {
		return nil, detailError("failed to query debrief messages", msgsErr)
	}

	resp.Body.Data = make([]oapi.IncidentDebriefMessage, len(msgs))
	for i, msg := range msgs {
		resp.Body.Data[i] = oapi.IncidentDebriefMessageFromEnt(msg)
	}

	return &resp, nil
}

func (h *incidentDebriefsHandler) AddIncidentDebriefUserMessage(ctx context.Context, request *oapi.AddIncidentDebriefUserMessageRequest) (*oapi.AddIncidentDebriefUserMessageResponse, error) {
	var resp oapi.AddIncidentDebriefUserMessageResponse

	msg, msgErr := h.incidents.AddUserDebriefMessage(ctx, request.Id, request.Body.Attributes.MessageContent)
	if msgErr != nil {
		return nil, detailError("failed to add user message", msgErr)
	}

	resp.Body.Data = oapi.IncidentDebriefMessageFromEnt(msg)

	return &resp, nil
}

func (h *incidentDebriefsHandler) ListIncidentDebriefSuggestions(ctx context.Context, request *oapi.ListIncidentDebriefSuggestionsRequest) (*oapi.ListIncidentDebriefSuggestionsResponse, error) {
	var resp oapi.ListIncidentDebriefSuggestionsResponse

	debrief, debriefErr := h.incidents.GetDebrief(ctx, request.Id)
	if debriefErr != nil {
		return nil, detailError("failed to get debrief", debriefErr)
	}

	sugs, sugsErr := debrief.QuerySuggestions().All(ctx)
	if sugsErr != nil {
		return nil, detailError("failed to query suggestions", sugsErr)
	}

	resp.Body.Data = make([]oapi.IncidentDebriefSuggestion, len(sugs))
	for i, sug := range sugs {
		resp.Body.Data[i] = oapi.IncidentDebriefSuggestionFromEnt(sug)
	}

	return &resp, nil
}

func (h *incidentDebriefsHandler) ListIncidentDebriefQuestions(ctx context.Context, request *oapi.ListIncidentDebriefQuestionsRequest) (*oapi.ListIncidentDebriefQuestionsResponse, error) {
	var resp oapi.ListIncidentDebriefQuestionsResponse

	query := h.questions.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}

	questions, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("Failed to query debrief questions", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentDebriefQuestion, len(questions))
	for i, q := range questions {
		resp.Body.Data[i] = oapi.IncidentDebriefQuestionFromEnt(q)
	}

	return &resp, nil
}

func (h *incidentDebriefsHandler) CreateIncidentDebriefQuestion(ctx context.Context, request *oapi.CreateIncidentDebriefQuestionRequest) (*oapi.CreateIncidentDebriefQuestionResponse, error) {
	var resp oapi.CreateIncidentDebriefQuestionResponse

	attr := request.Body.Attributes

	query := h.questions.Create().
		SetContent(attr.Content)

	question, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("Failed to create incident debrief question", createErr)
	}
	resp.Body.Data = oapi.IncidentDebriefQuestionFromEnt(question)

	return &resp, nil
}

func (h *incidentDebriefsHandler) GetIncidentDebriefQuestion(ctx context.Context, request *oapi.GetIncidentDebriefQuestionRequest) (*oapi.GetIncidentDebriefQuestionResponse, error) {
	var resp oapi.GetIncidentDebriefQuestionResponse

	question, queryErr := h.questions.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("Failed to query debrief question", queryErr)
	}

	resp.Body.Data = oapi.IncidentDebriefQuestionFromEnt(question)

	return &resp, nil
}

func (h *incidentDebriefsHandler) UpdateIncidentDebriefQuestion(ctx context.Context, request *oapi.UpdateIncidentDebriefQuestionRequest) (*oapi.UpdateIncidentDebriefQuestionResponse, error) {
	var resp oapi.UpdateIncidentDebriefQuestionResponse

	attr := request.Body.Attributes

	query := h.questions.UpdateOneID(request.Id).
		SetNillableContent(attr.Content)

	question, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("Failed to create incident debrief question", createErr)
	}
	resp.Body.Data = oapi.IncidentDebriefQuestionFromEnt(question)

	return &resp, nil
}

func (h *incidentDebriefsHandler) ArchiveIncidentDebriefQuestion(ctx context.Context, request *oapi.ArchiveIncidentDebriefQuestionRequest) (*oapi.ArchiveIncidentDebriefQuestionResponse, error) {
	var resp oapi.ArchiveIncidentDebriefQuestionResponse

	delErr := h.questions.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, detailError("Failed to archive incident debrief question", delErr)
	}

	return &resp, nil
}
