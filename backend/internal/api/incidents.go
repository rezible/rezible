package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentsHandler struct {
	incidents *ent.IncidentClient
	db        *ent.Client
}

func newIncidentsHandler(db *ent.Client) *incidentsHandler {
	return &incidentsHandler{
		db.Incident,
		db,
	}
}

func (h *incidentsHandler) CreateIncident(ctx context.Context, input *oapi.CreateIncidentRequest) (*oapi.CreateIncidentResponse, error) {
	var resp oapi.CreateIncidentResponse

	attr := input.Body.Attributes

	slug := "foo-incident" // TODO

	query := h.incidents.Create().
		SetTitle(attr.Title).
		SetSlug(slug).
		SetSummary(attr.Summary)

	inc, err := query.Save(ctx)
	if err != nil {
		return nil, detailError("failed to create incident", err)
	}
	resp.Body.Data = oapi.IncidentFromEnt(inc)

	return &resp, nil
}

func (h *incidentsHandler) ListIncidents(ctx context.Context, input *oapi.ListIncidentsRequest) (*oapi.ListIncidentsResponse, error) {
	var resp oapi.ListIncidentsResponse

	query := h.incidents.Query()
	// etc

	incidents, err := query.All(ctx)
	if err != nil {
		return nil, detailError("failed to list incidents", err)
	}

	resp.Body.Data = make([]oapi.Incident, len(incidents))
	for i, inc := range incidents {
		resp.Body.Data[i] = oapi.IncidentFromEnt(inc)
	}

	return &resp, nil
}

func (h *incidentsHandler) GetIncident(ctx context.Context, input *oapi.GetIncidentRequest) (*oapi.GetIncidentResponse, error) {
	var resp oapi.GetIncidentResponse

	idPredicate := oapi.GetEntPredicate(input.Id, incident.ID, incident.Slug)

	// TODO: use a view for this
	query := h.incidents.Query().
		Where(idPredicate).
		WithSeverity().
		WithType().
		WithFieldSelections().
		WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.WithRole().WithUser()
		}).
		WithTeamAssignments(func(q *ent.IncidentTeamAssignmentQuery) {
			q.WithTeam()
		}).
		WithRetrospective()

	inc, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, detailError("failed to get incident", queryErr)
	}

	resp.Body.Data = oapi.IncidentFromEnt(inc)

	return &resp, nil
}

func (h *incidentsHandler) ArchiveIncident(ctx context.Context, input *oapi.ArchiveIncidentRequest) (*oapi.ArchiveIncidentResponse, error) {
	var resp oapi.ArchiveIncidentResponse

	err := h.incidents.DeleteOneID(input.Id).Exec(ctx)
	if err != nil {
		return nil, detailError("failed to archive incident", err)
	}

	return &resp, nil
}

func (h *incidentsHandler) UpdateIncident(ctx context.Context, request *oapi.UpdateIncidentRequest) (*oapi.UpdateIncidentResponse, error) {
	var resp oapi.UpdateIncidentResponse

	inc, getErr := h.incidents.Query().
		Where(incident.ID(request.Id)).
		Only(ctx)
	if getErr != nil {
		return nil, detailError("failed to get incident", getErr)
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
		return nil, detailError("failed to update incident", txErr)
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
