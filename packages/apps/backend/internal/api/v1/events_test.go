package apiv1

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type EventsHandlerSuite struct {
	test.Suite
}

func TestEventsHandlerSuite(t *testing.T) {
	suite.Run(t, &EventsHandlerSuite{Suite: test.NewSuite()})
}

// An event read with its projection carries the projection receipt and the entities it recorded.
func (s *EventsHandlerSuite) TestEventsCarryTheirProjection() {
	_, tdb := s.SetupTestDatabase()
	ctx, identity := s.NewIdentity(tdb, "a")
	client := tdb.Client(ctx)
	at := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := client.NormalizedEvent.Create().
		SetKind("deployment").
		SetProvider("test").
		SetProviderNamespace("events").
		SetProviderResourceRef("deployment:1").
		SetProviderEventSource("test").
		SetProviderEventRef("event:1").
		SetAttributes([]byte(`{}`)).
		SetOccurredAt(at).
		SetLastOccurredAt(at).
		SetReceivedAt(at)
	event := createEvent.SaveX(ctx)
	receipt := client.NormalizedEventProjection.Create().SetEventID(event.ID).SaveX(ctx)
	entityID := uuid.New()
	createProjected := client.NormalizedEventProjectionEntity.Create().
		SetProjectionID(receipt.ID).
		SetDomainEntityKind("change_point").
		SetDomainEntityID(entityID)
	createProjected.ExecX(ctx)

	events, eventsErr := db.NewEventsService(tdb, mocks.NewMockSituationService(s.T()))
	s.Require().NoError(eventsErr)
	_, api := humatest.New(s.T())
	api.UseMiddleware(func(hctx huma.Context, next func(huma.Context)) {
		next(huma.WithContext(hctx, execution.NewUserContext(hctx.Context(), identity.Session)))
	})
	handler := newEventsHandler(nil, events)
	huma.Register(api, oapi.GetEvent, handler.GetEvent)
	huma.Register(api, oapi.ListEvents, handler.ListEvents)

	assertProjection := func(got oapi.Event) {
		s.Require().NotNil(got.Attributes.Projection)
		s.Equal(receipt.ID, got.Attributes.Projection.Id)
		s.Equal([]oapi.EventProjectionEntity{{EntityId: entityID, EntityKind: "change_point"}}, got.Attributes.Projection.Attributes.Entities)
	}

	getResponse := api.Get("/events/" + event.ID.String() + "?withProjection=true")
	s.Require().Equal(http.StatusOK, getResponse.Code, getResponse.Body.String())
	var got oapi.GetEventResponse
	s.Require().NoError(json.Unmarshal(getResponse.Body.Bytes(), &got.Body))
	assertProjection(got.Body.Data)

	listResponse := api.Get("/events?withProjection=true")
	s.Require().Equal(http.StatusOK, listResponse.Code, listResponse.Body.String())
	var listed oapi.ListEventsResponse
	s.Require().NoError(json.Unmarshal(listResponse.Body.Bytes(), &listed.Body))
	s.Require().Len(listed.Body.Data, 1)
	assertProjection(listed.Body.Data[0])
}
