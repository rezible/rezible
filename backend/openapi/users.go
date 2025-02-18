package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
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
		Name  string `json:"name"`
		Email string `json:"email"`
	}
)

func UserFromEnt(user *ent.User) User {
	return User{
		Id: user.ID,
		Attributes: UserAttributes{
			Name:  user.Name,
			Email: "",
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
	Errors:      errorCodes(),
}

type ListUsersRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"teamId" required:"false"`
}
type ListUsersResponse PaginatedResponse[User]

var GetUser = huma.Operation{
	OperationID: "get-user",
	Method:      http.MethodGet,
	Path:        "/users/{id}",
	Summary:     "Get a User",
	Tags:        usersTags,
	Errors:      errorCodes(),
}

type GetUserRequest GetIdRequest
type GetUserResponse ItemResponse[User]
