package openapi

/*
type LaddersHandler interface {
	ListLadders(context.Context, *ListLaddersRequest) (*ListLaddersResponse, error)
	GetLadder(context.Context, *GetLadderRequest) (*GetLadderResponse, error)
	CreateLadder(context.Context, *CreateLadderRequest) (*CreateLadderResponse, error)
	UpdateLadder(context.Context, *UpdateLadderRequest) (*UpdateLadderResponse, error)
	ArchiveLadder(context.Context, *ArchiveLadderRequest) (*ArchiveLadderResponse, error)
}

func (o operations) RegisterLadders(api huma.API) {
	huma.Register(api, ListLadders, o.ListLadders)
	huma.Register(api, GetLadder, o.GetLadder)
	huma.Register(api, CreateLadder, o.CreateLadder)
	huma.Register(api, UpdateLadder, o.UpdateLadder)
	huma.Register(api, ArchiveLadder, o.ArchiveLadder)
}

type Ladder struct {
	Id         uuid.UUID        `json:"id"`
	Attributes LadderAttributes `json:"attributes"`
}

type LadderAttributes struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func LadderFromEnt(l *ent.Ladder) Ladder {
	return Ladder{
		Id: l.ID,
		Attributes: LadderAttributes{
			Name:    l.Name,
			Private: false,
		},
	}
}

//

var laddersTags = []string{"Ladders"}

var ListLadders = huma.Operation{
	OperationID: "list-ladders",
	Method:      http.MethodGet,
	Path:        "/ladders",
	Summary:     "List Ladders",
	Tags:        laddersTags,
	Errors:      errorCodes(),
}

type ListLaddersRequest EmptyRequest
type ListLaddersResponse PaginatedResponse[Ladder]

var CreateLadder = huma.Operation{
	OperationID: "create-ladder",
	Method:      http.MethodPost,
	Path:        "/ladders",
	Summary:     "Create a Ladder",
	Tags:        laddersTags,
	Errors:      errorCodes(),
}

type CreateLadderAttributes struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}
type CreateLadderRequest RequestWithBodyAttributes[CreateLadderAttributes]
type CreateLadderResponse ItemResponse[Ladder]

var GetLadder = huma.Operation{
	OperationID: "get-ladder",
	Method:      http.MethodGet,
	Path:        "/ladders/{id}",
	Summary:     "Get a Ladder",
	Tags:        laddersTags,
	Errors:      errorCodes(),
}

type GetLadderRequest GetIdRequest
type GetLadderResponse ItemResponse[Ladder]

var UpdateLadder = huma.Operation{
	OperationID: "update-ladder",
	Method:      http.MethodPatch,
	Path:        "/ladders/{id}",
	Summary:     "Update a Ladder",
	Tags:        laddersTags,
	Errors:      errorCodes(),
}

type UpdateLadderAttributes struct {
	Name    OmittableNullable[string] `json:"name"`
	Private OmittableNullable[bool]   `json:"private"`
}
type UpdateLadderRequest UpdateIdRequest[UpdateLadderAttributes]
type UpdateLadderResponse ItemResponse[Ladder]

var ArchiveLadder = huma.Operation{
	OperationID: "archive-ladder",
	Method:      http.MethodDelete,
	Path:        "/ladders/{id}",
	Summary:     "Archive a Ladder",
	Tags:        laddersTags,
	Errors:      errorCodes(),
}

type ArchiveLadderRequest ArchiveIdRequest
type ArchiveLadderResponse EmptyResponse
*/
