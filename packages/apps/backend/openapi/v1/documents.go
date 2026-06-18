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
	RequestDocumentSession(context.Context, *RequestDocumentSessionRequest) (*RequestDocumentSessionResponse, error)
	GetDocumentAccess(context.Context, *GetDocumentAccessRequest) (*GetDocumentAccessResponse, error)
	LoadDocument(context.Context, *LoadDocumentRequest) (*LoadDocumentResponse, error)
	UpdateDocument(context.Context, *UpdateDocumentRequest) (*UpdateDocumentResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentSession, o.RequestDocumentSession)
	huma.Register(api, GetDocumentAccess, o.GetDocumentAccess)
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

	DocumentAccess struct {
		CanView   bool `json:"canView"`
		CanEdit   bool `json:"canEdit"`
		CanManage bool `json:"canManage"`
	}

	DocumentSession struct {
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

func DocumentAccessFromEnt(acc *ent.DocumentAccess) DocumentAccess {
	da := DocumentAccess{
		CanView:   acc.CanView,
		CanEdit:   acc.CanEdit,
		CanManage: acc.CanManage,
	}
	return da
}

func DocumentSessionFromRez(ds *rez.DocumentSession) DocumentSession {
	return DocumentSession{
		Token:     ds.Token,
		ServerUrl: ds.ServerUrl.String(),
	}
}

var documentsTags = []string{"documents"}

var RequestDocumentSession = huma.Operation{
	OperationID: "request-document-editor-session",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/session",
	Summary:     "Request a session for a document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type RequestDocumentSessionRequest EmptyIdRequest
type RequestDocumentSessionResponse ItemResponse[DocumentSession]

var GetDocumentAccess = huma.Operation{
	OperationID: "get-document-access",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/access",
	Summary:     "Get user access for a document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type GetDocumentAccessRequest EmptyIdRequest
type GetDocumentAccessResponse ItemResponse[DocumentAccess]

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
