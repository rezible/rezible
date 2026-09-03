package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organizationrole"
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
		Name             string `json:"name"`
		Email            string `json:"email"`
		OrganizationRole string `json:"organizationRole" enum:"admin,member"`
	}
)

func UserFromEnt(user *ent.User) User {
	attr := UserAttributes{
		Name:             user.Name,
		Email:            user.Email,
		OrganizationRole: organizationrole.RoleMember.String(),
	}
	if role := user.Edges.OrganizationRole; role != nil && role.Role == organizationrole.RoleAdmin {
		attr.OrganizationRole = organizationrole.RoleAdmin.String()
	}

	return User{
		Id:         user.ID,
		Attributes: attr,
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
