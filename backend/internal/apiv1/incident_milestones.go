package apiv1

import (
	"context"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentmilestone"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentMilestonesHandler struct {
	db *ent.Client
}

func newIncidentMilestonesHandler(db *ent.Client) *incidentMilestonesHandler {
	return &incidentMilestonesHandler{db: db}
}

func (h *incidentMilestonesHandler) ListIncidentMilestones(ctx context.Context, request *oapi.ListIncidentMilestonesRequest) (*oapi.ListIncidentMilestonesResponse, error) {
	var resp oapi.ListIncidentMilestonesResponse

	query := h.db.IncidentMilestone.Query()

	query.Limit(10)
	query.Offset(0)

	results, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query incident events", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentMilestone, len(results))
	for i, ev := range results {
		resp.Body.Data[i] = oapi.IncidentMilestoneFromEnt(ev)
	}

	return &resp, nil
}

func (h *incidentMilestonesHandler) CreateIncidentMilestone(ctx context.Context, request *oapi.CreateIncidentMilestoneRequest) (*oapi.CreateIncidentMilestoneResponse, error) {
	var resp oapi.CreateIncidentMilestoneResponse

	attrs := request.Body.Attributes
	query := h.db.IncidentMilestone.Create().
		SetIncidentID(request.Id).
		SetKind(incidentmilestone.Kind(attrs.Kind)).
		SetTime(attrs.Timestamp)

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, apiError("failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) UpdateIncidentMilestone(ctx context.Context, input *oapi.UpdateIncidentMilestoneRequest) (*oapi.UpdateIncidentMilestoneResponse, error) {
	var resp oapi.UpdateIncidentMilestoneResponse

	attrs := input.Body.Attributes

	query := h.db.IncidentMilestone.UpdateOneID(input.Id).
		SetNillableDescription(attrs.Description).
		SetNillableTime(attrs.Timestamp)

	if attrs.Kind != nil {
		query.SetKind(incidentmilestone.Kind(*attrs.Kind))
	}

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, apiError("failed to update incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) DeleteIncidentMilestone(ctx context.Context, input *oapi.DeleteIncidentMilestoneRequest) (*oapi.DeleteIncidentMilestoneResponse, error) {
	var resp oapi.DeleteIncidentMilestoneResponse

	deleteErr := h.db.IncidentMilestone.DeleteOneID(input.Id).Exec(ctx)
	if deleteErr != nil {
		return nil, apiError("failed to archive incident event", deleteErr)
	}

	return &resp, nil
}
