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
	"github.com/rezible/rezible/ent/systemhazard"
	"github.com/rezible/rezible/ent/systemhazardriskassessment"
)

type SystemHazardService struct {
	db rez.Database
}

func NewSystemHazardService(db rez.Database) (*SystemHazardService, error) {
	return &SystemHazardService{db: db}, nil
}

func (s *SystemHazardService) CreateSystemHazard(ctx context.Context, params rez.CreateSystemHazardParams) (*ent.SystemHazard, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: system hazard title is required", rez.ErrInvalidInput)
	}

	createHazard := s.db.Client(ctx).SystemHazard.Create().
		SetTitle(title).
		SetStatus(systemhazard.StatusActive)
	if description := strings.TrimSpace(params.Description); description != "" {
		createHazard.SetDescription(description)
	}
	if consequences := strings.TrimSpace(params.PotentialConsequences); consequences != "" {
		createHazard.SetPotentialConsequences(consequences)
	}
	return createHazard.Save(ctx)
}

func (s *SystemHazardService) GetSystemHazard(ctx context.Context, id uuid.UUID) (*ent.SystemHazard, error) {
	return s.db.Client(ctx).SystemHazard.Query().
		Where(systemhazard.ID(id)).
		WithRiskAssessments(func(query *ent.SystemHazardRiskAssessmentQuery) {
			query.Order(systemhazardriskassessment.ByRevision(sql.OrderAsc()))
		}).
		Only(ctx)
}

const systemHazardLockNamespace = "system_hazard"

func (s *SystemHazardService) RetireSystemHazard(ctx context.Context, params rez.RetireSystemHazardParams) (*ent.SystemHazard, error) {
	if params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: system hazard id is required", rez.ErrInvalidInput)
	}

	var retired *ent.SystemHazard
	return retired, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, systemHazardLockNamespace, params.SystemHazardID.String()); lockErr != nil {
			return fmt.Errorf("lock system hazard: %w", lockErr)
		}
		current, queryErr := tx.SystemHazard.Query().Where(systemhazard.ID(params.SystemHazardID)).Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("get system hazard: %w", queryErr)
		}
		if current.Status == systemhazard.StatusRetired {
			retired = current.Unwrap()
			return nil
		}
		updated, updateErr := tx.SystemHazard.UpdateOneID(params.SystemHazardID).
			SetStatus(systemhazard.StatusRetired).
			Save(ctx)
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

		nextRevision := 1
		latest, latestErr := tx.SystemHazardRiskAssessment.Query().
			Where(systemhazardriskassessment.SystemHazardID(params.SystemHazardID)).
			Order(systemhazardriskassessment.ByRevision(sql.OrderDesc())).
			First(ctx)
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
		Order(
			systemhazardriskassessment.ByRevision(params.GetOrder()),
			systemhazardriskassessment.ByID(params.GetOrder()),
		)
	if params.SystemHazardID != uuid.Nil {
		query.Where(systemhazardriskassessment.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SystemHazardRiskAssessment, *ent.SystemHazardRiskAssessmentQuery](ctx, query, params.ListParams)
}

func (s *SystemHazardService) GetLatestSystemHazardRiskAssessment(ctx context.Context, systemHazardID uuid.UUID) (*ent.SystemHazardRiskAssessment, error) {
	return s.db.Client(ctx).SystemHazardRiskAssessment.Query().
		Where(systemhazardriskassessment.SystemHazardID(systemHazardID)).
		Order(systemhazardriskassessment.ByRevision(sql.OrderDesc())).
		WithSystemHazard().
		First(ctx)
}
