package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type documentsHandler struct {
	documents rez.DocumentsService
	auth      rez.AuthSessionService
}

func newDocumentsHandler(documents rez.DocumentsService, auth rez.AuthSessionService) *documentsHandler {
	return &documentsHandler{documents, auth}
}

func (h *documentsHandler) GetDocumentAccess(ctx context.Context, request *oapi.GetDocumentAccessRequest) (*oapi.GetDocumentAccessResponse, error) {
	var resp oapi.GetDocumentAccessResponse

	docAccess, docErr := h.documents.GetDocumentAccess(ctx, request.Id)
	if docErr != nil {
		return nil, oapi.Error("get access", docErr)
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
		return nil, oapi.Error("failed to load document", docErr)
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
		return nil, oapi.Error("failed to update document", docErr)
	}
	resp.Body.Data = oapi.DocumentFromEnt(doc)

	return &resp, nil
}
