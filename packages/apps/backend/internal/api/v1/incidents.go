package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type incidentsHandler struct {
	*baseHandler
	db        rez.Database
	incidents rez.IncidentService
}

func newIncidentsHandler(base *baseHandler, db rez.Database, incidents rez.IncidentService) *incidentsHandler {
	return &incidentsHandler{baseHandler: base, db: db, incidents: incidents}
}

func incidentIdPredicate(id oapi.FlexibleId) predicate.Incident {
	if id.IsSlug {
		return incident.Slug(id.Slug)
	}
	return incident.ID(id.UUID)
}

func (h *incidentsHandler) ListIncidents(ctx context.Context, req *oapi.ListIncidentsRequest) (*oapi.ListIncidentsResponse, error) {
	var resp oapi.ListIncidentsResponse

	params := rez.ListIncidentsParams{
		ListParams:     req.ListParams(),
		ResponseStates: req.ResponseStates,
		SeverityId:     req.SeverityId,
	}
	params.Search = req.Search
	incs, listErr := h.incidents.ListIncidents(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list incidents", listErr)
	}
	resp.Body = oapi.ConvertPaginatedResultBody(incs, oapi.IncidentFromEnt)

	return &resp, nil
}

func (h *incidentsHandler) GetIncident(ctx context.Context, input *oapi.GetIncidentRequest) (*oapi.GetIncidentResponse, error) {
	var resp oapi.GetIncidentResponse

	inc, incErr := h.incidents.Get(ctx, incidentIdPredicate(input.Id))
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	resp.Body.Data = oapi.IncidentFromEnt(inc)

	return &resp, nil
}

func (h *incidentsHandler) CreateIncident(ctx context.Context, input *oapi.CreateIncidentRequest) (*oapi.CreateIncidentResponse, error) {
	return nil, oapi.Error(ctx, "create incident is not implemented", errs.ErrNotImplemented)
}

func (h *incidentsHandler) UpdateIncident(ctx context.Context, request *oapi.UpdateIncidentRequest) (*oapi.UpdateIncidentResponse, error) {
	return nil, oapi.Error(ctx, "incident is read-only", errs.ErrNotImplemented)
}

func (h *incidentsHandler) ArchiveIncident(ctx context.Context, input *oapi.ArchiveIncidentRequest) (*oapi.ArchiveIncidentResponse, error) {
	var resp oapi.ArchiveIncidentResponse

	if archiveErr := h.incidents.Archive(ctx, input.Id); archiveErr != nil {
		return nil, oapi.Error(ctx, "archive incident", archiveErr)
	}

	return &resp, nil
}

func (h *incidentsHandler) ListIncidentMilestones(ctx context.Context, request *oapi.ListIncidentMilestonesRequest) (*oapi.ListIncidentMilestonesResponse, error) {
	milestones, milestonesErr := h.incidents.ListMilestonesForIncident(ctx, request.Id)
	if milestonesErr != nil {
		return nil, oapi.Error(ctx, "failed to query incident events", milestonesErr)
	}
	var resp oapi.ListIncidentMilestonesResponse
	resp.Body.Data = oapi.ConvertSlice(milestones, oapi.IncidentMilestoneFromEnt)
	return &resp, nil
}

func (h *incidentsHandler) CreateIncidentRoleAssignment(ctx context.Context, request *oapi.CreateIncidentRoleAssignmentRequest) (*oapi.CreateIncidentRoleAssignmentResponse, error) {
	attr := request.Body.Attributes
	params := rez.SetIncidentRoleAssignmentParams{
		IncidentID: request.Id,
		RoleID:     attr.RoleId,
		UserID:     attr.UserId,
	}
	assignment, saveErr := h.incidents.SetIncidentRoleAssignment(ctx, uuid.Nil, params)
	if saveErr != nil {
		return nil, oapi.Error(ctx, "assign incident role", saveErr)
	}
	var resp oapi.CreateIncidentRoleAssignmentResponse
	resp.Body.Data = oapi.IncidentUserRoleAssignmentFromEnt(assignment)
	return &resp, nil
}

func (h *incidentsHandler) DeleteIncidentRoleAssignment(ctx context.Context, request *oapi.DeleteIncidentRoleAssignmentRequest) (*oapi.DeleteIncidentRoleAssignmentResponse, error) {
	if deleteErr := h.incidents.DeleteIncidentRoleAssignment(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete incident role assignment", deleteErr)
	}
	return &oapi.DeleteIncidentRoleAssignmentResponse{}, nil
}

func (*incidentsHandler) LinkIncidentSituation(ctx context.Context, _ *oapi.LinkIncidentSituationRequest) (*oapi.LinkIncidentSituationResponse, error) {
	return nil, oapi.Error(ctx, "incident situation links are not implemented", errs.ErrNotImplemented)
}

func (*incidentsHandler) UnlinkIncidentSituation(ctx context.Context, _ *oapi.UnlinkIncidentSituationRequest) (*oapi.UnlinkIncidentSituationResponse, error) {
	return nil, oapi.Error(ctx, "incident situation links are not implemented", errs.ErrNotImplemented)
}

func (h *incidentsHandler) ListIncidentUpdates(ctx context.Context, req *oapi.ListIncidentUpdatesRequest) (*oapi.ListIncidentUpdatesResponse, error) {
	if _, getErr := h.incidents.Get(ctx, incident.ID(req.Id)); getErr != nil {
		return nil, oapi.Error(ctx, "get incident for updates", getErr)
	}
	resp := &oapi.ListIncidentUpdatesResponse{}
	resp.Body.Data = make([]oapi.IncidentUpdate, 0)
	resp.Body.Pagination = oapi.Pagination{Page: 1, PageSize: req.PageSize, Total: 0}
	return resp, nil
}

func (*incidentsHandler) CreateIncidentUpdate(ctx context.Context, _ *oapi.CreateIncidentUpdateRequest) (*oapi.CreateIncidentUpdateResponse, error) {
	return nil, oapi.Error(ctx, "incident updates are not implemented", errs.ErrNotImplemented)
}
