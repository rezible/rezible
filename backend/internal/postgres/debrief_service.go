package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentdebrief"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/jobs"
)

type DebriefService struct {
	db   *ent.Client
	jobs rez.JobsService
	ai   rez.AiService
	chat rez.ChatService
}

func NewDebriefService(db *ent.Client, jobs rez.JobsService, ai rez.AiService, chat rez.ChatService) (*DebriefService, error) {
	svc := &DebriefService{
		db:   db,
		jobs: jobs,
		ai:   ai,
		chat: chat,
	}

	return svc, nil
}

func (s *DebriefService) CreateDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error) {
	isRequired, reqErr := s.isUserDebriefRequired(ctx, userId, incidentId)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to check if debrief is required: %w", reqErr)
	}

	return s.createDebrief(ctx, incidentId, userId, isRequired)
}

func (s *DebriefService) createDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID, required bool) (*ent.IncidentDebrief, error) {
	return s.db.IncidentDebrief.Create().
		SetIncidentID(incidentId).
		SetUserID(userId).
		SetRequired(required).
		SetStarted(false).
		Save(ctx)
}

func (s *DebriefService) isUserDebriefRequired(ctx context.Context, userId, incidentId uuid.UUID) (bool, error) {
	hadRole, roleErr := s.db.IncidentRoleAssignment.Query().
		Where(incidentroleassignment.UserID(userId)).
		Where(incidentroleassignment.IncidentID(incidentId)).
		Exist(ctx)
	if roleErr != nil {
		return false, fmt.Errorf("failed to query incident role: %w", roleErr)
	}
	if hadRole {
		return true, nil
	}

	// TODO: query settings?

	return false, nil
}

func (s *DebriefService) StartDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error) {
	debrief, getErr := s.GetDebrief(ctx, debriefId)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get debrief: %w", getErr)
	}
	if debrief.Started {
		return debrief, nil
	}

	updateTxFn := func(tx *ent.Tx) error {
		updated, updateErr := tx.IncidentDebrief.UpdateOneID(debriefId).
			SetStarted(true).
			Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("failed to start incident debrief: %w", updateErr)
		}

		job := jobs.GenerateIncidentDebriefResponse{DebriefId: debriefId}
		if genErr := s.jobs.InsertTx(ctx, tx, job, nil); genErr != nil {
			return fmt.Errorf("failed to request response generation: %w", genErr)
		}

		debrief = updated
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, updateTxFn); txErr != nil {
		return nil, fmt.Errorf("failed to start debrief: %w", txErr)
	}

	return debrief, nil
}

func (s *DebriefService) CompleteDebrief(ctx context.Context, debriefId uuid.UUID) (*ent.IncidentDebrief, error) {
	debrief, getErr := s.GetDebrief(ctx, debriefId)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get debrief: %w", getErr)
	}
	//if debrief.Completed {
	//	return debrief, nil
	//}

	updateTxFn := func(tx *ent.Tx) error {
		updated, updateErr := tx.IncidentDebrief.UpdateOneID(debriefId).
			SetStarted(true).
			Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("failed to save: %w", updateErr)
		}

		job := jobs.GenerateIncidentDebriefSuggestions{DebriefId: debriefId}
		if genErr := s.jobs.InsertTx(ctx, tx, job, nil); genErr != nil {
			return fmt.Errorf("failed to request suggestions generation: %w", genErr)
		}

		debrief = updated
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, updateTxFn); txErr != nil {
		return nil, fmt.Errorf("failed to start debrief: %w", txErr)
	}

	return debrief, nil
}

func (s *DebriefService) GetDebrief(ctx context.Context, id uuid.UUID) (*ent.IncidentDebrief, error) {
	return s.db.IncidentDebrief.Get(ctx, id)
}

func (s *DebriefService) GetUserDebrief(ctx context.Context, incidentId uuid.UUID, userId uuid.UUID) (*ent.IncidentDebrief, error) {
	return s.db.IncidentDebrief.Query().
		Where(incidentdebrief.And(incidentdebrief.IncidentID(incidentId), incidentdebrief.UserID(userId))).
		Only(ctx)
}

func (s *DebriefService) SendUserDebriefRequests(ctx context.Context, incidentId uuid.UUID) error {
	inc, incErr := s.db.Incident.Get(ctx, incidentId)
	if incErr != nil {
		return fmt.Errorf("get incident %s failed: %w", incidentId.String(), incErr)
	}

	debriefs, debriefsErr := s.db.IncidentDebrief.Query().
		Where(incidentdebrief.IncidentID(incidentId)).
		All(ctx)
	if debriefsErr != nil && !ent.IsNotFound(debriefsErr) {
		return fmt.Errorf("failed to query existing debriefs: %w", debriefsErr)
	}

	alreadyRequestedIds := make([]uuid.UUID, len(debriefs))
	for i, debrief := range debriefs {
		alreadyRequestedIds[i] = debrief.UserID
	}

	assnQuery := s.db.IncidentRoleAssignment.Query().
		Where(incidentroleassignment.IncidentID(incidentId)).
		Where(incidentroleassignment.UserIDNotIn(alreadyRequestedIds...)).
		WithUser()

	assignments, assnErr := assnQuery.All(ctx)
	if assnErr != nil {
		return fmt.Errorf("failed to query role assignments: %w", assnErr)
	}

	incidentUsers := make(map[uuid.UUID]*ent.User)
	for _, assn := range assignments {
		incidentUsers[assn.UserID] = assn.Edges.User
	}

	for _, user := range incidentUsers {
		s.prepareUserDebrief(ctx, user, inc)
	}

	return nil
}

func (s *DebriefService) prepareUserDebrief(ctx context.Context, user *ent.User, inc *ent.Incident) {
	_, createErr := s.createDebrief(ctx, inc.ID, user.ID, true)
	if createErr != nil {
		log.Error().Err(createErr).Msg("Failed to create incident debrief")
	}

	incFmt := inc.Slug
	if inc.ChatChannelID != "" {
		incFmt = fmt.Sprintf("<#%s>", inc.ChatChannelID)
	}

	msgText := fmt.Sprintf(
		"Thank you for your role in %s!\nPlease complete your incident debrief as soon as possible",
		incFmt)
	msgLinkUrl := fmt.Sprintf("%s/incidents/%s/retrospective", rez.FrontendUrl, inc.ID.String())
	msgLinkText := "Open Incident Debrief"
	if msgErr := s.chat.SendUserLinkMessage(ctx, user, msgText, msgLinkUrl, msgLinkText); msgErr != nil {
		log.Error().Err(msgErr).Msg("Failed to send incident debrief message")
	}
}

func (s *DebriefService) AddUserDebriefMessage(ctx context.Context, debriefId uuid.UUID, content string) (*ent.IncidentDebriefMessage, error) {
	debrief, getErr := s.db.IncidentDebrief.Get(ctx, debriefId)
	if getErr != nil {
		return nil, getErr
	}

	var msg *ent.IncidentDebriefMessage
	addMessageTx := func(tx *ent.Tx) error {
		created, msgErr := tx.IncidentDebriefMessage.Create().
			SetDebriefID(debrief.ID).
			SetType(incidentdebriefmessage.TypeUser).
			SetBody(content).
			Save(ctx)
		if msgErr != nil {
			return fmt.Errorf("failed to save incident debrief message: %w", msgErr)
		}

		job := jobs.GenerateIncidentDebriefResponse{DebriefId: debriefId}
		if genJobErr := s.jobs.InsertTx(ctx, tx, job, nil); genJobErr != nil {
			return fmt.Errorf("failed to request response generation: %w", genJobErr)
		}

		msg = created
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, addMessageTx); txErr != nil {
		return nil, fmt.Errorf("failed to add user debrief message: %w", txErr)
	}
	return msg, nil
}

func (s *DebriefService) GenerateResponse(ctx context.Context, debriefId uuid.UUID) error {
	debrief, debriefErr := s.db.IncidentDebrief.Query().
		Where(incidentdebrief.ID(debriefId)).
		WithMessages().
		Only(ctx)
	if debriefErr != nil {
		log.Warn().Str("id", debriefId.String()).Msg("get debrief")
		return fmt.Errorf("failed to get debrief: %w", debriefErr)
	}

	question, nextErr := s.getFirstUnusedDebriefQuestion(ctx, debrief)
	if nextErr != nil {
		return fmt.Errorf("failed to get next unused question: %w", nextErr)
	}

	var questionId *uuid.UUID
	var msg *ent.IncidentDebriefMessage
	if question != nil {
		questionMsg, formatErr := s.formatDebriefQuestionMessage(ctx, question, debrief)
		if formatErr != nil {
			return fmt.Errorf("failed to format debrief question: %w", formatErr)
		}
		msg = questionMsg
		questionId = &question.ID
	} else {
		assistantMsg, responseErr := s.ai.GenerateDebriefResponse(ctx, debrief)
		if responseErr != nil {
			return fmt.Errorf("failed to generate debrief message: %w", responseErr)
		}
		msg = assistantMsg
	}

	create := s.db.IncidentDebriefMessage.Create().
		SetDebriefID(debrief.ID).
		SetType(msg.Type).
		SetBody(msg.Body).
		SetNillableFromQuestionID(questionId)
	if createErr := create.Exec(ctx); createErr != nil {
		return fmt.Errorf("failed to create debrief message: %w", createErr)
	}

	return nil
}

func (s *DebriefService) getFirstUnusedDebriefQuestion(ctx context.Context, debrief *ent.IncidentDebrief) (*ent.IncidentDebriefQuestion, error) {
	messages, msgErr := debrief.Edges.MessagesOrErr()
	if msgErr != nil {
		return nil, fmt.Errorf("debrief messages not loaded: %w", msgErr)
	}

	questions, questionsErr := s.getApplicableQuestionsForDebrief(ctx, debrief)
	if questionsErr != nil {
		return nil, fmt.Errorf("failed to get applicable questions: %w", questionsErr)
	}

	for _, question := range questions {
		seen := false
		for _, msg := range messages {
			if msg.QuestionID == question.ID {
				seen = true
				break
			}
		}
		if !seen {
			return question, nil
		}
	}

	return nil, nil
}

func (s *DebriefService) formatDebriefQuestionMessage(ctx context.Context, q *ent.IncidentDebriefQuestion, debrief *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error) {
	// TODO: allow adding incident content into question
	return &ent.IncidentDebriefMessage{
		QuestionID: q.ID,
		Type:       incidentdebriefmessage.TypeQuestion,
		Body:       q.Content,
	}, nil
}

func (s *DebriefService) getApplicableQuestionsForDebrief(ctx context.Context, debrief *ent.IncidentDebrief) ([]*ent.IncidentDebriefQuestion, error) {
	// TODO: cache this
	var debriefQuestions []*ent.IncidentDebriefQuestion

	questions, qErr := s.db.IncidentDebriefQuestion.Query().
		WithIncidentFields().
		WithIncidentRoles().
		WithIncidentSeverities().
		WithIncidentTags().
		WithIncidentTypes().
		All(ctx)
	if qErr != nil {
		return nil, fmt.Errorf("failed to get debrief questions: %w", qErr)
	}

	if len(questions) == 0 {
		return debriefQuestions, nil
	}

	inc, incErr := debrief.QueryIncident().
		WithFieldSelections().
		WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.Where(incidentroleassignment.UserID(debrief.UserID))
		}).
		WithSeverity().
		WithTagAssignments().
		WithType().
		Only(ctx)
	if incErr != nil {
		return nil, fmt.Errorf("failed to get incident: %w", incErr)
	}

	questionMatchesIncident := makeDebriefQuestionMatcher(inc)
	for _, q := range questions {
		if questionMatchesIncident(q) {
			debriefQuestions = append(debriefQuestions, q)
		}
	}

	return debriefQuestions, nil
}

func makeDebriefQuestionMatcher(inc *ent.Incident) func(question *ent.IncidentDebriefQuestion) bool {
	fields := make(map[uuid.UUID]bool)
	for _, f := range inc.Edges.FieldSelections {
		fields[f.IncidentFieldID] = true
	}
	fieldsMatchIncident := func(questionFields []*ent.IncidentField) bool {
		for _, f := range questionFields {
			if _, ok := fields[f.ID]; ok {
				return true
			}
		}
		return false
	}

	roles := make(map[uuid.UUID]bool)
	for _, r := range inc.Edges.RoleAssignments {
		roles[r.RoleID] = true
	}
	rolesMatchIncident := func(questionRoles []*ent.IncidentRole) bool {
		for _, r := range questionRoles {
			if _, ok := roles[r.ID]; ok {
				return true
			}
		}
		return false
	}

	tags := make(map[uuid.UUID]bool)
	for _, t := range inc.Edges.TagAssignments {
		tags[t.ID] = true
	}
	tagsMatchIncident := func(questionTags []*ent.IncidentTag) bool {
		for _, t := range questionTags {
			if _, ok := roles[t.ID]; ok {
				return true
			}
		}
		return false
	}

	severitiesMatchIncident := func(questionSevs []*ent.IncidentSeverity) bool {
		for _, s := range questionSevs {
			if s.ID == inc.Edges.Severity.ID {
				return true
			}
		}
		return false
	}
	return func(question *ent.IncidentDebriefQuestion) bool {
		edges := question.Edges
		if len(edges.IncidentSeverities) > 0 && !severitiesMatchIncident(edges.IncidentSeverities) {
			return false
		}
		if len(edges.IncidentTags) > 0 && !tagsMatchIncident(edges.IncidentTags) {
			return false
		}
		if len(edges.IncidentRoles) > 0 && !rolesMatchIncident(edges.IncidentRoles) {
			return false
		}
		if len(edges.IncidentFields) > 0 && !fieldsMatchIncident(edges.IncidentFields) {
			return false
		}
		return true
	}
}
