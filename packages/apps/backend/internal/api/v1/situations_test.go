package apiv1

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test/mocks"
)

func TestRaiseSituationErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		name       string
		serviceErr error
		status     int
	}{
		{"closed", rez.ErrConflict, http.StatusConflict},
		{"absent", rez.ErrNotFound, http.StatusNotFound},
		{"invalid", rez.ErrInvalidInput, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			id := uuid.New()
			service := mocks.NewMockSituationService(t)
			service.EXPECT().RaiseSituation(ctx, id, rez.RaiseSituationParams{StartInvestigation: true}).
				Return(nil, fmt.Errorf("raise: %w", tc.serviceErr))
			handler := newSituationsHandler(nil, service)
			response, raiseErr := handler.RaiseSituation(ctx, &oapi.RaiseSituationRequest{Id: id})
			require.Nil(t, response)
			var statusErr huma.StatusError
			require.ErrorAs(t, raiseErr, &statusErr)
			require.Equal(t, tc.status, statusErr.GetStatus())
		})
	}
}

func TestListSituationsMutedQuery(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		muted *bool
	}{
		{"omitted", "", nil},
		{"false", "?muted=false", new(false)},
		{"true", "?muted=true", new(true)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := mocks.NewMockSituationService(t)
			service.EXPECT().ListSituations(mock.Anything, mock.Anything).
				RunAndReturn(func(_ context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
					require.Equal(t, tc.muted, params.Muted)
					return &ent.ListResult[ent.Situation]{}, nil
				})
			_, api := humatest.New(t)
			huma.Register(api, oapi.ListSituations, newSituationsHandler(nil, service).ListSituations)

			response := api.Get("/situations" + tc.query)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		})
	}

	t.Run("invalid", func(t *testing.T) {
		service := mocks.NewMockSituationService(t)
		_, api := humatest.New(t)
		huma.Register(api, oapi.ListSituations, newSituationsHandler(nil, service).ListSituations)

		response := api.Get("/situations?muted=sometimes")
		require.Equal(t, http.StatusUnprocessableEntity, response.Code, response.Body.String())
	})
}
