package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type documentsHandler struct {
	documents rez.DocumentsService
}

func newDocumentsHandler(documents rez.DocumentsService) *documentsHandler {
	return &documentsHandler{documents}
}

func (h *documentsHandler) RequestDocumentSession(ctx context.Context, req *oapi.RequestDocumentSessionRequest) (*oapi.RequestDocumentSessionResponse, error) {
	var resp oapi.RequestDocumentSessionResponse

	userId, userOK := execution.GetContext(ctx).UserID()
	if !userOK {
		return nil, oapi.Error(ctx, "no user", rez.ErrAuthSessionMissing)
	}
	ds, dsErr := h.documents.CreateDocumentEditorSession(ctx, req.Id, userId)
	if dsErr != nil {
		return nil, oapi.Error(ctx, "create session", dsErr)
	}
	resp.Body.Data = oapi.DocumentSessionFromRez(ds)

	return &resp, nil
}

func (h *documentsHandler) GetDocumentAccess(ctx context.Context, request *oapi.GetDocumentAccessRequest) (*oapi.GetDocumentAccessResponse, error) {
	var resp oapi.GetDocumentAccessResponse

	userId, userOK := execution.GetContext(ctx).UserID()
	if !userOK {
		return nil, oapi.Error(ctx, "no user", rez.ErrAuthSessionMissing)
	}
	docAccess, docErr := h.documents.GetUserDocumentAccess(ctx, request.Id, userId)
	if docErr != nil {
		return nil, oapi.Error(ctx, "get access", docErr)
	}
	if docAccess == nil {
		return nil, oapi.ErrForbidden
	}
	resp.Body.Data = oapi.DocumentAccessFromEnt(docAccess)

	return &resp, nil
}

func (h *documentsHandler) LoadDocument(ctx context.Context, req *oapi.LoadDocumentRequest) (*oapi.LoadDocumentResponse, error) {
	var resp oapi.LoadDocumentResponse

	doc, docErr := h.documents.GetDocument(ctx, req.Id)
	if docErr != nil {
		return nil, oapi.Error(ctx, "failed to load document", docErr)
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
		return nil, oapi.Error(ctx, "failed to update document", docErr)
	}
	resp.Body.Data = oapi.DocumentFromEnt(doc)

	return &resp, nil
}
