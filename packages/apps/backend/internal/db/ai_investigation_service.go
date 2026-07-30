package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ain "github.com/rezible/rezible/ent/alertinvestigation"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/messages"
)

type InvestigationService struct {
	db     rez.Database
	alerts rez.AlertService
	agents rez.AgentSessionService
}

func NewInvestigationService(db rez.Database, msgs rez.MessageService, jobSvc rez.JobService, alerts rez.AlertService, agents rez.AgentSessionService) (*InvestigationService, error) {
	is := &InvestigationService{
		db:     db,
		alerts: alerts,
		agents: agents,
	}

	handlersErr := msgs.AddHandlers(
		messages.NewEventHandler("db.InvestigationService.onAgentTurnFinished", is.onAgentTurnFinished))
	if handlersErr != nil {
		return nil, fmt.Errorf("adding handlers: %w", handlersErr)
	}

	return is, nil
}

func (s *InvestigationService) onAgentTurnFinished(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	// TODO: check for report in artifacts
	return nil
}

func (s *InvestigationService) CreateAlertInvestigation(ctx context.Context, instanceId uuid.UUID) (*ent.AlertInvestigation, error) {
	var investigation *ent.AlertInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		params := rez.CreateAgentSessionParams{
			AgentName:        rezai.AlertsAgent.Name,
			OwnerUserID:      nil,
			PermissionScopes: nil,
			Input:            rezai.AlertAgentInput{AlertInstanceID: instanceId},
			Metadata:         nil,
		}
		sess, sessErr := s.agents.CreateAgentSession(ctx, params)
		if sessErr != nil {
			return sessErr
		}

		createInvestigation := tx.AlertInvestigation.Create().
			SetAlertInstanceID(instanceId).
			SetAgentSessionID(sess.ID)
		created, invErr := createInvestigation.Save(ctx)
		if invErr != nil {
			return fmt.Errorf("failed to create alert investigation: %w", invErr)
		}
		investigation = created.Unwrap()
		return nil
	})
}

func (s *InvestigationService) SetAlertInvestigationReport(ctx context.Context, agentSessionID uuid.UUID, report schematypes.AlertInvestigationReport) (*ent.AlertInvestigation, error) {
	report.Text = strings.TrimSpace(report.Text)
	if report.Text == "" {
		return nil, fmt.Errorf("%w: report text is empty", rez.ErrInvalidInput)
	}
	var investigation *ent.AlertInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		curr, queryErr := tx.AlertInvestigation.Query().Where(ain.AgentSessionID(agentSessionID)).Only(ctx)
		if queryErr != nil {
			return queryErr
		}
		updated, updateErr := curr.Update().SetReport(report).Save(ctx)
		if updateErr != nil {
			return updateErr
		}
		investigation = updated.Unwrap()
		return nil
	})
}
