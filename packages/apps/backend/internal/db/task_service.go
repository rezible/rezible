package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/task"
)

type TaskService struct {
	db rez.Database
}

func NewTaskService(db rez.Database) (*TaskService, error) {
	return &TaskService{db: db}, nil
}

func (s *TaskService) ListTasks(ctx context.Context, params rez.ListTasksParams) (*ent.ListResult[ent.Task], error) {
	if params.State != "" {
		return nil, fmt.Errorf("%w: unsupported task state", rez.ErrUnprocessableInput)
	}
	query := s.db.Client(ctx).Task.Query()
	if params.IncidentID != uuid.Nil {
		query.Where(task.IncidentID(params.IncidentID))
	}
	if params.OwnerID != uuid.Nil {
		query.Where(task.AssigneeID(params.OwnerID))
	}
	if params.SourceEntryID != uuid.Nil {
		query.Where(task.OriginEntryID(params.SourceEntryID))
	}
	if params.State != "" {
		query.Where(task.StateEQ(params.State))
	}
	query.WithTickets().WithOriginEntry().Order(task.ByCreatedAt(sql.OrderAsc()), task.ByID(sql.OrderAsc()))
	results, queryErr := ent.DoListQuery[ent.Task, *ent.TaskQuery](ctx, query, params.ListParams)
	if queryErr != nil {
		return nil, fmt.Errorf("list incident tasks: %w", queryErr)
	}
	return results, nil
}

func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*ent.Task, error) {
	return s.db.Client(ctx).Task.Get(ctx, id)
}

func (s *TaskService) SetTask(ctx context.Context, id uuid.UUID, params rez.SetTaskParams) (*ent.Task, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Task, error) {
		var mut ent.EntityMutator[*ent.Task, *ent.TaskMutation]
		if id == uuid.Nil {
			mut = tx.Task.Create()
		} else {
			mut = tx.Task.UpdateOneID(id)
		}

		m := mut.Mutation()
		if params.Title != nil {
			m.SetTitle(*params.Title)
		}
		if params.Description != nil {
			m.SetDescription(*params.Description)
		}
		if params.OwnerID == nil {
			m.ClearAssigneeID()
		} else {
			m.SetAssigneeID(*params.OwnerID)
		}
		if params.DueAt == nil {
			m.ClearDueAt()
		} else {
			m.SetDueAt(*params.DueAt)
		}

		saved, saveErr := mut.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("update task: %w", saveErr)
		}
		return saved, nil
	})
}

func (s *TaskService) ArchiveTask(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: task ID required", rez.ErrInvalidInput)
	}
	return s.db.Client(ctx).Task.DeleteOneID(id).Exec(ctx)
}
