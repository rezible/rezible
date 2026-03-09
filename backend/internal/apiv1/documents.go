package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type documentsHandler struct {
	documents rez.DocumentsService
	auth      rez.AuthService
}

func newDocumentsHandler(documents rez.DocumentsService, auth rez.AuthService) *documentsHandler {
	return &documentsHandler{documents, auth}
}

func (h *documentsHandler) RequestDocumentEditorSession(ctx context.Context, request *oapi.RequestDocumentEditorSessionRequest) (*oapi.RequestDocumentEditorSessionResponse, error) {
	var resp oapi.RequestDocumentEditorSessionResponse

	docId := request.Id
	accessToken, accessErr := h.documents.CreateDocumentAccessAuthSessionToken(ctx, docId, getRequestAuthSession(ctx, h.auth))
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	wsUrl := "ws://" + rez.Config.DocumentsServerAddress()
	resp.Body.Data = oapi.DocumentEditorSession{
		DocumentId:    docId,
		AccessToken:   accessToken,
		ConnectionUrl: wsUrl,
	}

	return &resp, nil
}

func (h *documentsHandler) VerifyDocumentSessionAuth(ctx context.Context, req *oapi.VerifyDocumentSessionAuthRequest) (*oapi.VerifyDocumentSessionAuthResponse, error) {
	var resp oapi.VerifyDocumentSessionAuthResponse

	sess := getRequestAuthSession(ctx, h.auth)
	access, accessErr := h.documents.GetDocumentAccess(ctx, req.Id, sess)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}
	if access == nil {
		return nil, apiError("no document access", rez.ErrAuthSessionUnauthorized)
	}
	access.UserID = sess.UserId
	resp.Body.Data = oapi.DocumentEditorSessionAuthFromEnt(access)
	return &resp, nil
}

func (h *documentsHandler) LoadDocument(ctx context.Context, req *oapi.LoadDocumentRequest) (*oapi.LoadDocumentResponse, error) {
	var resp oapi.LoadDocumentResponse

	doc, docErr := h.documents.GetDocument(ctx, req.Id)
	if docErr != nil {
		return nil, apiError("failed to load document", docErr)
	}
	resp.Body.Data = oapi.DocumentFromEnt(doc)

	return &resp, nil
}

func (h *documentsHandler) UpdateDocument(ctx context.Context, req *oapi.UpdateDocumentRequest) (*oapi.UpdateDocumentResponse, error) {
	var resp oapi.UpdateDocumentResponse

	attr := req.Body.Attributes
	updateFn := func(m *ent.DocumentMutation) {
		m.SetContent(attr.Content)
	}
	doc, docErr := h.documents.SetDocument(ctx, req.Id, updateFn)
	if docErr != nil {
		return nil, apiError("failed to update document", docErr)
	}
	resp.Body.Data = oapi.DocumentFromEnt(doc)

	return &resp, nil
}
