package apiv1

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentfield"
	"github.com/rezible/rezible/ent/incidentfieldoption"
	"github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttag"
	"github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/ent/schema"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentMetadataHandler struct {
	db        *ent.Client
	incidents rez.IncidentService

	fields     *ent.IncidentFieldClient
	roles      *ent.IncidentRoleClient
	severities *ent.IncidentSeverityClient
	tags       *ent.IncidentTagClient
	types      *ent.IncidentTypeClient
}

func newIncidentMetadataHandler(db *ent.Client, incidents rez.IncidentService) *incidentMetadataHandler {
	return &incidentMetadataHandler{
		db:        db,
		incidents: incidents,

		fields:     db.IncidentField,
		roles:      db.IncidentRole,
		severities: db.IncidentSeverity,
		tags:       db.IncidentTag,
		types:      db.IncidentType,
	}
}

func (h *incidentMetadataHandler) GetIncidentMetadata(ctx context.Context, request *oapi.GetIncidentMetadataRequest) (*oapi.GetIncidentMetadataResponse, error) {
	var resp oapi.GetIncidentMetadataResponse

	md, mdErr := h.incidents.GetIncidentMetadata(ctx)
	if mdErr != nil {
		return nil, oapi.Error("get metadata", mdErr)
	}
	resp.Body.Data = oapi.IncidentMetadataFromRez(md)

	return &resp, nil
}

func (h *incidentMetadataHandler) ListIncidentSeverities(ctx context.Context, request *oapi.ListIncidentSeveritiesRequest) (*oapi.ListIncidentSeveritiesResponse, error) {
	var resp oapi.ListIncidentSeveritiesResponse

	query := h.severities.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidentseverity.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidentseverity.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to query incident severities", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentSeverity, len(res))
	for i, tag := range res {
		resp.Body.Data[i] = oapi.IncidentSeverityFromEnt(tag)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, oapi.Error("Failed to query incident severity count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentMetadataHandler) CreateIncidentSeverity(ctx context.Context, request *oapi.CreateIncidentSeverityRequest) (*oapi.CreateIncidentSeverityResponse, error) {
	var resp oapi.CreateIncidentSeverityResponse

	attr := request.Body.Attributes
	query := h.severities.Create().
		SetName(attr.Name)

	sev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error("Failed to create incident severity", createErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentMetadataHandler) GetIncidentSeverity(ctx context.Context, request *oapi.GetIncidentSeverityRequest) (*oapi.GetIncidentSeverityResponse, error) {
	var resp oapi.GetIncidentSeverityResponse

	sev, queryErr := h.severities.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error("Failed to get incident tag", queryErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentMetadataHandler) UpdateIncidentSeverity(ctx context.Context, request *oapi.UpdateIncidentSeverityRequest) (*oapi.UpdateIncidentSeverityResponse, error) {
	var resp oapi.UpdateIncidentSeverityResponse

	attr := request.Body.Attributes
	query := h.severities.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	sev, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error("Failed to update incident severity", updateErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentMetadataHandler) ArchiveIncidentSeverity(ctx context.Context, request *oapi.ArchiveIncidentSeverityRequest) (*oapi.ArchiveIncidentSeverityResponse, error) {
	var resp oapi.ArchiveIncidentSeverityResponse

	delErr := h.severities.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, oapi.Error("Failed to archive incident severity", delErr)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) ListIncidentTypes(ctx context.Context, request *oapi.ListIncidentTypesRequest) (*oapi.ListIncidentTypesResponse, error) {
	var resp oapi.ListIncidentTypesResponse

	query := h.types.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidenttype.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidenttype.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to query incident types", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentType, len(res))
	for i, t := range res {
		resp.Body.Data[i] = oapi.IncidentTypeFromEnt(t)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, oapi.Error("Failed to query incident type count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentMetadataHandler) CreateIncidentType(ctx context.Context, request *oapi.CreateIncidentTypeRequest) (*oapi.CreateIncidentTypeResponse, error) {
	var resp oapi.CreateIncidentTypeResponse

	attr := request.Body.Attributes
	query := h.types.Create().SetName(attr.Name)
	t, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error("Failed to create incident type", createErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentMetadataHandler) GetIncidentType(ctx context.Context, request *oapi.GetIncidentTypeRequest) (*oapi.GetIncidentTypeResponse, error) {
	var resp oapi.GetIncidentTypeResponse

	t, queryErr := h.types.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error("Failed to get incident type", queryErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentMetadataHandler) UpdateIncidentType(ctx context.Context, request *oapi.UpdateIncidentTypeRequest) (*oapi.UpdateIncidentTypeResponse, error) {
	var resp oapi.UpdateIncidentTypeResponse

	attr := request.Body.Attributes
	query := h.types.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	t, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error("Failed to update incident type", updateErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentMetadataHandler) ArchiveIncidentType(ctx context.Context, request *oapi.ArchiveIncidentTypeRequest) (*oapi.ArchiveIncidentTypeResponse, error) {
	var resp oapi.ArchiveIncidentTypeResponse

	delErr := h.types.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, oapi.Error("Failed to archive incident type", delErr)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) ListIncidentRoles(ctx context.Context, request *oapi.ListIncidentRolesRequest) (*oapi.ListIncidentRolesResponse, error) {
	var resp oapi.ListIncidentRolesResponse

	query := h.roles.Query()

	if true {
		ctx = schema.IncludeArchived(ctx)
	}

	res, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to query incident roles", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentRole, len(res))
	for i, role := range res {
		resp.Body.Data[i] = oapi.IncidentRoleFromEnt(role)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) CreateIncidentRole(ctx context.Context, request *oapi.CreateIncidentRoleRequest) (*oapi.CreateIncidentRoleResponse, error) {
	var resp oapi.CreateIncidentRoleResponse

	attr := request.Body.Attributes

	query := h.roles.Create().
		SetName(attr.Name).
		SetRequired(attr.Required)

	role, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error("Failed to create incident role", createErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentMetadataHandler) GetIncidentRole(ctx context.Context, request *oapi.GetIncidentRoleRequest) (*oapi.GetIncidentRoleResponse, error) {
	var resp oapi.GetIncidentRoleResponse

	ctx = schema.IncludeArchived(ctx)
	role, queryErr := h.roles.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error("Failed to get incident role", queryErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentMetadataHandler) UpdateIncidentRole(ctx context.Context, request *oapi.UpdateIncidentRoleRequest) (*oapi.UpdateIncidentRoleResponse, error) {
	var resp oapi.UpdateIncidentRoleResponse

	attr := request.Body.Attributes
	query := h.roles.UpdateOneID(request.Id).
		SetNillableName(attr.Name).
		SetNillableRequired(attr.Required)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	role, saveErr := query.Save(ctx)
	if saveErr != nil {
		return nil, oapi.Error("Failed to update incident role", saveErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentMetadataHandler) ArchiveIncidentRole(ctx context.Context, request *oapi.ArchiveIncidentRoleRequest) (*oapi.ArchiveIncidentRoleResponse, error) {
	var resp oapi.ArchiveIncidentRoleResponse

	delErr := h.roles.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, oapi.Error("Failed to archive incident role", delErr)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) ListIncidentTags(ctx context.Context, request *oapi.ListIncidentTagsRequest) (*oapi.ListIncidentTagsResponse, error) {
	var resp oapi.ListIncidentTagsResponse

	query := h.tags.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidenttag.ValueContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidenttag.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to query incident tags", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentTag, len(res))
	for i, tag := range res {
		resp.Body.Data[i] = oapi.IncidentTagFromEnt(tag)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, oapi.Error("Failed to query incident tag count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentMetadataHandler) CreateIncidentTag(ctx context.Context, request *oapi.CreateIncidentTagRequest) (*oapi.CreateIncidentTagResponse, error) {
	var resp oapi.CreateIncidentTagResponse

	attr := request.Body.Attributes
	query := h.tags.Create().SetValue(attr.Value)
	tag, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, oapi.Error("Failed to create incident tag", createErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentMetadataHandler) GetIncidentTag(ctx context.Context, request *oapi.GetIncidentTagRequest) (*oapi.GetIncidentTagResponse, error) {
	var resp oapi.GetIncidentTagResponse

	tag, queryErr := h.tags.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error("Failed to get incident tag", queryErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentMetadataHandler) UpdateIncidentTag(ctx context.Context, request *oapi.UpdateIncidentTagRequest) (*oapi.UpdateIncidentTagResponse, error) {
	var resp oapi.UpdateIncidentTagResponse

	attr := request.Body.Attributes
	query := h.tags.UpdateOneID(request.Id).
		SetNillableValue(attr.Value)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	tag, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error("Failed to update incident tag", updateErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentMetadataHandler) ArchiveIncidentTag(ctx context.Context, request *oapi.ArchiveIncidentTagRequest) (*oapi.ArchiveIncidentTagResponse, error) {
	var resp oapi.ArchiveIncidentTagResponse

	delErr := h.tags.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, oapi.Error("Failed to archive incident tag", delErr)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) ListIncidentFields(ctx context.Context, request *oapi.ListIncidentFieldsRequest) (*oapi.ListIncidentFieldsResponse, error) {
	var resp oapi.ListIncidentFieldsResponse

	query := h.fields.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}

	query = query.WithOptions(func(q *ent.IncidentFieldOptionQuery) {
		q.Where(incidentfieldoption.ArchiveTimeIsNil())
	})

	res, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to query incident fields", queryErr)
	}
	slog.Debug("ListIncidentFields", "res", res)

	resp.Body.Data = make([]oapi.IncidentField, len(res))
	for i, field := range res {
		resp.Body.Data[i] = oapi.IncidentFieldFromEnt(field)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) CreateIncidentField(ctx context.Context, request *oapi.CreateIncidentFieldRequest) (*oapi.CreateIncidentFieldResponse, error) {
	var resp oapi.CreateIncidentFieldResponse

	createFieldOptionsTx := func(tx *ent.Tx) error {
		attr := request.Body.Attributes
		query := tx.IncidentField.Create().
			SetName(attr.Name)

		if attr.IncidentType != nil {
			// TODO: check incident type
		}

		field, createErr := query.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("saving incident field: %w", createErr)
		}

		createQuery := tx.IncidentFieldOption.MapCreateBulk(attr.Options, func(c *ent.IncidentFieldOptionCreate, i int) {
			opt := attr.Options[i]
			c.SetType(incidentfieldoption.Type(opt.FieldOptionType)).
				SetValue(opt.Value).
				SetIncidentFieldID(field.ID)
		})
		opts, createOptsErr := createQuery.Save(ctx)
		if createOptsErr != nil {
			return fmt.Errorf("saving custom options: %w", createOptsErr)
		}
		field.Edges.Options = opts
		resp.Body.Data = oapi.IncidentFieldFromEnt(field)
		return nil
	}

	if txErr := ent.WithTx(ctx, h.db, createFieldOptionsTx); txErr != nil {
		return nil, oapi.Error("Failed to create incident field", txErr)
	}

	return &resp, nil
}

func (h *incidentMetadataHandler) GetIncidentField(ctx context.Context, request *oapi.GetIncidentFieldRequest) (*oapi.GetIncidentFieldResponse, error) {
	var resp oapi.GetIncidentFieldResponse

	ctx = schema.IncludeArchived(ctx)
	field, queryErr := h.fields.Query().
		Where(incidentfield.ID(request.Id)).
		WithOptions().
		Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error("Failed to get incident field", queryErr)
	}
	resp.Body.Data = oapi.IncidentFieldFromEnt(field)

	return &resp, nil
}

func (h *incidentMetadataHandler) updateIncidentFieldOptions(tx *ent.Tx, ctx context.Context, fieldId uuid.UUID, reqOptions []oapi.UpdateIncidentFieldOptionAttributes) error {
	currentOptions, optionsErr := tx.IncidentFieldOption.Query().
		Where(incidentfieldoption.IncidentFieldID(fieldId)).
		All(ctx)
	if optionsErr != nil {
		return oapi.Error("Failed to get incident field options", optionsErr)
	}

	options := make(map[string]*ent.IncidentFieldOption)
	var curType incidentfieldoption.Type
	for _, opt := range currentOptions {
		options[opt.ID.String()] = opt
		curType = opt.Type
	}

	var reqType incidentfieldoption.Type
	for _, o := range reqOptions {
		t := incidentfieldoption.Type(o.FieldOptionType)
		if len(reqType) > 0 && t != reqType {
			return oapi.Error("multiple field option types", nil)
		}
		reqType = t
	}

	if curType != reqType {
		//ctx = schema.IncludeArchived(ctx)
		deleteOthers := tx.IncidentFieldOption.Delete().
			Where(incidentfieldoption.And(
				incidentfieldoption.IncidentFieldID(fieldId),
				incidentfieldoption.TypeNEQ(reqType)))
		if _, err := deleteOthers.Exec(ctx); err != nil {
			return oapi.Error("Failed to delete existing options", err)
		}
	}

	var newOptions []*ent.IncidentFieldOption
	for _, o := range reqOptions {
		option := o
		opt := &ent.IncidentFieldOption{
			IncidentFieldID: fieldId,
			Type:            incidentfieldoption.Type(option.FieldOptionType),
			Value:           option.Value,
		}
		if option.Id == nil {
			newOptions = append(newOptions, opt)
			continue
		}
		cur, exists := options[*option.Id]
		if !exists {
			err := fmt.Errorf("cannot update non-existant option id: %s", *option.Id)
			return oapi.Error("failed to update field option", err)
		}
		archiveTime := cur.ArchiveTime
		if o.Archived && archiveTime.IsZero() {
			archiveTime = time.Now()
		}
		needsUpdate := opt.Value != cur.Value || o.Archived != (!cur.ArchiveTime.IsZero())
		if needsUpdate {
			update := tx.IncidentFieldOption.UpdateOneID(cur.ID).
				SetType(opt.Type).
				SetValue(opt.Value)
			if o.Archived {
				update.SetArchiveTime(archiveTime)
			} else {
				update.ClearArchiveTime()
			}
			if updateErr := update.Exec(ctx); updateErr != nil {
				return oapi.Error("failed to update field option", updateErr)
			}
		}
	}
	create := tx.IncidentFieldOption.MapCreateBulk(newOptions, func(c *ent.IncidentFieldOptionCreate, i int) {
		opt := newOptions[i]
		c.SetValue(opt.Value).
			SetType(opt.Type).
			SetIncidentFieldID(opt.IncidentFieldID)
	})
	if createErr := create.Exec(ctx); createErr != nil {
		return oapi.Error("failed to create new field options", createErr)
	}
	return nil
}

func (h *incidentMetadataHandler) UpdateIncidentField(ctx context.Context, request *oapi.UpdateIncidentFieldRequest) (*oapi.UpdateIncidentFieldResponse, error) {
	var resp oapi.UpdateIncidentFieldResponse

	attr := request.Body.Attributes

	tx, txErr := h.db.BeginTx(ctx, nil)
	if txErr != nil {
		return nil, oapi.Error("Failed to start db transaction", txErr)
	}
	defer tx.Rollback()

	query := tx.IncidentField.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil {
		if *attr.Archived == false {
			query.ClearArchiveTime()
		} else {
			query.SetArchiveTime(time.Now())
		}
	}

	if attr.Options != nil {
		updateOptionsErr := h.updateIncidentFieldOptions(tx, ctx, request.Id, *attr.Options)
		if updateOptionsErr != nil {
			return nil, updateOptionsErr
		}
	}

	field, saveErr := query.Save(ctx)
	if saveErr != nil {
		return nil, oapi.Error("Failed to update incident field", saveErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, oapi.Error("Failed to create field", commitErr)
	}

	resp.Body.Data = oapi.IncidentFieldFromEnt(field)

	return &resp, nil
}

func (h *incidentMetadataHandler) ArchiveIncidentField(ctx context.Context, request *oapi.ArchiveIncidentFieldRequest) (*oapi.ArchiveIncidentFieldResponse, error) {
	var resp oapi.ArchiveIncidentFieldResponse

	delErr := h.fields.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, oapi.Error("Failed to archive incident field", delErr)
	}

	return &resp, nil
}
