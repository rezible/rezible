package api

/*
import (
	"context"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type laddersHandler struct {
	ladders *ent.LadderClient
}

func newLaddersHandler(ladders *ent.LadderClient) *laddersHandler {
	return &laddersHandler{ladders}
}

func (h *laddersHandler) ListLadders(ctx context.Context, input *oapi.ListLaddersRequest) (*oapi.ListLaddersResponse, error) {
	var resp oapi.ListLaddersResponse

	query := h.ladders.Query()

	ladders, err := query.All(ctx)
	if err != nil {
		return nil, detailError("failed to list ladders", err)
	}
	resp.Body.Data = make([]oapi.Ladder, len(ladders))
	for i, l := range ladders {
		resp.Body.Data[i] = oapi.LadderFromEnt(l)
	}

	return &resp, nil
}

func (h *laddersHandler) GetLadder(ctx context.Context, input *oapi.GetLadderRequest) (*oapi.GetLadderResponse, error) {
	var resp oapi.GetLadderResponse

	ladder, err := h.ladders.Get(ctx, input.Id)
	if err != nil {
		return nil, detailError("failed to get ladder", err)
	}
	resp.Body.Data = oapi.LadderFromEnt(ladder)

	return &resp, nil
}

func (h *laddersHandler) CreateLadder(ctx context.Context, input *oapi.CreateLadderRequest) (*oapi.CreateLadderResponse, error) {
	var resp oapi.CreateLadderResponse

	attr := input.Body.Attributes
	query := h.ladders.Create().
		SetName(attr.Name)

	ladder, err := query.Save(ctx)
	if err != nil {
		return nil, detailError("failed to create ladder", err)
	}
	resp.Body.Data = oapi.LadderFromEnt(ladder)

	return &resp, nil
}

func (h *laddersHandler) UpdateLadder(ctx context.Context, input *oapi.UpdateLadderRequest) (*oapi.UpdateLadderResponse, error) {
	var resp oapi.UpdateLadderResponse

	attr := input.Body.Attributes
	query := h.ladders.UpdateOneID(input.Id).
		SetNillableName(attr.Name.NillableValue())

	ladder, err := query.Save(ctx)
	if err != nil {
		return nil, detailError("failed to update ladder", err)
	}
	resp.Body.Data = oapi.LadderFromEnt(ladder)

	return &resp, nil
}

func (h *laddersHandler) ArchiveLadder(ctx context.Context, input *oapi.ArchiveLadderRequest) (*oapi.ArchiveLadderResponse, error) {
	var resp oapi.ArchiveLadderResponse

	err := h.ladders.DeleteOneID(input.Id).Exec(ctx)
	if err != nil {
		return nil, detailError("failed to archive ladder", err)
	}

	return &resp, nil
}
*/
