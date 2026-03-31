package db

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallshiftmetrics"
	"github.com/rezible/rezible/jobs"
)

type OncallMetricsService struct {
	db     *ent.Client
	jobs   rez.JobsService
	shifts rez.OncallShiftsService
}

func NewOncallMetricsService(db *ent.Client, jobSvc rez.JobsService, shifts rez.OncallShiftsService) (*OncallMetricsService, error) {
	s := &OncallMetricsService{
		db:     db,
		jobs:   jobSvc,
		shifts: shifts,
	}

	jobs.RegisterWorkerFunc(s.handleGenerateShiftMetrics)

	return s, nil
}

func (s *OncallMetricsService) handleGenerateShiftMetrics(ctx context.Context, args jobs.GenerateShiftMetrics) error {
	_, genErr := s.generateMetricsForShift(ctx, args.ShiftId)
	return genErr
}

func (s *OncallMetricsService) queryShiftMetrics(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftMetrics, error) {
	return s.db.OncallShiftMetrics.Query().Where(oncallshiftmetrics.ShiftID(shiftId)).Only(ctx)
}

func (s *OncallMetricsService) getShiftIncidents(ctx context.Context, shift *ent.OncallShift) ([]*ent.Incident, error) {
	return nil, nil
}

func (s *OncallMetricsService) upsertShiftMetrics(ctx context.Context, m *ent.OncallShiftMetrics) (*ent.OncallShiftMetrics, error) {
	create := s.db.OncallShiftMetrics.Create().
		SetShiftID(m.ShiftID).
		SetBurdenScore(m.BurdenScore).
		SetEventFrequency(m.EventFrequency).
		SetLifeImpact(m.LifeImpact).
		SetTimeImpact(m.TimeImpact).
		SetResponseRequirements(m.ResponseRequirements).
		SetIsolation(m.Isolation).
		SetEventsTotal(m.EventsTotal).
		SetIncidentsTotal(m.IncidentsTotal).
		SetIncidentResponseTime(m.IncidentResponseTime).
		SetAlertsTotal(m.AlertsTotal).
		SetInterruptsTotal(m.InterruptsTotal).
		SetInterruptsBusinessHours(m.InterruptsBusinessHours).
		SetInterruptsNight(m.InterruptsNight)

	upsert := create.OnConflict(sql.ConflictColumns(oncallshiftmetrics.ShiftColumn)).
		UpdateNewValues()

	metricsId, upsertErr := upsert.ID(ctx)
	if upsertErr != nil {
		return nil, fmt.Errorf("create or update shift metrics: %w", upsertErr)
	}
	m.ID = metricsId
	return m, nil
}

func (s *OncallMetricsService) generateMetricsForShift(ctx context.Context, id uuid.UUID) (*ent.OncallShiftMetrics, error) {
	shift, shiftErr := s.shifts.GetShiftByID(ctx, id)
	if shiftErr != nil {
		return nil, shiftErr
	}

	incidents, incErr := s.getShiftIncidents(ctx, shift)
	if incErr != nil {
		return nil, fmt.Errorf("shift incidents: %w", incErr)
	}

	m := &ent.OncallShiftMetrics{
		ShiftID: id,

		BurdenScore:          0,
		EventFrequency:       0,
		LifeImpact:           0,
		TimeImpact:           0,
		ResponseRequirements: 0,
		Isolation:            0,

		EventsTotal:          0,
		IncidentsTotal:       float32(len(incidents)),
		IncidentResponseTime: 0,

		AlertsTotal:             0,
		InterruptsTotal:         0,
		InterruptsNight:         0,
		InterruptsBusinessHours: 0,
	}
	return s.upsertShiftMetrics(ctx, m)
}

func (s *OncallMetricsService) GetShiftMetrics(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftMetrics, error) {
	metrics, metErr := s.queryShiftMetrics(ctx, shiftId)
	if metErr == nil {
		return metrics, nil
	}
	if !ent.IsNotFound(metErr) {
		return nil, fmt.Errorf("querying metrics: %w", metErr)
	}
	generated, genErr := s.generateMetricsForShift(ctx, shiftId)
	if genErr != nil {
		return nil, fmt.Errorf("generating metrics: %w", genErr)
	}
	return generated, nil
}

func (s *OncallMetricsService) GetComparisonShiftMetrics(ctx context.Context, from, to time.Time) (*ent.OncallShiftMetrics, error) {
	// TODO
	return &ent.OncallShiftMetrics{
		BurdenScore:          5.9,
		EventFrequency:       4.3,
		LifeImpact:           4.5,
		TimeImpact:           4.2,
		ResponseRequirements: 3.0,
		Isolation:            3.4,

		IncidentsTotal:       1.1,
		IncidentResponseTime: 33,

		InterruptsTotal:         19,
		AlertsTotal:             15,
		InterruptsNight:         4,
		InterruptsBusinessHours: 8,
	}, nil
}
