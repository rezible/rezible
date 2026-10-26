package apiv1

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type SituationsHandlerSuite struct {
	test.Suite
}

func TestSituationsHandlerSuite(t *testing.T) {
	suite.Run(t, &SituationsHandlerSuite{Suite: test.NewSuite()})
}

func (s *SituationsHandlerSuite) TestGetSituationByReference() {
	_, tdb := s.SetupTestDatabase()
	ctx, identity := s.NewIdentity(tdb, "a")
	client := tdb.Client(ctx)
	seed := client.KnowledgeEntity.Create().SetCategory(knowledgeentity.CategoryContainer).SetKind("service").SaveX(ctx)
	for number, title := range map[int]string{1: "First", 2: "Second"} {
		createSituation := client.Situation.Create().
			SetNumber(number).
			SetTitle(title).
			SetOpenedAt(time.Now()).
			SetSeedEntityID(seed.ID)
		createSituation.ExecX(ctx)
	}
	situations, situationsErr := db.NewSituationService(rez.SituationsConfig{}, tdb, test.NewClock(time.Now()), nil, nil, nil, nil, nil, noJudge{t: s.T()})
	s.Require().NoError(situationsErr)

	_, api := humatest.New(s.T())
	api.UseMiddleware(func(hctx huma.Context, next func(huma.Context)) {
		next(huma.WithContext(hctx, execution.NewUserContext(hctx.Context(), identity.Session)))
	})
	huma.Register(api, oapi.GetSituation, newSituationsHandler(nil, situations, nil, nil).GetSituation)

	for _, key := range []string{"SIT-2", "sit-2", "2"} {
		response := api.Get("/situations/" + key)
		s.Require().Equal(http.StatusOK, response.Code, "%s: %s", key, response.Body.String())
		s.Contains(response.Body.String(), `"reference":"SIT-2"`, key)
	}
	for _, key := range []string{"SIT-x", "SIT-3", "INC-2"} {
		s.Equal(http.StatusNotFound, api.Get("/situations/"+key).Code, key)
	}
}

func TestRaiseSituationErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		name       string
		serviceErr error
		status     int
	}{
		{"closed", errs.ErrConflict, http.StatusConflict},
		{"absent", errs.ErrNotFound, http.StatusNotFound},
		{"invalid", errs.ErrInvalidInput, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			id := uuid.New()
			service := mocks.NewMockSituationService(t)
			service.EXPECT().RaiseSituation(ctx, id, rez.RaiseSituationParams{StartInvestigation: true}).
				Return(nil, fmt.Errorf("raise: %w", tc.serviceErr))
			handler := newSituationsHandler(nil, service, nil, nil)
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
			huma.Register(api, oapi.ListSituations, newSituationsHandler(nil, service, nil, nil).ListSituations)

			response := api.Get("/situations" + tc.query)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		})
	}

	t.Run("invalid", func(t *testing.T) {
		service := mocks.NewMockSituationService(t)
		_, api := humatest.New(t)
		huma.Register(api, oapi.ListSituations, newSituationsHandler(nil, service, nil, nil).ListSituations)

		response := api.Get("/situations?muted=sometimes")
		require.Equal(t, http.StatusUnprocessableEntity, response.Code, response.Body.String())
	})
}
