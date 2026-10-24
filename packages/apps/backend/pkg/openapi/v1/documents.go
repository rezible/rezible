package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type DocumentsHandler interface {
	RequestDocumentSessionAuth(context.Context, *RequestDocumentSessionAuthRequest) (*RequestDocumentSessionAuthResponse, error)
	GetDocumentSession(context.Context, *GetDocumentSessionRequest) (*GetDocumentSessionResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentSessionAuth, o.RequestDocumentSessionAuth)
	huma.Register(api, GetDocumentSession, o.GetDocumentSession)
}

type (
	DocumentSession struct {
		User   User           `json:"user"`
		Access DocumentAccess `json:"access"`
	}
	DocumentAccess struct {
		CanView   bool `json:"canView"`
		CanEdit   bool `json:"canEdit"`
		CanManage bool `json:"canManage"`
	}

	DocumentSessionAuth struct {
		Name      string `json:"name"`
		Token     string `json:"token"`
		ServerUrl string `json:"serverUrl"`
	}
)

func DocumentSessionAuthFromRez(ds *rez.DocumentSessionAuth) DocumentSessionAuth {
	return DocumentSessionAuth{
		Name:      ds.DocumentName,
		Token:     ds.Token,
		ServerUrl: ds.ServerUrl.String(),
	}
}

func DocumentAccessFromEnt(acc *ent.DocumentAccess) DocumentAccess {
	return DocumentAccess{
		CanView:   acc.CanView,
		CanEdit:   acc.CanEdit,
		CanManage: acc.CanManage,
	}
}

var documentsTags = []string{"documents"}

var RequestDocumentSessionAuth = huma.Operation{
	OperationID: "request-document-session-auth",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/session/new",
	Summary:     "Request a session for a document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type RequestDocumentSessionAuthRequest IdRequest
type RequestDocumentSessionAuthResponse ItemResponse[DocumentSessionAuth]

var GetDocumentSession = huma.Operation{
	OperationID: "get-document-session",
	Method:      http.MethodGet,
	Path:        "/documents/{id}/session",
	Summary:     "Get document session",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
	Security: OperationSecurityOptions{
		{SecurityMethodScopedSessionToken: {"documents:*"}},
	},
}

type GetDocumentSessionRequest IdRequest
type GetDocumentSessionResponse ItemResponse[DocumentSession]
