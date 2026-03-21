package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type DocumentsHandler interface {
	GetDocumentAccess(context.Context, *GetDocumentAccessRequest) (*GetDocumentAccessResponse, error)
	LoadDocument(context.Context, *LoadDocumentRequest) (*LoadDocumentResponse, error)
	UpdateDocument(context.Context, *UpdateDocumentRequest) (*UpdateDocumentResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, GetDocumentAccess, o.GetDocumentAccess)
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

	DocumentAccess struct {
		User      User `json:"user"`
		CanEdit   bool `json:"canEdit"`
		CanManage bool `json:"canManage"`
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
		User:      User{Id: acc.UserID},
		CanEdit:   acc.CanEdit,
		CanManage: acc.CanManage,
	}
	if acc.Edges.User != nil {
		da.User = UserFromEnt(acc.Edges.User)
	}
	return da
}

var documentsTags = []string{"documents"}

var GetDocumentAccess = huma.Operation{
	OperationID: "get-document-access",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/access",
	Summary:     "Get user access for a document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type GetDocumentAccessRequest GetIdRequest
type GetDocumentAccessResponse ItemResponse[DocumentAccess]

var LoadDocument = huma.Operation{
	OperationID: "load-document",
	Method:      http.MethodGet,
	Path:        "/documents/{id}/load",
	Summary:     "Load document",
	Tags:        documentsTags,
	Errors:      ErrorCodes(),
}

type LoadDocumentRequest GetIdRequest
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
type UpdateDocumentRequest PostIdRequest[UpdateDocumentRequestAttributes]
type UpdateDocumentResponse ItemResponse[Document]
