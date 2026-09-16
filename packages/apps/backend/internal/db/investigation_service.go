package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/investigation"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type InvestigationService struct {
	db     rez.Database
	agents rez.AgentSessionService
}

func NewInvestigationService(database rez.Database, agents rez.AgentSessionService) *InvestigationService {
	return &InvestigationService{db: database, agents: agents}
}

func (s *InvestigationService) CreateInvestigation(ctx context.Context, params rez.CreateInvestigationParams) (*ent.Investigation, error) {
	input := rezai.InvestigationAgentSessionInput{}
	if params.Query != nil {
		input.Query = params.Query
	}
	var result *ent.Investigation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		createAnalysis := tx.SystemAnalysis.Create()
		if params.SubjectEntityID != nil {
			createAnalysis.SetSubjectEntityID(*params.SubjectEntityID)
		}
		analysis, analysisErr := createAnalysis.Save(ctx)
		if analysisErr != nil {
			return fmt.Errorf("create system analysis: %w", analysisErr)
		}
		if params.SubjectEntityID != nil {
			createAnalysisEntity := tx.SystemAnalysisEntity.Create().SetAnalysisID(analysis.ID).SetKnowledgeEntityID(*params.SubjectEntityID)
			if entityErr := createAnalysisEntity.Exec(ctx); entityErr != nil {
				return fmt.Errorf("seed system analysis entity: %w", entityErr)
			}
		}
		session, sessionErr := s.agents.CreateAgentSession(ctx, rez.CreateAgentSessionParams{AgentName: rezai.InvestigationAgent.Name, Input: input})
		if sessionErr != nil {
			return fmt.Errorf("create investigation agent session: %w", sessionErr)
		}
		createInvestigation := tx.Investigation.Create().SetSystemAnalysisID(analysis.ID).SetAgentSessionID(session.ID)
		created, saveErr := createInvestigation.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create investigation: %w", saveErr)
		}
		result = created.Unwrap()
		return nil
	})
}

func (s *InvestigationService) GetInvestigation(ctx context.Context, id uuid.UUID) (*ent.Investigation, error) {
	return s.LookupInvestigation(ctx, investigation.ID(id))
}

func (s *InvestigationService) LookupInvestigation(ctx context.Context, predicates ...predicate.Investigation) (*ent.Investigation, error) {
	return s.db.Client(ctx).Investigation.Query().
		Where(predicates...).
		WithSystemAnalysis().
		WithAgentSession().
		WithSituations().
		WithReport().
		Only(ctx)
}
