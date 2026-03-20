package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/openapi"
)

type DocumentsHandler interface {
	RequestDocumentEditorSession(context.Context, *RequestDocumentEditorSessionRequest) (*RequestDocumentEditorSessionResponse, error)
	LoadDocument(context.Context, *LoadDocumentRequest) (*LoadDocumentResponse, error)
	UpdateDocument(context.Context, *UpdateDocumentRequest) (*UpdateDocumentResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentEditorSession, o.RequestDocumentEditorSession)
	huma.Register(api, LoadDocument, o.LoadDocument)
	huma.Register(api, UpdateDocument, o.UpdateDocument)
}

type (
	Document struct {
		Id         uuid.UUID          `json:"id"`
		Attributes DocumentAttributes `json:"attributes"`
	}

	DocumentAttributes struct {
		Content string `json:"content"`
	}

	DocumentEditorSession struct {
		DocumentId    uuid.UUID `json:"documentId"`
		ConnectionUrl string    `json:"connectionUrl"`
		AccessToken   string    `json:"accessToken"`
	}

	DocumentEditorSessionAuth struct {
		UserId    uuid.UUID `json:"userId"`
		CanEdit   bool      `json:"canEdit"`
		CanManage bool      `json:"canManage"`
	}

	DocumentEditorSessionUser struct {
		Id       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	}
)

func DocumentFromEnt(doc *ent.Document) Document {
	attrs := DocumentAttributes{
		Content: string(doc.Content),
	}
	return Document{Id: doc.ID, Attributes: attrs}
}

func DocumentEditorSessionAuthFromEnt(access *ent.DocumentAccess) DocumentEditorSessionAuth {
	sa := DocumentEditorSessionAuth{
		UserId:    access.UserID,
		CanEdit:   access.CanEdit,
		CanManage: access.CanManage,
	}
	return sa
}

var documentsTags = []string{"documents"}

var RequestDocumentEditorSession = huma.Operation{
	OperationID: "request-document-editor-session",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/session",
	Summary:     "Request a Document Editor Session",
	Tags:        documentsTags,
	Errors:      openapi.ErrorCodes(),
}

type RequestDocumentEditorSessionRequest PostIdEmptyRequest
type RequestDocumentEditorSessionResponse ItemResponse[DocumentEditorSession]

var LoadDocument = huma.Operation{
	OperationID: "load-document",
	Method:      http.MethodGet,
	Path:        "/documents/{id}/load",
	Summary:     "Load document",
	Tags:        documentsTags,
	Errors:      openapi.ErrorCodes(),
}

type LoadDocumentRequest GetIdRequest
type LoadDocumentResponse ItemResponse[Document]

var UpdateDocument = huma.Operation{
	OperationID: "update-document",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/update",
	Summary:     "Update document",
	Tags:        documentsTags,
	Errors:      openapi.ErrorCodes(),
}

type UpdateDocumentRequestAttributes struct {
	Content json.RawMessage `json:"content"`
}
type UpdateDocumentRequest PostIdRequest[UpdateDocumentRequestAttributes]
type UpdateDocumentResponse ItemResponse[Document]
