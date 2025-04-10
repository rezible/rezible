package api

import (
	"context"
	"errors"
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

func (h *documentsHandler) RequestDocumentEditorSession(ctx context.Context, request *oapi.RequestDocumentEditorSessionRequest) (*oapi.RequestDocumentEditorSessionResponse, error) {
	var resp oapi.RequestDocumentEditorSessionResponse

	sess, authErr := h.auth.GetSession(ctx)
	if authErr != nil {
		return nil, detailError("failed to get auth session", authErr)
	}

	documentName := request.Body.Attributes.DocumentName
	_, accessErr := h.documents.CheckUserDocumentAccess(ctx, sess.UserId, documentName)
	if accessErr != nil {
		return nil, detailError("no document access", accessErr)
	}

	token, tokenErr := h.auth.IssueSessionToken(sess)
	if tokenErr != nil {
		return nil, detailError("failed to create auth token", tokenErr)
	}

	resp.Body.Data = oapi.DocumentEditorSession{
		DocumentName:  documentName,
		Token:         token,
		ConnectionUrl: h.documents.GetWebsocketAddress(),
	}

	return &resp, nil
}

func (h *documentsHandler) VerifyDocumentEditorSession(ctx context.Context, request *oapi.VerifyDocumentEditorSessionRequest) (*oapi.VerifyDocumentEditorSessionResponse, error) {
	var resp oapi.VerifyDocumentEditorSessionResponse

	sess, authErr := h.auth.GetSession(ctx)
	if authErr != nil {
		return nil, detailError("failed to get auth session", authErr)
	}

	user, userErr := h.users.GetById(ctx, sess.UserId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}

	documentName := request.Body.Attributes.DocumentName

	readOnly, accessErr := h.documents.CheckUserDocumentAccess(ctx, sess.UserId, documentName)
	if accessErr != nil {
		if errors.Is(accessErr, rez.ErrUnauthorized) {
			return nil, oapi.ErrorUnauthorized("no access to document")
		}
		return nil, detailError("check user document access", accessErr)
	}

	resp.Body.Data = oapi.DocumentEditorSessionAuth{
		User: oapi.DocumentEditorSessionUser{
			Id:       sess.UserId,
			Username: user.Name,
		},
		ReadOnly: readOnly,
	}

	return &resp, nil
}
