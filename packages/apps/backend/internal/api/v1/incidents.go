package apiv1

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentsHandler struct {
	incidents rez.IncidentService
	agents    rez.AgentService
}

func newIncidentsHandler(incidents rez.IncidentService, agents rez.AgentService) *incidentsHandler {
	return &incidentsHandler{incidents: incidents, agents: agents}
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
		ListParams: req.ListParams(),
	}
	incs, listErr := h.incidents.ListIncidents(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list incidents", listErr)
	}
	resp.Body.Data = make([]oapi.Incident, len(incs.Data))
	for i, inc := range incs.Data {
		resp.Body.Data[i] = oapi.IncidentFromEnt(inc)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: incs.Count,
	}

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
	var resp oapi.CreateIncidentResponse

	/*
		attr := input.Body.Attributes
		setFn := func(m *ent.IncidentMutation) []ent.Mutation {
			m.SetTitle(attr.Title)
			m.SetSeverityID(attr.SeverityId)
			m.SetTypeID(attr.TypeId)
			if attr.Summary != nil {
				m.SetSummary(*attr.Summary)
			}
			if len(attr.TagIds) > 0 {
				m.AddTagAssignmentIDs(attr.TagIds...)
			}
			if len(attr.FieldSelectionIds) > 0 {
				m.AddFieldSelectionIDs(attr.FieldSelectionIds...)
			}

			incidentId, exists := m.ID()
			if !exists {
				return nil
			}

			createMilestone := m.Client().IncidentMilestone.Create().
				SetKind(im.KindOpened).
				SetDescription("Incident created via API").
				SetTimestamp(time.Now()).
				SetSource("api").
				SetIncidentID(incidentId)

			return []ent.Mutation{createMilestone.Mutation()}
		}

		created, createErr := h.incidents.Set(ctx, setFn)
		if createErr != nil {
			return nil, oapi.Error(ctx, "create incident", createErr)
		}
		resp.Body.Data = oapi.IncidentFromEnt(created)
	*/

	return &resp, nil
}

func (h *incidentsHandler) UpdateIncident(ctx context.Context, request *oapi.UpdateIncidentRequest) (*oapi.UpdateIncidentResponse, error) {
	var resp oapi.UpdateIncidentResponse

	attr := request.Body.Attributes
	setFn := func(m *ent.IncidentMutation) {
		if attr.Title != nil {
			m.SetTitle(*attr.Title)
		}
		if attr.Summary != nil {
			m.SetSummary(*attr.Summary)
		}
		if attr.SeverityId != uuid.Nil {
			m.SetSeverityID(attr.SeverityId)
		}
		if attr.TypeId != uuid.Nil {
			m.SetTypeID(attr.TypeId)
		}
	}

	updated, updateErr := h.incidents.Set(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update incident", updateErr)
	}
	resp.Body.Data = oapi.IncidentFromEnt(updated)

	return &resp, nil
}

func (h *incidentsHandler) ArchiveIncident(ctx context.Context, input *oapi.ArchiveIncidentRequest) (*oapi.ArchiveIncidentResponse, error) {
	var resp oapi.ArchiveIncidentResponse

	if archiveErr := h.incidents.Archive(ctx, input.Id); archiveErr != nil {
		return nil, oapi.Error(ctx, "archive incident", archiveErr)
	}

	return &resp, nil
}

func (h *incidentsHandler) ListIncidentImpacts(ctx context.Context, input *oapi.ListIncidentImpactsRequest) (*oapi.ListIncidentImpactsResponse, error) {
	var resp oapi.ListIncidentImpactsResponse

	inc, incErr := h.incidents.Get(ctx, incidentIdPredicate(input.Id))
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	impacts, impactsErr := h.incidents.ListIncidentImpacts(ctx, inc.ID)
	if impactsErr != nil {
		return nil, oapi.Error(ctx, "list incident impacts", impactsErr)
	}
	resp.Body.Data = make([]oapi.IncidentImpact, len(impacts))
	for i, impact := range impacts {
		resp.Body.Data[i] = oapi.IncidentImpactFromEnt(impact)
	}
	resp.Body.Pagination.Total = len(impacts)
	return &resp, nil
}

func (h *incidentsHandler) SetIncidentImpacts(ctx context.Context, input *oapi.SetIncidentImpactsRequest) (*oapi.SetIncidentImpactsResponse, error) {
	var resp oapi.SetIncidentImpactsResponse

	inc, incErr := h.incidents.Get(ctx, incidentIdPredicate(input.Id))
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	params := rez.SetIncidentImpactsParams{
		IncidentID: inc.ID,
		Impacts:    make([]rez.IncidentImpactInput, len(input.Body.Attributes.Impacts)),
	}
	for i, impact := range input.Body.Attributes.Impacts {
		if impact.KnowledgeEntityId != nil {
			params.Impacts[i].KnowledgeEntityID = *impact.KnowledgeEntityId
		}
		params.Impacts[i].Kind = impact.Kind
		params.Impacts[i].DisplayName = impact.DisplayName
		params.Impacts[i].Description = impact.Description
		params.Impacts[i].Source = impact.Source
		params.Impacts[i].Note = impact.Note
	}
	impacts, impactsErr := h.incidents.SetIncidentImpacts(ctx, params)
	if impactsErr != nil {
		return nil, oapi.Error(ctx, "set incident impacts", impactsErr)
	}
	resp.Body.Data = make([]oapi.IncidentImpact, len(impacts))
	for i, impact := range impacts {
		resp.Body.Data[i] = oapi.IncidentImpactFromEnt(impact)
	}
	resp.Body.Pagination.Total = len(impacts)
	return &resp, nil
}

func (h *incidentsHandler) GetIncidentContextPack(ctx context.Context, input *oapi.GetIncidentContextPackRequest) (*oapi.GetIncidentContextPackResponse, error) {
	var resp oapi.GetIncidentContextPackResponse

	inc, incErr := h.incidents.Get(ctx, incidentIdPredicate(input.Id))
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	pack, packErr := h.incidents.GetIncidentContextPack(ctx, inc.ID)
	if packErr != nil {
		return nil, oapi.Error(ctx, "get incident context pack", packErr)
	}
	resp.Body.Data = oapi.IncidentContextPackFromDomain(pack)
	return &resp, nil
}

func (h *incidentsHandler) RequestIncidentContextPackAgentRun(ctx context.Context, input *oapi.RequestIncidentContextPackAgentRunRequest) (*oapi.RequestIncidentContextPackAgentRunResponse, error) {
	var resp oapi.RequestIncidentContextPackAgentRunResponse

	inc, incErr := h.incidents.Get(ctx, incidentIdPredicate(input.Id))
	if incErr != nil {
		return nil, oapi.Error(ctx, "get incident", incErr)
	}
	run, runErr := h.agents.RequestRun(ctx, rez.AgentRunRequest{
		WorkflowKind:   agentrun.WorkflowKindIncidentContextPack,
		IdempotencyKey: fmt.Sprintf("incident-context-pack:manual:%s:%d", inc.ID, time.Now().UTC().UnixNano()),
		SubjectKind:    "incident",
		SubjectID:      inc.ID,
		Metadata: map[string]any{
			"trigger": "manual",
		},
	})
	if runErr != nil {
		return nil, oapi.Error(ctx, "request incident context pack agent run", runErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}
