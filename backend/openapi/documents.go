package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type DocumentsHandler interface {
	RequestDocumentEditorSession(context.Context, *RequestDocumentEditorSessionRequest) (*RequestDocumentEditorSessionResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentEditorSession, o.RequestDocumentEditorSession)
}

type (
	DocumentEditorSession struct {
		DocumentId    uuid.UUID `json:"documentId"`
		ConnectionUrl string    `json:"connectionUrl"`
		Token         string    `json:"token"`
	}

	DocumentEditorSessionAuth struct {
		User     DocumentEditorSessionUser `json:"user"`
		ReadOnly bool                      `json:"readOnly"`
	}

	DocumentEditorSessionUser struct {
		Id       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	}
)

var documentsTags = []string{"Documents"}

var RequestDocumentEditorSession = huma.Operation{
	OperationID: "request-document-editor-session",
	Method:      http.MethodPost,
	Path:        "/documents/{id}/session",
	Summary:     "Request a Document Editor Session",
	Tags:        documentsTags,
	Errors:      errorCodes(),
}

type RequestDocumentEditorSessionRequest PostIdRequest
type RequestDocumentEditorSessionResponse ItemResponse[DocumentEditorSession]
