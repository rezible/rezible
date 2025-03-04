package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentmilestone"
	oapi "github.com/rezible/rezible/openapi"
	"time"
)

type incidentMilestonesHandler struct {
	events *ent.IncidentMilestoneClient
}

func newIncidentMilestonesHandler(events *ent.IncidentMilestoneClient) *incidentMilestonesHandler {
	return &incidentMilestonesHandler{events}
}

func (h *incidentMilestonesHandler) ListIncidentMilestones(ctx context.Context, request *oapi.ListIncidentMilestonesRequest) (*oapi.ListIncidentMilestonesResponse, error) {
	var resp oapi.ListIncidentMilestonesResponse

	query := h.events.Query()

	query.Limit(10)
	query.Offset(0)

	results, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query incident events", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentMilestone, len(results)+1)
	for i, ev := range results {
		resp.Body.Data[i] = oapi.IncidentMilestoneFromEnt(ev)
	}
	resp.Body.Data[len(resp.Body.Data)-1] = oapi.IncidentMilestone{
		Id: uuid.New(),
		Attributes: oapi.IncidentMilestoneAttributes{
			IncidentId: request.Id,
			Type:       "mitigated",
			Title:      "Incident Mitigated",
			Timestamp:  time.Now(),
		},
	}

	return &resp, nil
}

func (h *incidentMilestonesHandler) CreateIncidentMilestone(ctx context.Context, input *oapi.CreateIncidentMilestoneRequest) (*oapi.CreateIncidentMilestoneResponse, error) {
	var resp oapi.CreateIncidentMilestoneResponse

	attrs := input.Body.Attributes
	query := h.events.Create().
		SetType(incidentmilestone.Type(attrs.Type))

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) UpdateIncidentMilestone(ctx context.Context, input *oapi.UpdateIncidentMilestoneRequest) (*oapi.UpdateIncidentMilestoneResponse, error) {
	var resp oapi.UpdateIncidentMilestoneResponse

	attrs := input.Body.Attributes

	query := h.events.UpdateOneID(input.Id)

	if attrs.Type != nil {
		query.SetType(incidentmilestone.Type(*attrs.Type))
	}

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to update incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) ArchiveIncidentMilestone(ctx context.Context, input *oapi.ArchiveIncidentMilestoneRequest) (*oapi.ArchiveIncidentMilestoneResponse, error) {
	var resp oapi.ArchiveIncidentMilestoneResponse

	deleteErr := h.events.DeleteOneID(input.Id).Exec(ctx)
	if deleteErr != nil {
		return nil, detailError("failed to archive incident event", deleteErr)
	}

	return &resp, nil
}
