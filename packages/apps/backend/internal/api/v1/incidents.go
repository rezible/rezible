package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type incidentsHandler struct {
	*baseHandler
	incidents  rez.IncidentDetailsService
	management rez.IncidentManagementService
}

func newIncidentsHandler(base *baseHandler, incidents rez.IncidentDetailsService, management rez.IncidentManagementService) *incidentsHandler {
	return &incidentsHandler{baseHandler: base, incidents: incidents, management: management}
}

func incidentIdPredicate(id oapi.FlexibleId) (predicate.Incident, error) {
	if !id.IsKey {
		return incident.ID(id.UUID), nil
	}
	number, isReference := ent.ParseIncidentReference(id.Key)
	if !isReference {
		return nil, errs.ErrNotFound
	}
	return incident.Number(number), nil
}

func (h *incidentsHandler) ListIncidents(ctx context.Context, req *oapi.ListIncidentsRequest) (*oapi.ListIncidentsResponse, error) {
	var resp oapi.ListIncidentsResponse

	params := rez.ListIncidentsParams{
		ListParams: req.ListParams(),
		SeverityId: req.SeverityId,
	}
	if req.Resolved.IsSet {
		params.Resolved = &req.Resolved.Value
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

	pred, predErr := incidentIdPredicate(input.Id)
	if predErr != nil {
		return nil, oapi.Error(ctx, "get incident", predErr)
	}
	inc, incErr := h.incidents.GetIncident(ctx, pred)
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	resp.Body.Data = oapi.IncidentFromEnt(inc)

	return &resp, nil
}

func (h *incidentsHandler) CreateIncident(ctx context.Context, input *oapi.CreateIncidentRequest) (*oapi.CreateIncidentResponse, error) {
	userID, hasUser := execution.GetContext(ctx).UserID()
	if !hasUser {
		return nil, oapi.Error(ctx, "declare incident", errs.ErrAuthSessionMissing)
	}
	attr := input.Body.Attributes
	params := rez.DeclareIncidentParams{
		Title:          attr.Title,
		SeverityID:     attr.SeverityId,
		TypeID:         attr.TypeId,
		FieldOptionIDs: attr.FieldSelectionIds,
		TagIDs:         attr.TagIds,
		SituationIDs:   attr.SituationIds,
		DeclaredByID:   userID,
	}
	if attr.Summary != nil {
		params.Summary = *attr.Summary
	}
	inc, declareErr := h.management.Declare(ctx, params)
	if declareErr != nil {
		return nil, oapi.Error(ctx, "declare incident", declareErr)
	}
	var resp oapi.CreateIncidentResponse
	resp.Body.Data = oapi.CreatedIncident{Id: inc.ID, Reference: inc.Reference()}
	return &resp, nil
}

func (h *incidentsHandler) UpdateIncident(ctx context.Context, request *oapi.UpdateIncidentRequest) (*oapi.UpdateIncidentResponse, error) {
	attr := request.Body.Attributes
	params := rez.UpdateIncidentParams{
		Title:          attr.Title,
		Summary:        attr.Summary,
		SeverityID:     attr.SeverityId.NillableValue(),
		ClearSeverity:  attr.SeverityId.Sent && attr.SeverityId.Null,
		TypeID:         attr.TypeId.NillableValue(),
		ClearType:      attr.TypeId.Sent && attr.TypeId.Null,
		FieldOptionIDs: attr.FieldSelectionIds,
		TagIDs:         attr.TagIds,
	}
	if _, updateErr := h.incidents.UpdateIncident(ctx, request.Id, params); updateErr != nil {
		return nil, oapi.Error(ctx, "update incident", updateErr)
	}
	return &oapi.UpdateIncidentResponse{}, nil
}

func (h *incidentsHandler) ArchiveIncident(ctx context.Context, input *oapi.ArchiveIncidentRequest) (*oapi.ArchiveIncidentResponse, error) {
	var resp oapi.ArchiveIncidentResponse

	if archiveErr := h.incidents.ArchiveIncident(ctx, input.Id); archiveErr != nil {
		return nil, oapi.Error(ctx, "archive incident", archiveErr)
	}

	return &resp, nil
}

func (h *incidentsHandler) ListIncidentMilestones(ctx context.Context, request *oapi.ListIncidentMilestonesRequest) (*oapi.ListIncidentMilestonesResponse, error) {
	milestones, milestonesErr := h.incidents.ListMilestones(ctx, request.Id)
	if milestonesErr != nil {
		return nil, oapi.Error(ctx, "failed to query incident events", milestonesErr)
	}
	var resp oapi.ListIncidentMilestonesResponse
	resp.Body.Data = oapi.ConvertSlice(milestones, oapi.IncidentMilestoneFromEnt)
	return &resp, nil
}

func (h *incidentsHandler) CreateIncidentMilestone(ctx context.Context, request *oapi.CreateIncidentMilestoneRequest) (*oapi.CreateIncidentMilestoneResponse, error) {
	userID, hasUser := execution.GetContext(ctx).UserID()
	if !hasUser {
		return nil, oapi.Error(ctx, "record incident milestone", errs.ErrAuthSessionMissing)
	}
	attr := request.Body.Attributes
	params := rez.AddIncidentMilestoneParams{
		IncidentID: request.Id,
		Kind:       incidentmilestone.Kind(attr.Kind),
		UserID:     &userID,
		Source:     "web",
	}
	if attr.Timestamp != nil {
		params.Timestamp = *attr.Timestamp
	}
	if attr.Description != nil {
		params.Description = *attr.Description
	}
	if _, addErr := h.management.AddMilestone(ctx, params); addErr != nil {
		return nil, oapi.Error(ctx, "record incident milestone", addErr)
	}
	return &oapi.CreateIncidentMilestoneResponse{}, nil
}

func (h *incidentsHandler) UpdateIncidentMilestone(ctx context.Context, request *oapi.UpdateIncidentMilestoneRequest) (*oapi.UpdateIncidentMilestoneResponse, error) {
	attr := request.Body.Attributes
	params := rez.UpdateIncidentMilestoneParams{
		Timestamp:   attr.Timestamp,
		Description: attr.Description,
	}
	if _, updateErr := h.management.UpdateMilestone(ctx, request.Id, params); updateErr != nil {
		return nil, oapi.Error(ctx, "update incident milestone", updateErr)
	}
	return &oapi.UpdateIncidentMilestoneResponse{}, nil
}

func (h *incidentsHandler) CreateIncidentRoleAssignment(ctx context.Context, request *oapi.CreateIncidentRoleAssignmentRequest) (*oapi.CreateIncidentRoleAssignmentResponse, error) {
	attr := request.Body.Attributes
	if _, saveErr := h.management.SetRoleAssignment(ctx, request.Id, attr.RoleId, attr.UserId); saveErr != nil {
		return nil, oapi.Error(ctx, "assign incident role", saveErr)
	}
	return &oapi.CreateIncidentRoleAssignmentResponse{}, nil
}

func (h *incidentsHandler) DeleteIncidentRoleAssignment(ctx context.Context, request *oapi.DeleteIncidentRoleAssignmentRequest) (*oapi.DeleteIncidentRoleAssignmentResponse, error) {
	if deleteErr := h.management.RemoveRoleAssignment(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete incident role assignment", deleteErr)
	}
	return &oapi.DeleteIncidentRoleAssignmentResponse{}, nil
}

func (h *incidentsHandler) LinkIncidentSituation(ctx context.Context, request *oapi.LinkIncidentSituationRequest) (*oapi.LinkIncidentSituationResponse, error) {
	if linkErr := h.incidents.LinkSituation(ctx, request.Id, request.Body.Attributes.SituationId); linkErr != nil {
		return nil, oapi.Error(ctx, "link incident situation", linkErr)
	}
	return &oapi.LinkIncidentSituationResponse{}, nil
}

func (h *incidentsHandler) UnlinkIncidentSituation(ctx context.Context, request *oapi.UnlinkIncidentSituationRequest) (*oapi.UnlinkIncidentSituationResponse, error) {
	if unlinkErr := h.incidents.UnlinkSituation(ctx, request.Id, request.SituationId); unlinkErr != nil {
		return nil, oapi.Error(ctx, "unlink incident situation", unlinkErr)
	}
	return &oapi.UnlinkIncidentSituationResponse{}, nil
}

func (h *incidentsHandler) ListIncidentUpdates(ctx context.Context, req *oapi.ListIncidentUpdatesRequest) (*oapi.ListIncidentUpdatesResponse, error) {
	if _, getErr := h.incidents.GetIncident(ctx, incident.ID(req.Id)); getErr != nil {
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
