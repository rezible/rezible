package api

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type documentsHandler struct {
	documents rez.DocumentsService
	auth      rez.AuthService
	users     rez.UserService
}

func newDocumentsHandler(documents rez.DocumentsService, auth rez.AuthService, users rez.UserService) *documentsHandler {
	return &documentsHandler{documents, auth, users}
}

func (h *documentsHandler) verifyUserDocumentAccess(ctx context.Context, userId uuid.UUID, docId uuid.UUID) (bool, error) {
	// TODO: lookup document using ent, check tenant
	const readOnly = false
	return readOnly, nil
}

func (h *documentsHandler) RequestDocumentEditorSession(ctx context.Context, request *oapi.RequestDocumentEditorSessionRequest) (*oapi.RequestDocumentEditorSessionResponse, error) {
	var resp oapi.RequestDocumentEditorSessionResponse

	sess := getRequestAuthSession(ctx, h.auth)

	docId := request.Id
	_, accessErr := h.verifyUserDocumentAccess(ctx, sess.UserId, docId)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	token, tokenErr := h.documents.CreateEditorSessionToken(sess, docId)
	if tokenErr != nil {
		return nil, apiError("failed to create session token", tokenErr)
	}

	resp.Body.Data = oapi.DocumentEditorSession{
		DocumentId:    docId,
		Token:         token,
		ConnectionUrl: h.documents.GetServerWebsocketAddress(),
	}

	return &resp, nil
}
