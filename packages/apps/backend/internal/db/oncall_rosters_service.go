package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	ocsp "github.com/rezible/rezible/ent/oncallscheduleparticipant"
)

type OncallRostersService struct {
	db   *ent.Client
	jobs rez.JobsService
}

func NewOncallRostersService(db *ent.Client, jobs rez.JobsService) (*OncallRostersService, error) {
	s := &OncallRostersService{
		db:   db,
		jobs: jobs,
	}

	return s, nil
}

func (s *OncallRostersService) GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error) {
	roster, rosterErr := s.db.OncallRoster.Get(ctx, id)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallRostersService) GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.OncallSchedule.Query().
		Where(oncallschedule.ID(scheduleId)).
		QueryRoster().
		Only(ctx)
}

func (s *OncallRostersService) GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error) {
	query := s.db.OncallRoster.Query().
		Where(oncallroster.Slug(slug))

	roster, rosterErr := query.Only(ctx)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallRostersService) ListRosters(ctx context.Context, params rez.ListOncallRostersParams) (*ent.ListResult[*ent.OncallRoster], error) {
	query := s.db.OncallRoster.Query()

	if params.UserID != uuid.Nil {
		schedList, schedulesErr := s.ListSchedules(ctx, rez.ListOncallSchedulesParams{
			UserID: params.UserID,
		})
		if schedulesErr != nil {
			return nil, fmt.Errorf("failed to list oncall schedules: %w", schedulesErr)
		}
		var rosterIds []uuid.UUID
		for _, sched := range schedList.Data {
			rosterIds = append(rosterIds, sched.RosterID)
		}
		query = query.Where(oncallroster.IDIn(rosterIds...))
	}

	return ent.DoListQuery[*ent.OncallRoster, *ent.OncallRosterQuery](ctx, query, params.ListParams)
}

func (s *OncallRostersService) ListSchedules(ctx context.Context, params rez.ListOncallSchedulesParams) (*ent.ListResult[*ent.OncallSchedule], error) {
	var query *ent.OncallScheduleQuery
	if params.UserID != uuid.Nil {
		query = s.db.OncallScheduleParticipant.Query().
			Where(ocsp.UserID(params.UserID)).
			QuerySchedule()
	} else {
		query = s.db.OncallSchedule.Query()
	}

	return ent.DoListQuery[*ent.OncallSchedule, *ent.OncallScheduleQuery](ctx, query, params.ListParams)
}

func (s *OncallRostersService) GetCurrentOncallForComponent(ctx context.Context, componentID uuid.UUID) ([]*ent.User, error) {
	// TODO: Implement logic to get current oncall users for a component's linked roster(s)
	return []*ent.User{}, nil
}
