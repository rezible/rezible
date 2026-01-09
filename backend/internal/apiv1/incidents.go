package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentsHandler struct {
	db        *ent.Client
	incidents rez.IncidentService
}

func newIncidentsHandler(db *ent.Client, incidents rez.IncidentService) *incidentsHandler {
	return &incidentsHandler{
		db:        db,
		incidents: incidents,
	}
}

func (h *incidentsHandler) CreateIncident(ctx context.Context, input *oapi.CreateIncidentRequest) (*oapi.CreateIncidentResponse, error) {
	var resp oapi.CreateIncidentResponse

	attr := input.Body.Attributes

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetTitle(attr.Title)
		m.SetSummary(attr.Summary)
		return nil
	}

	created, createErr := h.incidents.Set(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, apiError("failed to create incident", createErr)
	}
	resp.Body.Data = oapi.IncidentFromEnt(created)

	return &resp, nil
}

func (h *incidentsHandler) ListIncidents(ctx context.Context, req *oapi.ListIncidentsRequest) (*oapi.ListIncidentsResponse, error) {
	var resp oapi.ListIncidentsResponse

	params := rez.ListIncidentsParams{
		ListParams: req.ListParams(),
	}
	listRes, listErr := h.incidents.ListIncidents(ctx, params)
	// etc
	if listErr != nil {
		return nil, apiError("failed to list incidents", listErr)
	}
	resp.Body.Data = make([]oapi.Incident, len(listRes.Data))
	for i, inc := range listRes.Data {
		resp.Body.Data[i] = oapi.IncidentFromEnt(inc)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: listRes.Count,
	}

	return &resp, nil
}

func (h *incidentsHandler) GetIncident(ctx context.Context, input *oapi.GetIncidentRequest) (*oapi.GetIncidentResponse, error) {
	var resp oapi.GetIncidentResponse

	var inc *ent.Incident
	var incErr error
	if input.Id.IsSlug {
		inc, incErr = h.incidents.GetBySlug(ctx, input.Id.Slug)
	} else {
		inc, incErr = h.incidents.Get(ctx, input.Id.UUID)
	}
	if incErr != nil {
		return nil, apiError("failed to get incident", incErr)
	}

	resp.Body.Data = oapi.IncidentFromEnt(inc)

	return &resp, nil
}

func (h *incidentsHandler) ArchiveIncident(ctx context.Context, input *oapi.ArchiveIncidentRequest) (*oapi.ArchiveIncidentResponse, error) {
	var resp oapi.ArchiveIncidentResponse

	err := h.db.Incident.DeleteOneID(input.Id).Exec(ctx)
	if err != nil {
		return nil, apiError("failed to archive incident", err)
	}

	return &resp, nil
}

func (h *incidentsHandler) UpdateIncident(ctx context.Context, request *oapi.UpdateIncidentRequest) (*oapi.UpdateIncidentResponse, error) {
	var resp oapi.UpdateIncidentResponse

	inc, getErr := h.incidents.Get(ctx, request.Id)
	if getErr != nil {
		return nil, apiError("failed to get incident", getErr)
	}

	attr := request.Body.Attributes

	var updateSeverityId uuid.UUID
	if attr.SeverityId != nil {
		sevId, sevErr := uuid.Parse(*attr.SeverityId)
		if sevErr != nil {
			return nil, oapi.ErrorBadRequest("invalid severity id", sevErr)
		}
		updateSeverityId = sevId
	}

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		if attr.Title != nil {
			m.SetTitle(*attr.Title)
		}
		if attr.Summary != nil {
			m.SetSummary(*attr.Summary)
		}
		if updateSeverityId != uuid.Nil {
			m.SetSeverityID(updateSeverityId)
		}

		return nil
	}

	updated, updateErr := h.incidents.Set(ctx, inc.ID, setFn)
	if updateErr != nil {
		return nil, apiError("failed to update incident", updateErr)
	}
	resp.Body.Data = oapi.IncidentFromEnt(updated)

	return &resp, nil
}
