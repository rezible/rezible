package apiv1

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/task"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type tasksHandler struct {
	db rez.Database
}

func newTasksHandler(db rez.Database) *tasksHandler {
	return &tasksHandler{db: db}
}

func (h *tasksHandler) ListTasks(ctx context.Context, request *oapi.ListTasksRequest) (*oapi.ListTasksResponse, error) {
	var resp oapi.ListTasksResponse

	query := h.db.Client(ctx).Task.Query()
	if request.IncidentId != uuid.Nil {
		query = query.Where(task.IncidentID(request.IncidentId))
	}
	if request.OwnerId != uuid.Nil {
		query = query.Where(task.AssigneeID(request.OwnerId))
	}
	if request.OriginEntryId != uuid.Nil {
		query = query.Where(task.OriginEntryID(request.OriginEntryId))
	}
	if request.State != "" {
		query = query.Where(task.StateEQ(task.State(request.State)))
	}
	query = query.WithTickets().Order(task.ByID())
	params := request.ListParams()
	params.IncludeArchived = request.IncludeArchived
	tasks, queryErr := ent.DoListQuery[ent.Task, *ent.TaskQuery](ctx, query, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch tasks", queryErr)
	}

	resp.Body = oapi.ConvertPaginatedResultBody(tasks, oapi.TaskFromEnt)

	return &resp, nil
}

func (h *tasksHandler) CreateTask(ctx context.Context, request *oapi.CreateTaskRequest) (*oapi.CreateTaskResponse, error) {
	return nil, oapi.Error(ctx, "create task is not implemented", rez.ErrNotImplemented)
}

func (h *tasksHandler) GetTask(ctx context.Context, request *oapi.GetTaskRequest) (*oapi.GetTaskResponse, error) {
	var resp oapi.GetTaskResponse

	taskQuery := h.db.Client(ctx).Task.Query().Where(task.ID(request.Id)).WithTickets()
	t, queryErr := taskQuery.Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch task", queryErr)
	}
	resp.Body.Data = oapi.TaskFromEnt(t)

	return &resp, nil
}

func (h *tasksHandler) UpdateTask(ctx context.Context, request *oapi.UpdateTaskRequest) (*oapi.UpdateTaskResponse, error) {
	return nil, oapi.Error(ctx, "update task is not implemented", rez.ErrNotImplemented)
}

func (h *tasksHandler) ArchiveTask(ctx context.Context, request *oapi.ArchiveTaskRequest) (*oapi.ArchiveTaskResponse, error) {
	var resp oapi.ArchiveTaskResponse

	archiveErr := h.db.Client(ctx).Task.DeleteOneID(request.Id).Exec(ctx)
	if archiveErr != nil {
		return nil, oapi.Error(ctx, "failed to archive task", archiveErr)
	}

	return &resp, nil
}
