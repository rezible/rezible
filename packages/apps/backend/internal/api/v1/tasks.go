package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type tasksHandler struct {
	*baseHandler
	tasks rez.TaskService
}

func newTasksHandler(bh *baseHandler, tasks rez.TaskService) *tasksHandler {
	return &tasksHandler{bh, tasks}
}

func (h *tasksHandler) ListTasks(ctx context.Context, request *oapi.ListTasksRequest) (*oapi.ListTasksResponse, error) {
	params := rez.ListTasksParams{
		ListParams:    request.ListParams(),
		IncidentID:    request.IncidentId,
		OwnerID:       request.OwnerId,
		SourceEntryID: request.OriginEntryId,
		State:         request.State,
	}
	params.IncludeArchived = request.IncludeArchived
	tasks, queryErr := h.tasks.ListTasks(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch tasks", queryErr)
	}
	var resp oapi.ListTasksResponse
	resp.Body = oapi.ConvertPaginatedResultBody(tasks, oapi.TaskFromEnt)
	return &resp, nil
}

func (h *tasksHandler) CreateTask(ctx context.Context, request *oapi.CreateTaskRequest) (*oapi.CreateTaskResponse, error) {
	attrs := request.Body.Attributes
	params := rez.SetTaskParams{
		Title:         &attrs.Title,
		Description:   &attrs.Description,
		IncidentID:    attrs.IncidentId,
		DueAt:         attrs.DueAt,
		SourceEntryID: attrs.OriginEntryId,
	}
	created, createErr := h.tasks.SetTask(ctx, uuid.Nil, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create task", createErr)
	}
	var resp oapi.CreateTaskResponse
	resp.Body.Data = oapi.TaskFromEnt(created)
	return &resp, nil
}

func (h *tasksHandler) GetTask(ctx context.Context, request *oapi.GetTaskRequest) (*oapi.GetTaskResponse, error) {
	taskRecord, queryErr := h.tasks.GetTask(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch task", queryErr)
	}
	var resp oapi.GetTaskResponse
	resp.Body.Data = oapi.TaskFromEnt(taskRecord)
	return &resp, nil
}

func (h *tasksHandler) UpdateTask(ctx context.Context, request *oapi.UpdateTaskRequest) (*oapi.UpdateTaskResponse, error) {
	attrs := request.Body.Attributes
	params := rez.SetTaskParams{
		Title:       attrs.Title.NillableValue(),
		Description: attrs.Description.NillableValue(),
	}
	updated, updateErr := h.tasks.SetTask(ctx, request.Id, params)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update task", updateErr)
	}
	var resp oapi.UpdateTaskResponse
	resp.Body.Data = oapi.TaskFromEnt(updated)
	return &resp, nil
}

func (h *tasksHandler) ArchiveTask(ctx context.Context, request *oapi.ArchiveTaskRequest) (*oapi.ArchiveTaskResponse, error) {
	if archiveErr := h.tasks.ArchiveTask(ctx, request.Id); archiveErr != nil {
		return nil, oapi.Error(ctx, "archive task", archiveErr)
	}
	return &oapi.ArchiveTaskResponse{}, nil
}
