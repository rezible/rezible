package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type PlaybookServiceProjectionSuite struct {
	test.Suite
}

func TestPlaybookServiceProjectionSuite(t *testing.T) {
	suite.Run(t, &PlaybookServiceProjectionSuite{Suite: test.NewSuite()})
}

func (s *PlaybookServiceProjectionSuite) TestPlaybookProjectionUpsertsByTitleAndLinksAlerts() {
	ctx := s.SeedTenantContext()
	svc, err := NewPlaybookService(s.Database())
	s.Require().NoError(err)

	alertEntity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(knowledgeEntityKindAlert).
		SetDisplayName("Search API response time high").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeEntityAlias.Create().
		SetEntityID(alertEntity.ID).
		SetProvider("test").
		SetProviderSubjectRef("demo:alert:search-api-latency").
		Save(ctx)
	s.Require().NoError(err)
	alert, err := s.Client(ctx).Alert.Create().
		SetKnowledgeEntityID(alertEntity.ID).
		SetTitle("Search API response time high").
		SetDescription("p95 latency is high").
		Save(ctx)
	s.Require().NoError(err)

	attrs := projections.PlaybookSubjectAttributes{
		Title:         "Checkout search latency triage",
		Content:       "Check Search API latency and Elasticsearch CPU.",
		RelatedAlerts: []string{"demo:alert:search-api-latency"},
	}
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	occurredAt := time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC)
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("playbooks").
		SetProviderEventRef("playbook-event-" + uuid.NewString()).
		SetProviderSubjectRef("demo:playbook:checkout-search-latency").
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindPlaybook.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)

	_, projErr := svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(projErr)

	_, proj2Err := svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(proj2Err)

	playbooks, err := s.Client(ctx).Playbook.Query().
		WithAlerts().
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(playbooks, 1)
	s.Equal(attrs.Title, playbooks[0].Title)
	s.Equal([]byte(attrs.Content), playbooks[0].Content)
	s.Require().Len(playbooks[0].Edges.Alerts, 1)
	s.Equal(alert.ID, playbooks[0].Edges.Alerts[0].ID)
}
