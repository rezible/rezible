package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	oapi "github.com/rezible/rezible/openapi"
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

	slug := "foo-incident" // TODO

	query := h.db.Incident.Create().
		SetTitle(attr.Title).
		SetSlug(slug).
		SetSummary(attr.Summary)

	inc, err := query.Save(ctx)
	if err != nil {
		return nil, apiError("failed to create incident", err)
	}
	resp.Body.Data = oapi.IncidentFromEnt(inc)

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

	inc, getErr := h.db.Incident.Query().
		Where(incident.ID(request.Id)).
		Only(ctx)
	if getErr != nil {
		return nil, apiError("failed to get incident", getErr)
	}

	updateIncidentTxFn := func(tx *ent.Tx) error {
		updatedIncident, updateErr := h.updateIncident(ctx, tx, inc, request.Body.Attributes)
		if updateErr != nil {
			return fmt.Errorf("failed to update incident: %w", updateErr)
		}
		resp.Body.Data = oapi.IncidentFromEnt(updatedIncident)
		return nil
	}

	if txErr := ent.WithTx(ctx, h.db, updateIncidentTxFn); txErr != nil {
		return nil, apiError("failed to update incident", txErr)
	}

	return &resp, nil
}

func (h *incidentsHandler) updateIncident(ctx context.Context, tx *ent.Tx, inc *ent.Incident, attr oapi.UpdateIncidentAttributes) (
	*ent.Incident, error,
) {
	update := tx.Incident.UpdateOneID(inc.ID).
		SetNillableTitle(attr.Title).
		SetNillableSummary(attr.Summary).
		SetNillablePrivate(attr.Private)

	if attr.SeverityId != nil {
		sevId, sevErr := uuid.Parse(*attr.SeverityId)
		if sevErr != nil {
			return nil, oapi.ErrorBadRequest("invalid severity id", sevErr)
		}
		update.SetSeverityID(sevId)
	}

	return update.Save(ctx)
}
