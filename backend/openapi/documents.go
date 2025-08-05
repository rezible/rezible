package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
)

type DocumentsHandler interface {
	RequestDocumentEditorSession(context.Context, *RequestDocumentEditorSessionRequest) (*RequestDocumentEditorSessionResponse, error)
	VerifyDocumentEditorSession(context.Context, *VerifyDocumentEditorSessionRequest) (*VerifyDocumentEditorSessionResponse, error)
}

func (o operations) RegisterDocuments(api huma.API) {
	huma.Register(api, RequestDocumentEditorSession, o.RequestDocumentEditorSession)
	huma.Register(api, VerifyDocumentEditorSession, o.VerifyDocumentEditorSession)
}

type (
	DocumentEditorSession struct {
		DocumentName  string `json:"documentName"`
		ConnectionUrl string `json:"connectionUrl"`
		Token         string `json:"token"`
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
	Path:        "/documents/session/new",
	Summary:     "Request a Document Editor Session",
	Tags:        documentsTags,
	Errors:      errorCodes(),
}

type RequestDocumentEditorSessionAttributes struct {
	DocumentName string `json:"documentName"`
}
type RequestDocumentEditorSessionRequest RequestWithBodyAttributes[RequestDocumentEditorSessionAttributes]
type RequestDocumentEditorSessionResponse ItemResponse[DocumentEditorSession]

var VerifyDocumentEditorSession = huma.Operation{
	OperationID: "verify-document-editor-session",
	Method:      http.MethodPost,
	Path:        "/documents/session/verify",
	Summary:     "Verify a Document Editor Session",
	Tags:        documentsTags,
	Errors:      errorCodes(),
	Security: []map[string][]string{
		{SecurityMethodAuthSessionToken: {}},
	},
}

type VerifyDocumentEditorSessionRequestAttributes struct {
	DocumentName string `json:"documentName"`
}
type VerifyDocumentEditorSessionRequest RequestWithBodyAttributes[VerifyDocumentEditorSessionRequestAttributes]
type VerifyDocumentEditorSessionResponse ItemResponse[DocumentEditorSessionAuth]
