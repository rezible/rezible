package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type DocumentsHandler interface {
	RequestDocumentSessionAuth(context.Context, *RequestDocumentSessionAuthRequest) (*RequestDocumentSessionAuthResponse, error)
	GetDocumentSession(context.Context, *GetDocumentSessionRequest) (*GetDocumentSessionResponse, error)
	LoadDocument(context.Context, *LoadDocumentRequest) (*LoadDocumentResponse, error)
	UpdateDocument(context.Context, *UpdateDocumentRequest) (*UpdateDocumentResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentSessionAuth, o.RequestDocumentSessionAuth)
	huma.Register(api, GetDocumentSession, o.GetDocumentSession)
	//huma.Register(api, LoadDocument, o.LoadDocument)
	//huma.Register(api, UpdateDocument, o.UpdateDocument)
}

type (
	Document struct {
		Id         uuid.UUID          `json:"id"`
		Attributes DocumentAttributes `json:"attributes"`
	}

	DocumentAttributes struct {
		Content string `json:"content"`
	}

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

func DocumentFromEnt(doc *ent.Document) Document {
	attrs := DocumentAttributes{
		Content: string(doc.Content),
	}
	return Document{Id: doc.ID, Attributes: attrs}
}

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

type RequestDocumentSessionAuthRequest EmptyIdRequest
type RequestDocumentSessionAuthResponse ItemResponse[DocumentSessionAuth]

var GetDocumentSession = huma.Operation{
	OperationID: "get-document-session",
	Method:      http.MethodGet,
	Path:        "/documents/{id}/session",
	Summary:     "Get document session",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
	Security: SecurityMethodOptions{
		{SecurityMethodScopedSessionToken: {"documents:*"}},
	},
}

type GetDocumentSessionRequest EmptyIdRequest
type GetDocumentSessionResponse ItemResponse[DocumentSession]

var LoadDocument = huma.Operation{
	OperationID: "load-document",
	Method:      http.MethodGet,
	Path:        "/documents/{id}/load",
	Summary:     "Load document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type LoadDocumentRequest EmptyIdRequest
type LoadDocumentResponse ItemResponse[Document]

var UpdateDocument = huma.Operation{
	OperationID: "update-document",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/update",
	Summary:     "Update document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type UpdateDocumentRequestAttributes struct {
	Content json.RawMessage `json:"content"`
}
type UpdateDocumentRequest IdRequest[UpdateDocumentRequestAttributes]
type UpdateDocumentResponse ItemResponse[Document]
