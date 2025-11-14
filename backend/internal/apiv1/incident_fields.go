package apiv1

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentfield"
	"github.com/rezible/rezible/ent/incidentfieldoption"
	"github.com/rezible/rezible/ent/schema"
	oapi "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
	"time"
)

type incidentFieldsHandler struct {
	fields *ent.IncidentFieldClient
	db     *ent.Client
}

func newIncidentFieldsHandler(db *ent.Client) *incidentFieldsHandler {
	return &incidentFieldsHandler{db.IncidentField, db}
}

func (h *incidentFieldsHandler) ListIncidentFields(ctx context.Context, request *oapi.ListIncidentFieldsRequest) (*oapi.ListIncidentFieldsResponse, error) {
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
		return nil, apiError("Failed to query incident fields", queryErr)
	}
	log.Debug().Interface("res", res).Msg("ListIncidentFields")

	resp.Body.Data = make([]oapi.IncidentField, len(res))
	for i, field := range res {
		resp.Body.Data[i] = oapi.IncidentFieldFromEnt(field)
	}

	return &resp, nil
}

func (h *incidentFieldsHandler) CreateIncidentField(ctx context.Context, request *oapi.CreateIncidentFieldRequest) (*oapi.CreateIncidentFieldResponse, error) {
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
		return nil, apiError("Failed to create incident field", txErr)
	}

	return &resp, nil
}

func (h *incidentFieldsHandler) GetIncidentField(ctx context.Context, request *oapi.GetIncidentFieldRequest) (*oapi.GetIncidentFieldResponse, error) {
	var resp oapi.GetIncidentFieldResponse

	ctx = schema.IncludeArchived(ctx)
	field, queryErr := h.db.IncidentField.Query().
		Where(incidentfield.ID(request.Id)).
		WithOptions().
		Only(ctx)
	if queryErr != nil {
		return nil, apiError("Failed to get incident field", queryErr)
	}
	resp.Body.Data = oapi.IncidentFieldFromEnt(field)

	return &resp, nil
}

func (h *incidentFieldsHandler) updateIncidentFieldOptions(tx *ent.Tx, ctx context.Context, fieldId uuid.UUID, reqOptions []oapi.UpdateIncidentFieldOptionAttributes) error {
	currentOptions, optionsErr := tx.IncidentFieldOption.Query().
		Where(incidentfieldoption.IncidentFieldID(fieldId)).
		All(ctx)
	if optionsErr != nil {
		return apiError("Failed to get incident field options", optionsErr)
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
			return apiError("multiple field option types", nil)
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
			return apiError("Failed to delete existing options", err)
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
			return apiError("failed to update field option", err)
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
				return apiError("failed to update field option", updateErr)
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
		return apiError("failed to create new field options", createErr)
	}
	return nil
}

func (h *incidentFieldsHandler) UpdateIncidentField(ctx context.Context, request *oapi.UpdateIncidentFieldRequest) (*oapi.UpdateIncidentFieldResponse, error) {
	var resp oapi.UpdateIncidentFieldResponse

	attr := request.Body.Attributes

	tx, txErr := h.db.BeginTx(ctx, nil)
	if txErr != nil {
		return nil, apiError("Failed to start db transaction", txErr)
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
		return nil, apiError("Failed to update incident field", saveErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, apiError("Failed to create field", commitErr)
	}

	resp.Body.Data = oapi.IncidentFieldFromEnt(field)

	return &resp, nil
}

func (h *incidentFieldsHandler) ArchiveIncidentField(ctx context.Context, request *oapi.ArchiveIncidentFieldRequest) (*oapi.ArchiveIncidentFieldResponse, error) {
	var resp oapi.ArchiveIncidentFieldResponse

	delErr := h.db.IncidentField.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, apiError("Failed to archive incident field", delErr)
	}

	return &resp, nil
}
