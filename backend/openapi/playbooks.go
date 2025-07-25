package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type PlaybooksHandler interface {
	ListPlaybooks(context.Context, *ListPlaybooksRequest) (*ListPlaybooksResponse, error)
	CreatePlaybook(context.Context, *CreatePlaybookRequest) (*CreatePlaybookResponse, error)
	GetPlaybook(context.Context, *GetPlaybookRequest) (*GetPlaybookResponse, error)
	UpdatePlaybook(context.Context, *UpdatePlaybookRequest) (*UpdatePlaybookResponse, error)
	ArchivePlaybook(context.Context, *ArchivePlaybookRequest) (*ArchivePlaybookResponse, error)
}

func (o operations) RegisterPlaybooks(api huma.API) {
	huma.Register(api, ListPlaybooks, o.ListPlaybooks)
	huma.Register(api, CreatePlaybook, o.CreatePlaybook)
	huma.Register(api, GetPlaybook, o.GetPlaybook)
	huma.Register(api, UpdatePlaybook, o.UpdatePlaybook)
	huma.Register(api, ArchivePlaybook, o.ArchivePlaybook)
}

type (
	Playbook struct {
		Id         uuid.UUID          `json:"id"`
		Attributes PlaybookAttributes `json:"attributes"`
	}

	PlaybookAttributes struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
)

func PlaybookFromEnt(pb *ent.Playbook) Playbook {
	attrs := PlaybookAttributes{
		Title:   pb.Title,
		Content: string(pb.Content),
	}
	return Playbook{
		Id:         pb.ID,
		Attributes: attrs,
	}
}

var playbooksTags = []string{"Playbooks"}

// ops

var ListPlaybooks = huma.Operation{
	OperationID: "list-playbooks",
	Method:      http.MethodGet,
	Path:        "/playbooks",
	Summary:     "List Playbooks",
	Tags:        playbooksTags,
	Errors:      errorCodes(),
}

type ListPlaybooksRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"teamId" required:"false"`
}
type ListPlaybooksResponse PaginatedResponse[Playbook]

var GetPlaybook = huma.Operation{
	OperationID: "get-playbook",
	Method:      http.MethodGet,
	Path:        "/playbooks/{id}",
	Summary:     "Get Playbook",
	Tags:        playbooksTags,
	Errors:      errorCodes(),
}

type GetPlaybookRequest GetIdRequest
type GetPlaybookResponse ItemResponse[Playbook]

var CreatePlaybook = huma.Operation{
	OperationID: "create-playbook",
	Method:      http.MethodPost,
	Path:        "/playbooks",
	Summary:     "Create a Playbook",
	Tags:        playbooksTags,
	Errors:      errorCodes(),
}

type CreatePlaybookAttributes struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type CreatePlaybookRequest RequestWithBodyAttributes[CreatePlaybookAttributes]
type CreatePlaybookResponse ItemResponse[Playbook]

var UpdatePlaybook = huma.Operation{
	OperationID: "update-playbook",
	Method:      http.MethodPatch,
	Path:        "/playbooks/{id}",
	Summary:     "Update a Playbook",
	Tags:        playbooksTags,
	Errors:      errorCodes(),
}

type UpdatePlaybookAttributes struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}
type UpdatePlaybookRequest UpdateIdRequest[UpdatePlaybookAttributes]
type UpdatePlaybookResponse ItemResponse[Playbook]

var ArchivePlaybook = huma.Operation{
	OperationID: "archive-playbook",
	Method:      http.MethodDelete,
	Path:        "/playbooks/{id}",
	Summary:     "Archive a Playbook",
	Tags:        playbooksTags,
	Errors:      errorCodes(),
}

type ArchivePlaybookRequest ArchiveIdRequest
type ArchivePlaybookResponse EmptyResponse
