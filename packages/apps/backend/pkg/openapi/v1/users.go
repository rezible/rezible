package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
)

type UsersHandler interface {
	ListUsers(context.Context, *ListUsersRequest) (*ListUsersResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
}

func (o operations) RegisterUsers(api huma.API) {
	huma.Register(api, ListUsers, o.ListUsers)
	huma.Register(api, GetUser, o.GetUser)
}

type (
	User struct {
		Id         uuid.UUID      `json:"id"`
		Attributes UserAttributes `json:"attributes"`
	}

	UserAttributes struct {
		Name             string    `json:"name"`
		Email            string    `json:"email"`
		OrganizationRole user.Role `json:"organizationRole"`
	}
)

func UserFromEnt(u *ent.User) User {
	return User{
		Id: u.ID,
		Attributes: UserAttributes{
			Name:             u.Name,
			Email:            u.Email,
			OrganizationRole: u.Role,
		},
	}
}

// Operations

var usersTags = []string{"Users"}

var ListUsers = huma.Operation{
	OperationID: "list-users",
	Method:      http.MethodGet,
	Path:        "/users",
	Summary:     "List Users",
	Tags:        usersTags,
	Errors:      ErrorCodes(),
}

type ListUsersRequest struct {
	PaginationRequest
	Search string    `query:"search" required:"false" nullable:"false"`
	TeamId uuid.UUID `query:"teamId" required:"false"`
}
type ListUsersResponse PaginatedResponse[User]

var GetUser = huma.Operation{
	OperationID: "get-user",
	Method:      http.MethodGet,
	Path:        "/users/{id}",
	Summary:     "Get a User",
	Tags:        usersTags,
	Errors:      ErrorCodes(),
}

type GetUserRequest IdRequest
type GetUserResponse ItemResponse[User]
