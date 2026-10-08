package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type documentsHandler struct {
	*baseHandler
	documents rez.DocumentsService
}

func newDocumentsHandler(bh *baseHandler, docs rez.DocumentsService) *documentsHandler {
	return &documentsHandler{bh, docs}
}

func (h *documentsHandler) RequestDocumentSessionAuth(ctx context.Context, req *oapi.RequestDocumentSessionAuthRequest) (*oapi.RequestDocumentSessionAuthResponse, error) {
	var resp oapi.RequestDocumentSessionAuthResponse

	ds, dsErr := h.documents.CreateDocumentEditorSessionAuth(ctx, req.Id, h.mustUserID(ctx))
	if dsErr != nil {
		return nil, oapi.Error(ctx, "create session", dsErr)
	}
	resp.Body.Data = oapi.DocumentSessionAuthFromRez(ds)

	return &resp, nil
}

func (h *documentsHandler) GetDocumentSession(ctx context.Context, request *oapi.GetDocumentSessionRequest) (*oapi.GetDocumentSessionResponse, error) {
	var resp oapi.GetDocumentSessionResponse

	usr, usrErr := h.currentUser(ctx)
	if usrErr != nil {
		return nil, oapi.Error(ctx, "get user", usrErr)
	}
	docAccess, docErr := h.documents.GetUserDocumentAccess(ctx, request.Id, usr.ID)
	if docErr != nil {
		return nil, oapi.Error(ctx, "get access", docErr)
	}
	if docAccess == nil {
		return nil, oapi.Error(ctx, "get access", errs.ErrForbidden)
	}
	resp.Body.Data = oapi.DocumentSession{
		User:   oapi.UserFromEnt(usr),
		Access: oapi.DocumentAccessFromEnt(docAccess),
	}

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
