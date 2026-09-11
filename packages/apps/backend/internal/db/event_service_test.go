package db

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type EventsServiceSuite struct {
	test.Suite
}

func TestEventsServiceSuite(t *testing.T) {
	suite.Run(t, &EventsServiceSuite{Suite: test.NewSuite()})
}

func (s *EventsServiceSuite) TestListEventsFiltersThroughObservationGroupsWithoutDuplicates() {
	ctx := s.SeedTenantContext()
	db := s.CreateTestDatabase()
	client := db.Client(ctx)
	situationID := uuid.New()
	knowledgeEntityID := uuid.New()
	_, sitErr := client.Situation.Create().SetID(situationID).SetKnowledgeEntityID(knowledgeEntityID).SetTitle("Situation").SetOpenedAt(time.Now()).Save(ctx)
	s.Require().NoError(sitErr)
	event := client.NormalizedEvent.Create().SetKind("incident").SetProviderEventSource("test").SetProviderEventRef(uuid.NewString()).SetAttributes(json.RawMessage(`{}`)).SetOccurredAt(time.Now()).SetReceivedAt(time.Now()).SaveX(ctx)
	for range 2 {
		client.SituationObservationGroup.Create().SetSituationID(situationID).SetTitle("Observations").AddEventIDs(event.ID).SaveX(ctx)
	}
	service, serviceErr := NewEventsService(db)
	s.Require().NoError(serviceErr)
	result, listErr := service.ListEvents(ctx, rez.ListEventsParams{SituationID: situationID})
	s.Require().NoError(listErr)
	s.Require().Len(result.Items, 1)
	s.Require().Equal(event.ID, result.Items[0].ID)
}
