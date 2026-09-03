package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	ocsp "github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"github.com/rezible/rezible/ent/team"
)

type OncallRostersService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewOncallRostersService(db rez.Database, jobSvc rez.JobService) (*OncallRostersService, error) {
	s := &OncallRostersService{
		db:   db,
		jobs: jobSvc,
	}

	return s, nil
}

func (s *OncallRostersService) GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error) {
	roster, rosterErr := s.db.Client(ctx).OncallRoster.Get(ctx, id)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallRostersService) GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.Client(ctx).OncallSchedule.Query().
		Where(oncallschedule.ID(scheduleId)).
		QueryRoster().
		Only(ctx)
}

func (s *OncallRostersService) GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error) {
	query := s.db.Client(ctx).OncallRoster.Query().
		Where(oncallroster.Slug(slug))

	roster, rosterErr := query.Only(ctx)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallRostersService) ListRosters(ctx context.Context, params rez.ListOncallRostersParams) (*ent.ListResult[ent.OncallRoster], error) {
	query := s.db.Client(ctx).OncallRoster.Query().
		Order(oncallroster.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(oncallroster.NameContainsFold(search))
	}
	if params.TeamID != uuid.Nil {
		query.Where(oncallroster.HasTeamsWith(team.ID(params.TeamID)))
	}

	if params.UserID != uuid.Nil {
		schedList, schedulesErr := s.db.Client(ctx).OncallScheduleParticipant.Query().
			Where(ocsp.UserID(params.UserID)).
			QuerySchedule().
			All(ctx)
		if schedulesErr != nil {
			return nil, fmt.Errorf("failed to list oncall schedules: %w", schedulesErr)
		}
		var rosterIds []uuid.UUID
		for _, sched := range schedList {
			rosterIds = append(rosterIds, sched.RosterID)
		}
		query = query.Where(oncallroster.IDIn(rosterIds...))
	}

	return ent.DoListQuery[ent.OncallRoster, *ent.OncallRosterQuery](ctx, query, params.ListParams)
}

func (s *OncallRostersService) ListSchedules(ctx context.Context, params rez.ListOncallSchedulesParams) (*ent.ListResult[ent.OncallSchedule], error) {
	var query *ent.OncallScheduleQuery
	if params.UserID != uuid.Nil {
		query = s.db.Client(ctx).OncallScheduleParticipant.Query().
			Where(ocsp.UserID(params.UserID)).
			QuerySchedule()
	} else {
		query = s.db.Client(ctx).OncallSchedule.Query()
	}
	query.Order(oncallschedule.ByID(params.GetOrder()))

	return ent.DoListQuery[ent.OncallSchedule, *ent.OncallScheduleQuery](ctx, query, params.ListParams)
}
