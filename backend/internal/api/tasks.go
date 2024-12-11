package api

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/schema"
	oapi "github.com/twohundreds/rezible/openapi"
)

type tasksHandler struct {
	db *ent.Client
}

func newTasksHandler(db *ent.Client) *tasksHandler {
	return &tasksHandler{db}
}

func (h *tasksHandler) ListTasks(ctx context.Context, request *oapi.ListTasksRequest) (*oapi.ListTasksResponse, error) {
	var resp oapi.ListTasksResponse

	query := h.db.Task.Query().
		Limit(request.Limit).
		Offset(request.Offset)

	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}

	tasks, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to fetch tasks", queryErr)
	}

	resp.Body.Data = make([]oapi.Task, len(tasks))
	for i, task := range tasks {
		resp.Body.Data[i] = oapi.TaskFromEnt(task)
	}

	return &resp, nil
}

func (h *tasksHandler) CreateTask(ctx context.Context, request *oapi.CreateTaskRequest) (*oapi.CreateTaskResponse, error) {
	var resp oapi.CreateTaskResponse

	return &resp, nil
}

func (h *tasksHandler) GetTask(ctx context.Context, request *oapi.GetTaskRequest) (*oapi.GetTaskResponse, error) {
	var resp oapi.GetTaskResponse

	task, queryErr := h.db.Task.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to fetch task", queryErr)
	}
	resp.Body.Data = oapi.TaskFromEnt(task)

	return &resp, nil
}

func (h *tasksHandler) UpdateTask(ctx context.Context, request *oapi.UpdateTaskRequest) (*oapi.UpdateTaskResponse, error) {
	var resp oapi.UpdateTaskResponse

	return &resp, nil
}

func (h *tasksHandler) ArchiveTask(ctx context.Context, request *oapi.ArchiveTaskRequest) (*oapi.ArchiveTaskResponse, error) {
	var resp oapi.ArchiveTaskResponse

	archiveErr := h.db.Task.DeleteOneID(request.Id).Exec(ctx)
	if archiveErr != nil {
		return nil, detailError("failed to archive task", archiveErr)
	}

	return &resp, nil
}
