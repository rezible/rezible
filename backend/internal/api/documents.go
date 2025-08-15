package api

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type documentsHandler struct {
	documents rez.DocumentsService
	auth      rez.AuthSessionService
	users     rez.UserService
}

func newDocumentsHandler(documents rez.DocumentsService, auth rez.AuthSessionService, users rez.UserService) *documentsHandler {
	return &documentsHandler{documents, auth, users}
}

func (h *documentsHandler) verifyUserDocumentAccess(ctx context.Context, userId uuid.UUID, document string) (bool, error) {
	// TODO: lookup document using ent, check tenant
	const readOnly = false
	return readOnly, nil
}

func (h *documentsHandler) RequestDocumentEditorSession(ctx context.Context, request *oapi.RequestDocumentEditorSessionRequest) (*oapi.RequestDocumentEditorSessionResponse, error) {
	var resp oapi.RequestDocumentEditorSessionResponse

	sess := getRequestAuthSession(ctx, h.auth)

	documentName := request.Body.Attributes.DocumentName
	_, accessErr := h.verifyUserDocumentAccess(ctx, sess.UserId, documentName)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	token, tokenErr := h.auth.IssueAuthSessionToken(sess)
	if tokenErr != nil {
		return nil, apiError("failed to create auth token", tokenErr)
	}

	resp.Body.Data = oapi.DocumentEditorSession{
		DocumentName:  documentName,
		Token:         token,
		ConnectionUrl: h.documents.GetServerWebsocketAddress(),
	}

	return &resp, nil
}

func (h *documentsHandler) VerifyDocumentEditorSession(ctx context.Context, request *oapi.VerifyDocumentEditorSessionRequest) (*oapi.VerifyDocumentEditorSessionResponse, error) {
	var resp oapi.VerifyDocumentEditorSessionResponse

	userId := requestUserId(ctx, h.auth)

	user, userErr := h.users.GetById(ctx, userId)
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}

	documentName := request.Body.Attributes.DocumentName

	readOnly, accessErr := h.verifyUserDocumentAccess(ctx, userId, documentName)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	resp.Body.Data = oapi.DocumentEditorSessionAuth{
		User: oapi.DocumentEditorSessionUser{
			Id:       userId,
			Username: user.Name,
		},
		ReadOnly: readOnly,
	}

	return &resp, nil
}
