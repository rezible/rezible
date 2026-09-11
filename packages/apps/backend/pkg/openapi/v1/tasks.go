package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
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
		Name          string       `json:"name"`
		Description   string       `json:"description"`
		IncidentId    *uuid.UUID   `json:"incidentId,omitempty"`
		OwnerId       *uuid.UUID   `json:"ownerId,omitempty"`
		State         string       `json:"state" enum:"open,completed,cancelled"`
		DueAt         *time.Time   `json:"dueAt,omitempty"`
		OriginEntryId *uuid.UUID   `json:"originEntryId,omitempty"`
		TicketIds     []uuid.UUID  `json:"ticketIds"`
		Tickets       []TaskTicket `json:"tickets"`
		CreatedAt     time.Time    `json:"createdAt"`
		UpdatedAt     time.Time    `json:"updatedAt"`
	}

	ExternalTicketProvider string

	ExternalTicket struct {
		Provider ExternalTicketProvider `json:"provider"`
	}

	TaskTicket struct {
		Id                  uuid.UUID `json:"id"`
		Title               string    `json:"title"`
		Reference           *string   `json:"reference,omitempty"`
		URL                 *string   `json:"url,omitempty"`
		Provider            *string   `json:"provider,omitempty"`
		ProviderNamespace   *string   `json:"providerNamespace,omitempty"`
		ProviderResourceRef *string   `json:"providerResourceRef,omitempty"`
	}
)

var (
	ExternalTicketProviderJira = ExternalTicketProvider("jira")
)

func TaskFromEnt(task *ent.Task) Task {
	var incidentID *uuid.UUID
	if task.IncidentID != uuid.Nil {
		incidentID = &task.IncidentID
	}
	var ownerID *uuid.UUID
	if task.AssigneeID != uuid.Nil {
		ownerID = &task.AssigneeID
	}
	var originEntryID *uuid.UUID
	if task.OriginEntryID != nil {
		originEntryID = task.OriginEntryID
	}
	ticketIDs := make([]uuid.UUID, 0, len(task.Edges.Tickets))
	tickets := make([]TaskTicket, 0, len(task.Edges.Tickets))
	for _, ticket := range task.Edges.Tickets {
		ticketIDs = append(ticketIDs, ticket.ID)
		tickets = append(tickets, TaskTicket{
			Id:                  ticket.ID,
			Title:               ticket.Title,
			Reference:           ticket.Reference,
			URL:                 ticket.URL,
			Provider:            ticket.Provider,
			ProviderNamespace:   ticket.ProviderNamespace,
			ProviderResourceRef: ticket.ProviderResourceRef,
		})
	}
	return Task{
		Id: task.ID,
		Attributes: TaskAttributes{
			Name:          task.Title,
			IncidentId:    incidentID,
			OwnerId:       ownerID,
			State:         task.State.String(),
			DueAt:         task.DueAt,
			OriginEntryId: originEntryID,
			TicketIds:     ticketIDs,
			Tickets:       tickets,
			CreatedAt:     task.CreatedAt,
			UpdatedAt:     task.UpdatedAt,
		},
	}
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
	IncludeArchived bool      `query:"archived" required:"false" nullable:"false" default:"false"`
	IncidentId      uuid.UUID `query:"incidentId" required:"false"`
	OwnerId         uuid.UUID `query:"ownerId" required:"false"`
	OriginEntryId   uuid.UUID `query:"originEntryId" required:"false"`
	State           string    `query:"state" required:"false" enum:"open,completed,cancelled"`
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
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type CreateTaskAttributes struct {
	Name          string     `json:"title"`
	IncidentId    *uuid.UUID `json:"incidentId,omitempty"`
	OwnerId       *uuid.UUID `json:"ownerId,omitempty"`
	OriginEntryId *uuid.UUID `json:"originEntryId,omitempty"`
	State         string     `json:"state" enum:"open,completed,cancelled"`
	DueAt         *time.Time `json:"dueAt,omitempty"`
}

type CreateTaskRequest RequestWithBodyAttributes[CreateTaskAttributes]
type CreateTaskResponse ItemResponse[Task]

var UpdateTask = huma.Operation{
	OperationID: "update-task",
	Method:      http.MethodPatch,
	Path:        "/tasks/{id}",
	Summary:     "Update a Task",
	Tags:        tasksTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type UpdateTaskAttributes struct {
	Name          OmittableNullable[string]    `json:"name,omitempty"`
	OwnerId       OmittableNullable[uuid.UUID] `json:"ownerId,omitempty"`
	State         OmittableNullable[string]    `json:"state,omitempty" enum:"open,completed,cancelled"`
	DueAt         OmittableNullable[time.Time] `json:"dueAt,omitempty"`
	OriginEntryId OmittableNullable[uuid.UUID] `json:"originEntryId,omitempty"`
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
