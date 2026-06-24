package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/incidentmilestone"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type incidentMilestonesHandler struct {
	db rez.Database
}

func newIncidentMilestonesHandler(db rez.Database) *incidentMilestonesHandler {
	return &incidentMilestonesHandler{db: db}
}

func (h *incidentMilestonesHandler) ListIncidentMilestones(ctx context.Context, request *oapi.ListIncidentMilestonesRequest) (*oapi.ListIncidentMilestonesResponse, error) {
	var resp oapi.ListIncidentMilestonesResponse

	query := h.db.Client(ctx).IncidentMilestone.Query()

	query.Limit(10)
	query.Offset(0)

	results, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query incident events", queryErr)
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
	query := h.db.Client(ctx).IncidentMilestone.Create().
		SetIncidentID(request.Id).
		SetKind(incidentmilestone.Kind(attrs.Kind)).
		SetTimestamp(attrs.Timestamp)

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) UpdateIncidentMilestone(ctx context.Context, input *oapi.UpdateIncidentMilestoneRequest) (*oapi.UpdateIncidentMilestoneResponse, error) {
	var resp oapi.UpdateIncidentMilestoneResponse

	attrs := input.Body.Attributes

	query := h.db.Client(ctx).IncidentMilestone.UpdateOneID(input.Id).
		SetNillableDescription(attrs.Description).
		SetNillableTimestamp(attrs.Timestamp)

	if attrs.Kind != nil {
		query.SetKind(incidentmilestone.Kind(*attrs.Kind))
	}

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to update incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentMilestoneFromEnt(ev)

	return &resp, nil
}

func (h *incidentMilestonesHandler) DeleteIncidentMilestone(ctx context.Context, input *oapi.DeleteIncidentMilestoneRequest) (*oapi.DeleteIncidentMilestoneResponse, error) {
	var resp oapi.DeleteIncidentMilestoneResponse

	deleteErr := h.db.Client(ctx).IncidentMilestone.DeleteOneID(input.Id).Exec(ctx)
	if deleteErr != nil {
		return nil, oapi.Error(ctx, "failed to archive incident event", deleteErr)
	}

	return &resp, nil
}
