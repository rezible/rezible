package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/task"
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

func (o operations) RegisterTaskEnums(api huma.API) {
	registerEnumAlias[task.State, taskStateSchema](api)
}

type taskStateSchema task.State

func (taskStateSchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(task.StateValues)
}

type (
	Task struct {
		Id         uuid.UUID      `json:"id"`
		Attributes TaskAttributes `json:"attributes"`
	}

	TaskAttributes struct {
		Title           string                      `json:"title"`
		Description     string                      `json:"description"`
		Owner           *Expandable[UserAttributes] `json:"ownerId,omitempty"`
		Author          *Expandable[UserAttributes] `json:"author,omitempty"`
		State           task.State                  `json:"state"`
		IncidentId      *uuid.UUID                  `json:"incidentId,omitempty"`
		DueAt           *time.Time                  `json:"dueAt,omitempty"`
		CreatedAt       time.Time                   `json:"createdAt"`
		UpdatedAt       time.Time                   `json:"updatedAt"`
		ArchivedAt      *time.Time                  `json:"archivedAt,omitempty"`
		ExternalTickets []ExternalTicket            `json:"externalTickets"`
	}

	ExternalTicket struct {
		Id                  uuid.UUID `json:"id"`
		Title               string    `json:"title"`
		Reference           *string   `json:"reference,omitempty"`
		URL                 *string   `json:"url,omitempty"`
		Provider            *string   `json:"provider,omitempty"`
		ProviderNamespace   *string   `json:"providerNamespace,omitempty"`
		ProviderResourceRef *string   `json:"providerResourceRef,omitempty"`
	}
)

func TaskFromEnt(t *ent.Task) Task {
	attrs := TaskAttributes{
		Title:           t.Title,
		Description:     t.Description,
		State:           t.State,
		IncidentId:      t.IncidentID,
		DueAt:           t.DueAt,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		ArchivedAt:      t.ArchiveTime,
		ExternalTickets: nil,
	}

	if t.CreatorID != nil {
		attrs.Author = &Expandable[UserAttributes]{Id: *t.CreatorID}
		if t.Edges.Creator != nil {
			attrs.Author.Attributes = new(UserFromEnt(t.Edges.Creator).Attributes)
		}
	}

	if t.AssigneeID != nil {
		attrs.Owner = &Expandable[UserAttributes]{Id: *t.AssigneeID}
		if t.Edges.Assignee != nil {
			attrs.Owner.Attributes = new(UserFromEnt(t.Edges.Assignee).Attributes)
		}
	}

	if tickets, ticketsErr := t.Edges.TicketsOrErr(); ticketsErr == nil {
		attrs.ExternalTickets = make([]ExternalTicket, len(tickets))
		for i, ticket := range tickets {
			attrs.ExternalTickets[i] = ExternalTicket{
				Id:                  ticket.ID,
				Title:               "",
				Reference:           nil,
				URL:                 nil,
				Provider:            nil,
				ProviderNamespace:   nil,
				ProviderResourceRef: nil,
			}
		}
	}

	return Task{Id: t.ID, Attributes: attrs}
}

var tasksTags = []string{"Tasks"}

// ops

var ListTasks = huma.Operation{
	OperationID: "list-tasks",
	Method:      http.MethodGet,
	Path:        "/tasks",
	Summary:     "List Tasks",
	Tags:        tasksTags,
	Errors:      ErrorCodes(),
}

type ListTasksRequest struct {
	PaginationRequest
	IncludeArchived bool       `query:"archived" required:"false" nullable:"false" default:"false"`
	IncidentId      uuid.UUID  `query:"incidentId" required:"false"`
	OwnerId         uuid.UUID  `query:"ownerId" required:"false"`
	OriginEntryId   uuid.UUID  `query:"originEntryId" required:"false"`
	State           task.State `query:"state" required:"false"`
}

type ListTasksResponse PaginatedResponse[Task]

var GetTask = huma.Operation{
	OperationID: "get-task",
	Method:      http.MethodGet,
	Path:        "/tasks/{id}",
	Summary:     "Get Task",
	Tags:        tasksTags,
	Errors:      ErrorCodes(),
}

type GetTaskRequest IdRequest
type GetTaskResponse ItemResponse[Task]

var CreateTask = huma.Operation{
	OperationID: "create-task",
	Method:      http.MethodPost,
	Path:        "/tasks",
	Summary:     "Create a Task",
	Tags:        tasksTags,
	Errors:      ErrorCodes(),
}

type CreateTaskAttributes struct {
	Title         string      `json:"title"`
	Description   string      `json:"description,omitempty"`
	IncidentId    *uuid.UUID  `json:"incidentId,omitempty"`
	OriginEntryId *uuid.UUID  `json:"originEntryId,omitempty"`
	State         *task.State `json:"state,omitempty"`
	DueAt         *time.Time  `json:"dueAt,omitempty"`
}

type CreateTaskRequest RequestWithBodyAttributes[CreateTaskAttributes]
type CreateTaskResponse ItemResponse[Task]

var UpdateTask = huma.Operation{
	OperationID: "update-task",
	Method:      http.MethodPatch,
	Path:        "/tasks/{id}",
	Summary:     "Update a Task",
	Tags:        tasksTags,
	Errors:      ErrorCodes(),
}

type UpdateTaskAttributes struct {
	Title         OmittableNullable[string]     `json:"title"`
	Description   OmittableNullable[string]     `json:"description"`
	OwnerId       OmittableNullable[uuid.UUID]  `json:"ownerId"`
	State         OmittableNullable[task.State] `json:"state"`
	DueAt         OmittableNullable[time.Time]  `json:"dueAt"`
	OriginEntryId OmittableNullable[uuid.UUID]  `json:"originEntryId"`
}

type UpdateTaskRequest IdRequestWithBody[UpdateTaskAttributes]
type UpdateTaskResponse ItemResponse[Task]

var ArchiveTask = huma.Operation{
	OperationID: "archive-task",
	Method:      http.MethodDelete,
	Path:        "/tasks/{id}",
	Summary:     "Archive a Task",
	Tags:        tasksTags,
	Errors:      ErrorCodes(),
}

type ArchiveTaskRequest IdRequest
type ArchiveTaskResponse EmptyResponse
