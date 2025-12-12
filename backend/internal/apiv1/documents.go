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

	sess := getRequestAuthSession(ctx, h.auth)

	docId := request.Id
	_, accessErr := h.documents.GetUserDocumentAccess(ctx, sess.UserId, docId)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	token, tokenErr := h.documents.CreateEditorSessionToken(sess, docId)
	if tokenErr != nil {
		return nil, apiError("failed to create session token", tokenErr)
	}

	wsUrl := "ws://" + rez.Config.DocumentsServerAddress()
	resp.Body.Data = oapi.DocumentEditorSession{
		DocumentId:    docId,
		Token:         token,
		ConnectionUrl: wsUrl,
	}

	return &resp, nil
}

func (h *documentsHandler) VerifyDocumentSessionAuth(ctx context.Context, req *oapi.VerifyDocumentSessionAuthRequest) (*oapi.VerifyDocumentSessionAuthResponse, error) {
	var resp oapi.VerifyDocumentSessionAuthResponse

	sess := getRequestAuthSession(ctx, h.auth)

	docId := req.Id
	readOnly, accessErr := h.documents.GetUserDocumentAccess(ctx, sess.UserId, docId)
	if accessErr != nil {
		return nil, apiError("no document access", accessErr)
	}

	resp.Body.Data = oapi.DocumentEditorSessionAuth{
		User: oapi.DocumentEditorSessionUser{
			Id:       sess.UserId,
			Username: "document-user",
		},
		ReadOnly: readOnly,
	}

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
