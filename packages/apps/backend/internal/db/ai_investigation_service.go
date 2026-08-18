package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentartifact"
	"github.com/rezible/rezible/ent/alertinstance"
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

func NewInvestigationService(db rez.Database, msgs rez.MessageService, alerts rez.AlertService, agents rez.AgentSessionService) (*InvestigationService, error) {
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

func (s *InvestigationService) CreateAlertInvestigation(ctx context.Context, instanceId uuid.UUID) (*ent.AlertInvestigation, error) {
	var investigation *ent.AlertInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryInstance := tx.AlertInstance.Query().
			Where(alertinstance.ID(instanceId)).
			WithAlert()
		instance, instanceErr := queryInstance.Only(ctx)
		if instanceErr != nil {
			return fmt.Errorf("get alert instance: %w", instanceErr)
		}
		alrt, alertErr := instance.Edges.AlertOrErr()
		if alertErr != nil {
			return fmt.Errorf("get alert: %w", alertErr)
		}

		createAnalysis := tx.SystemAnalysis.Create()
		if alrt.KnowledgeEntityID != nil {
			createAnalysis.SetSubjectEntityID(*alrt.KnowledgeEntityID)
		}
		analysis, analysisErr := createAnalysis.Save(ctx)
		if analysisErr != nil {
			return fmt.Errorf("create system analysis: %w", analysisErr)
		}
		if alrt.KnowledgeEntityID != nil {
			createAnalysisEntity := tx.SystemAnalysisEntity.Create().
				SetAnalysisID(analysis.ID).
				SetKnowledgeEntityID(*alrt.KnowledgeEntityID)
			if _, entityErr := createAnalysisEntity.Save(ctx); entityErr != nil {
				return fmt.Errorf("seed system analysis entity: %w", entityErr)
			}
		}

		params := rez.CreateAgentSessionParams{
			AgentName:        rezai.AlertsAgent.Name,
			OwnerUserID:      nil,
			PermissionScopes: nil,
			Input:            rezai.AlertAgentInput{AlertInstanceID: instanceId},
			SystemAnalysisID: &analysis.ID,
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

func (s *InvestigationService) onAgentTurnFinished(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	// TODO: check type of agent session? Maybe use a session binding?
	queryInvestigation := s.db.Client(ctx).AlertInvestigation.Query().
		Where(ain.AgentSessionID(ev.AgentSessionId))
	exists, invErr := queryInvestigation.Exist(ctx)
	if invErr != nil {
		return fmt.Errorf("query alert investigation: %w", invErr)
	}
	if !exists {
		return nil
	}

	queryArtifact := s.db.Client(ctx).AgentArtifact.Query().
		Where(
			agentartifact.AgentSessionID(ev.AgentSessionId),
			agentartifact.Name("investigation_report"),
		)
	artifact, artifactErr := queryArtifact.Only(ctx)
	if artifactErr != nil {
		if ent.IsNotFound(artifactErr) {
			return nil
		}
		return fmt.Errorf("query investigation report artifact: %w", artifactErr)
	}

	report, reportErr := s.alertInvestigationReportFromArtifact(artifact.Parts)
	if reportErr != nil {
		return reportErr
	}
	_, saveErr := s.SetAlertInvestigationReport(ctx, ev.AgentSessionId, report)
	return saveErr
}

func (s *InvestigationService) alertInvestigationReportFromArtifact(parts []*ai.Part) (schematypes.AlertInvestigationReport, error) {
	var report schematypes.AlertInvestigationReport
	for _, part := range parts {
		text := strings.TrimSpace(part.Text)
		if !strings.HasPrefix(text, "{") {
			continue
		}
		if jsonErr := json.Unmarshal([]byte(text), &report); jsonErr != nil {
			return report, fmt.Errorf("parse investigation report artifact: %w", jsonErr)
		}
		return report, nil
	}
	return report, fmt.Errorf("%w: investigation report artifact has no JSON part", rez.ErrInvalidInput)
}

func (s *InvestigationService) SetAlertInvestigationReport(ctx context.Context, agentSessionID uuid.UUID, report schematypes.AlertInvestigationReport) (*ent.AlertInvestigation, error) {
	report.Text = strings.TrimSpace(report.Text)
	if report.Text == "" {
		return nil, fmt.Errorf("%w: report text is empty", rez.ErrInvalidInput)
	}
	var investigation *ent.AlertInvestigation
	return investigation, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryInvestigation := tx.AlertInvestigation.Query().
			Where(ain.AgentSessionID(agentSessionID))
		curr, queryErr := queryInvestigation.Only(ctx)
		if queryErr != nil {
			return queryErr
		}
		updateInvestigation := curr.Update().
			SetReport(report)
		updated, updateErr := updateInvestigation.Save(ctx)
		if updateErr != nil {
			return updateErr
		}
		investigation = updated.Unwrap()
		return nil
	})
}
