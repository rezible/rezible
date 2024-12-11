package api

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentresourceimpact"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentsHandler struct {
	incidents      *ent.IncidentClient
	resourceImpact *ent.IncidentResourceImpactClient
	db             *ent.Client
}

func newIncidentsHandler(db *ent.Client) *incidentsHandler {
	return &incidentsHandler{
		db.Incident,
		db.IncidentResourceImpact,
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

	if true {
		query.WithImpactedResources(func(q *ent.IncidentResourceImpactQuery) {
			q.WithIncident().WithService().WithFunctionality()
		})
	}

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
		WithImpactedResources().
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

	impactTx := tx.IncidentResourceImpact
	createdImpacts, deletedImpactIds, impactErr := h.updateIncidentResourceImpacts(ctx, impactTx, inc, attr)
	if impactErr != nil {
		return nil, fmt.Errorf("failed to update impacted resources: %w", impactErr)
	}
	update.AddImpactedResources(createdImpacts...)
	update.RemoveImpactedResourceIDs(deletedImpactIds...)

	return update.Save(ctx)
}

func (h *incidentsHandler) updateIncidentResourceImpacts(
	ctx context.Context,
	client *ent.IncidentResourceImpactClient,
	inc *ent.Incident,
	attr oapi.UpdateIncidentAttributes,
) (
	[]*ent.IncidentResourceImpact,
	[]uuid.UUID,
	error,
) {
	current, currentErr := inc.Edges.ImpactedResourcesOrErr()
	if currentErr != nil {
		return nil, nil, fmt.Errorf("no impacted resources edge: %w", currentErr)
	}

	curSvcIds := mapset.NewThreadUnsafeSet[uuid.UUID]()
	curFncIds := mapset.NewThreadUnsafeSet[uuid.UUID]()
	for _, res := range current {
		curSvcIds.Add(res.ServiceID)
		curFncIds.Add(res.FunctionalityID)
	}
	curSvcIds.Remove(uuid.Nil)
	curFncIds.Remove(uuid.Nil)

	var newImpactCreations []*ent.IncidentResourceImpactCreate
	omittedIds := mapset.NewThreadUnsafeSet[uuid.UUID]()

	if attr.Services != nil {
		reqSvcIds, svcsErr := parseUUIDs(*attr.Services)
		if svcsErr != nil {
			return nil, nil, fmt.Errorf("invalid services: %w", svcsErr)
		}
		reqSvcIds.Difference(curSvcIds).Each(func(newId uuid.UUID) bool {
			creation := client.Create().SetIncidentID(inc.ID).SetServiceID(newId)
			newImpactCreations = append(newImpactCreations, creation)
			return false
		})
		omittedIds = omittedIds.Union(curSvcIds.Difference(reqSvcIds))
	}

	if attr.Functionalities != nil {
		reqFncIds, fncsErr := parseUUIDs(*attr.Functionalities)
		if fncsErr != nil {
			return nil, nil, fmt.Errorf("invalid functionalities: %w", fncsErr)
		}
		reqFncIds.Difference(curFncIds).Each(func(newId uuid.UUID) bool {
			creation := client.Create().SetIncidentID(inc.ID).SetServiceID(newId)
			newImpactCreations = append(newImpactCreations, creation)
			return false
		})
		omittedIds = omittedIds.Union(curFncIds.Difference(reqFncIds))
	}

	var created []*ent.IncidentResourceImpact
	if len(newImpactCreations) > 0 {
		var createErr error
		created, createErr = client.CreateBulk(newImpactCreations...).
			Save(ctx)
		if createErr != nil {
			return nil, nil, fmt.Errorf("failed to create resource impacts: %w", createErr)
		}
	}

	deletedIds := omittedIds.ToSlice()
	if len(deletedIds) == 0 {
		numDeleted, delErr := client.Delete().Where(incidentresourceimpact.IDIn(deletedIds...)).Exec(ctx)
		if delErr != nil {
			return nil, nil, fmt.Errorf("failed to delete resource impacts: %w", delErr)
		}
		if numDeleted != len(deletedIds) {
			log.Warn().
				Int("deleted", numDeleted).
				Int("expected to delete", len(deletedIds)).
				Msg("unexpected number of deleted resource impacts")
		}
	}

	return created, deletedIds, nil
}

func parseUUIDs(rawIds []string) (mapset.Set[uuid.UUID], error) {
	ids := mapset.NewThreadUnsafeSet[uuid.UUID]()
	for _, rawId := range rawIds {
		id, idErr := uuid.Parse(rawId)
		if idErr != nil {
			return nil, fmt.Errorf("%s not a uuid: %w", rawId, idErr)
		}
		ids.Add(id)
	}
	return ids, nil
}
