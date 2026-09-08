package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	sh "github.com/rezible/rezible/ent/systemhazard"
	shra "github.com/rezible/rezible/ent/systemhazardriskassessment"
	"github.com/rezible/rezible/pkg/projections"
)

type SystemHazardService struct {
	db        rez.Database
	knowledge rez.KnowledgeGraphService
}

func NewSystemHazardService(db rez.Database, knowledge rez.KnowledgeGraphService) (*SystemHazardService, error) {
	return &SystemHazardService{db: db, knowledge: knowledge}, nil
}

func (s *SystemHazardService) CreateSystemHazard(ctx context.Context, params rez.CreateSystemHazardParams) (*ent.SystemHazard, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: system hazard title is required", rez.ErrInvalidInput)
	}

	var created *ent.SystemHazard
	return created, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		hazardId := uuid.New()
		ka, kaErr := s.knowledge.ResolveInternalEntity(ctx, rez.KnowledgeEntityRef{
			Category:            kne.CategoryConcern,
			Kind:                "system_hazard",
			ProviderResourceRef: projections.InternalEntityResourceRef(hazardId),
		})
		if kaErr != nil || ka.EntityID == nil {
			return fmt.Errorf("create situation knowledge entity: %w", kaErr)
		}

		createHazard := tx.SystemHazard.Create().
			SetID(hazardId).
			SetKnowledgeEntityID(*ka.EntityID).
			SetTitle(title).
			SetStatus(sh.StatusActive)
		if description := strings.TrimSpace(params.Description); description != "" {
			createHazard.SetDescription(description)
		}
		if consequences := strings.TrimSpace(params.PotentialConsequences); consequences != "" {
			createHazard.SetPotentialConsequences(consequences)
		}
		createdHazard, saveErr := createHazard.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create system hazard: %w", saveErr)
		}
		created = createdHazard.Unwrap()
		return nil
	})
}

func (s *SystemHazardService) GetSystemHazard(ctx context.Context, id uuid.UUID) (*ent.SystemHazard, error) {
	return s.db.Client(ctx).SystemHazard.Query().
		Where(sh.ID(id)).
		WithRiskAssessments(func(query *ent.SystemHazardRiskAssessmentQuery) {
			query.Order(shra.ByRevision(sql.OrderAsc()))
		}).
		Only(ctx)
}

const systemHazardLockNamespace = "system_hazard"

func (s *SystemHazardService) RetireSystemHazard(ctx context.Context, id uuid.UUID) (*ent.SystemHazard, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: system hazard id is required", rez.ErrInvalidInput)
	}

	var retired *ent.SystemHazard
	return retired, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, systemHazardLockNamespace, id.String()); lockErr != nil {
			return fmt.Errorf("lock system hazard: %w", lockErr)
		}

		current, queryErr := tx.SystemHazard.Get(ctx, id)
		if queryErr != nil {
			return fmt.Errorf("get system hazard: %w", queryErr)
		}
		if current.Status == sh.StatusRetired {
			retired = current.Unwrap()
			return nil
		}
		update := current.Update().
			SetStatus(sh.StatusRetired)
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("retire system hazard: %w", updateErr)
		}

		retired = updated.Unwrap()
		return nil
	})
}

func (s *SystemHazardService) AddSystemHazardRiskAssessment(ctx context.Context, params rez.AddSystemHazardRiskAssessmentParams) (*ent.SystemHazardRiskAssessment, error) {
	if params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: system hazard id is required", rez.ErrInvalidInput)
	}
	likelihood := strings.TrimSpace(params.Likelihood)
	consequence := strings.TrimSpace(params.Consequence)
	riskLevel := strings.TrimSpace(params.RiskLevel)
	if likelihood == "" || consequence == "" || riskLevel == "" {
		return nil, fmt.Errorf("%w: likelihood, consequence, and risk level are required", rez.ErrInvalidInput)
	}
	assessedAt := params.AssessedAt
	if assessedAt.IsZero() {
		assessedAt = time.Now().UTC()
	}

	var assessment *ent.SystemHazardRiskAssessment
	return assessment, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, systemHazardLockNamespace, params.SystemHazardID.String()); lockErr != nil {
			return fmt.Errorf("lock system hazard: %w", lockErr)
		}
		if _, hazardErr := tx.SystemHazard.Get(ctx, params.SystemHazardID); hazardErr != nil {
			return fmt.Errorf("get system hazard: %w", hazardErr)
		}

		lookupByRevision := tx.SystemHazardRiskAssessment.Query().
			Where(shra.SystemHazardID(params.SystemHazardID)).
			Order(shra.ByRevision(sql.OrderDesc()))
		latest, latestErr := lookupByRevision.First(ctx)
		nextRevision := 1
		if latestErr == nil {
			nextRevision = latest.Revision + 1
		} else if !ent.IsNotFound(latestErr) {
			return fmt.Errorf("get latest system hazard risk assessment: %w", latestErr)
		}

		createAssessment := tx.SystemHazardRiskAssessment.Create().
			SetSystemHazardID(params.SystemHazardID).
			SetRevision(nextRevision).
			SetLikelihood(likelihood).
			SetConsequence(consequence).
			SetRiskLevel(riskLevel).
			SetAssessedAt(assessedAt)
		if rationale := strings.TrimSpace(params.Rationale); rationale != "" {
			createAssessment.SetRationale(rationale)
		}
		created, saveErr := createAssessment.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("add system hazard risk assessment: %w", saveErr)
		}
		assessment = created.Unwrap()
		return nil
	})
}

func (s *SystemHazardService) ListSystemHazardRiskAssessments(ctx context.Context, params rez.ListSystemHazardRiskAssessmentsParams) (*ent.ListResult[ent.SystemHazardRiskAssessment], error) {
	query := s.db.Client(ctx).SystemHazardRiskAssessment.Query().
		Order(shra.ByRevision(params.GetOrder()), shra.ByID(params.GetOrder()))
	if params.SystemHazardID != uuid.Nil {
		query.Where(shra.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SystemHazardRiskAssessment, *ent.SystemHazardRiskAssessmentQuery](ctx, query, params.ListParams)
}

func (s *SystemHazardService) GetLatestSystemHazardRiskAssessment(ctx context.Context, systemHazardID uuid.UUID) (*ent.SystemHazardRiskAssessment, error) {
	return s.db.Client(ctx).SystemHazardRiskAssessment.Query().
		Where(shra.SystemHazardID(systemHazardID)).
		Order(shra.ByRevision(sql.OrderDesc())).
		WithSystemHazard().
		First(ctx)
}
