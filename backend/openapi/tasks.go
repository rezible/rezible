package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/twohundreds/rezible/ent"
)

type TasksHandler interface {
	ListTasks(context.Context, *ListTasksRequest) (*ListTasksResponse, error)
	CreateTask(context.Context, *CreateTaskRequest) (*CreateTaskResponse, error)
	GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error)
	UpdateTask(context.Context, *UpdateTaskRequest) (*UpdateTaskResponse, error)
	ArchiveTask(context.Context, *ArchiveTaskRequest) (*ArchiveTaskResponse, error)
}

func (o operations) RegisterTasks(api huma.API) {
	huma.Register(api, ListTasks, o.ListTasks)
	huma.Register(api, CreateTask, o.CreateTask)
	huma.Register(api, GetTask, o.GetTask)
	huma.Register(api, UpdateTask, o.UpdateTask)
	huma.Register(api, ArchiveTask, o.ArchiveTask)
}

type (
	Task struct {
		Id         uuid.UUID      `json:"id"`
		Attributes TaskAttributes `json:"attributes"`
	}

	TaskAttributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	ExternalTicketProvider string

	ExternalTicket struct {
		Provider ExternalTicketProvider `json:"provider"`
	}
)

var (
	ExternalTicketProviderJira = ExternalTicketProvider("jira")
)

func TaskFromEnt(task *ent.Task) Task {
	return Task{
		Id:         task.ID,
		Attributes: TaskAttributes{},
	}
}

var tasksTags = []string{"Incident Tasks"}

// ops

var ListTasks = huma.Operation{
	OperationID: "list-tasks",
	Method:      http.MethodGet,
	Path:        "/tasks",
	Summary:     "List Tasks",
	Tags:        tasksTags,
	Errors:      errorCodes(),
}

type ListTasksRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"team_id" required:"false"`
}
type ListTasksResponse PaginatedResponse[Task]

var GetTask = huma.Operation{
	OperationID: "get-task",
	Method:      http.MethodGet,
	Path:        "/tasks/{id}",
	Summary:     "Get Task",
	Tags:        tasksTags,
	Errors:      errorCodes(),
}

type GetTaskRequest GetIdRequest
type GetTaskResponse ItemResponse[Task]

var CreateTask = huma.Operation{
	OperationID: "create-task",
	Method:      http.MethodPost,
	Path:        "/tasks",
	Summary:     "Create a Task",
	Tags:        tasksTags,
	Errors:      errorCodes(),
}

type CreateTaskAttributes struct {
	Name string `json:"title"`
}
type CreateTaskRequest RequestWithBodyAttributes[CreateTaskAttributes]
type CreateTaskResponse ItemResponse[Task]

var UpdateTask = huma.Operation{
	OperationID: "update-task",
	Method:      http.MethodPatch,
	Path:        "/tasks/{id}",
	Summary:     "Update a Task",
	Tags:        tasksTags,
	Errors:      errorCodes(),
}

type UpdateTaskAttributes struct {
	Name OmittableNullable[string] `json:"name,omitempty"`
}
type UpdateTaskRequest UpdateIdRequest[UpdateTaskAttributes]
type UpdateTaskResponse ItemResponse[Task]

var ArchiveTask = huma.Operation{
	OperationID: "archive-task",
	Method:      http.MethodDelete,
	Path:        "/tasks/{id}",
	Summary:     "Archive a Task",
	Tags:        tasksTags,
	Errors:      errorCodes(),
}

type ArchiveTaskRequest ArchiveIdRequest
type ArchiveTaskResponse EmptyResponse
