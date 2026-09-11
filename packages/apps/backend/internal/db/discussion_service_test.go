package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/discussioncomment"
	"github.com/rezible/rezible/ent/discussionthread"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DiscussionSuite struct{ test.Suite }

func TestDiscussionSuite(t *testing.T) { suite.Run(t, &DiscussionSuite{Suite: test.NewSuite()}) }

func (s *DiscussionSuite) createIncident(client *ent.Client, ctx context.Context) *ent.Incident {
	severity, severityErr := client.IncidentSeverity.Create().SetName(uuid.NewString()).SetRank(1).SetDescription("critical").Save(ctx)
	s.Require().NoError(severityErr)
	kind, kindErr := client.IncidentType.Create().SetName(uuid.NewString()).Save(ctx)
	s.Require().NoError(kindErr)
	incident, incidentErr := client.Incident.Create().SetSlug(uuid.NewString()).SetTitle("outage").SetSeverity(severity).SetType(kind).Save(ctx)
	s.Require().NoError(incidentErr)
	return incident
}

func (s *DiscussionSuite) TestThreadRequiresExactlyOneOwnerAndReverseEdges() {
	ctx := s.SeedTenantContext()
	db := s.CreateTestDatabase()
	client := db.Client(ctx)
	user, userErr := client.User.Query().Only(ctx)
	s.Require().NoError(userErr)
	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)
	incident := s.createIncident(client, ctx)
	retro, retroErr := (&RetrospectiveService{db: db}).createForIncident(ctx, incident)
	s.Require().NoError(retroErr)

	_, bothErr := client.DiscussionThread.Create().SetUser(user).SetAnalysis(analysis).SetRetrospective(retro).SetKind(discussionthread.KindComment).Save(ctx)
	s.Error(bothErr)
	_, neitherErr := client.DiscussionThread.Create().SetUser(user).SetKind(discussionthread.KindComment).Save(ctx)
	s.Error(neitherErr)

	thread, createErr := client.DiscussionThread.Create().SetUser(user).SetAnalysis(analysis).SetKind(discussionthread.KindComment).Save(ctx)
	s.Require().NoError(createErr)
	fetchedAnalysis, fetchErr := client.SystemAnalysis.Query().Where(systemanalysis.ID(analysis.ID)).WithDiscussionThreads().Only(ctx)
	s.Require().NoError(fetchErr)
	s.Equal(thread.ID, fetchedAnalysis.Edges.DiscussionThreads[0].ID)
	fetchedThread, fetchThreadErr := client.DiscussionThread.Query().Where(discussionthread.ID(thread.ID)).WithAnalysis().Only(ctx)
	s.Require().NoError(fetchThreadErr)
	s.Equal(analysis.ID, fetchedThread.Edges.Analysis.ID)
}

func (s *DiscussionSuite) TestCommentListIsThreadScopedAndPaginated() {
	ctx := s.SeedTenantContext()
	db := s.CreateTestDatabase()
	client := db.Client(ctx)
	user, userErr := client.User.Query().Only(ctx)
	s.Require().NoError(userErr)
	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)
	thread, createErr := client.DiscussionThread.Create().SetUser(user).SetAnalysis(analysis).SetKind(discussionthread.KindComment).Save(ctx)
	s.Require().NoError(createErr)
	for range 3 {
		_, commentErr := client.DiscussionComment.Create().SetThread(thread).SetUser(user).SetContent(uuid.NewString()).Save(ctx)
		s.Require().NoError(commentErr)
	}
	query := client.DiscussionComment.Query().Where(discussioncomment.ThreadID(thread.ID))
	result, listErr := ent.DoListQuery[ent.DiscussionComment, *ent.DiscussionCommentQuery](ctx, query, ent.ListParams{Page: 2, PageSize: 2})
	s.Require().NoError(listErr)
	s.Len(result.Data, 1)
	s.Equal(3, result.Total)
}
