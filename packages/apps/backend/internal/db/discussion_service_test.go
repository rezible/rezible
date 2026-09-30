package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/discussionthread"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DiscussionSuite struct{ test.Suite }

func TestDiscussionSuite(t *testing.T) {
	suite.Run(t, &DiscussionSuite{
		Suite: test.NewSuite(),
	})
}

func (s *DiscussionSuite) createIncident(client *ent.Client, ctx context.Context) *ent.Incident {
	createSeverity := client.IncidentSeverity.Create().
		SetName(uuid.NewString()).
		SetRank(1).
		SetDescription("critical")
	severity, severityErr := createSeverity.Save(ctx)
	s.Require().NoError(severityErr)

	createKind := client.IncidentType.Create().
		SetName(uuid.NewString())
	kind, kindErr := createKind.Save(ctx)
	s.Require().NoError(kindErr)

	createIncident := client.Incident.Create().
		SetSlug(uuid.NewString()).
		SetTitle("outage").
		SetResponseState(incident.ResponseStateResolved).
		SetSeverity(severity).
		SetType(kind)
	incident, incidentErr := createIncident.Save(ctx)
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
	retrospectives, retrospectiveServiceErr := NewRetrospectiveService(db)
	s.Require().NoError(retrospectiveServiceErr)

	retro, retroErr := retrospectives.CreateForIncident(ctx, incident.ID)
	s.Require().NoError(retroErr)

	createMultipleOwnerThread := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis).
		SetRetrospective(retro)
	_, bothErr := createMultipleOwnerThread.Save(ctx)
	s.Error(bothErr)
	createOwnerlessThread := client.DiscussionThread.Create().
		SetUser(user)
	_, neitherErr := createOwnerlessThread.Save(ctx)
	s.Error(neitherErr)

	createThread := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis)
	thread, createErr := createThread.Save(ctx)
	s.Require().NoError(createErr)

	queryFetchedAnalysis := client.SystemAnalysis.Query().
		Where(systemanalysis.ID(analysis.ID)).
		WithDiscussionThreads()
	fetchedAnalysis, fetchErr := queryFetchedAnalysis.Only(ctx)
	s.Require().NoError(fetchErr)
	s.Require().Len(fetchedAnalysis.Edges.DiscussionThreads, 1)
	s.Equal(thread.ID, fetchedAnalysis.Edges.DiscussionThreads[0].ID)
	queryFetchedThread := client.DiscussionThread.Query().
		Where(discussionthread.ID(thread.ID)).
		WithAnalysis()
	fetchedThread, fetchThreadErr := queryFetchedThread.Only(ctx)
	s.Require().NoError(fetchThreadErr)
	s.Require().NotNil(fetchedThread.Edges.Analysis)
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

	createThread := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis)
	thread, createErr := createThread.Save(ctx)
	s.Require().NoError(createErr)

	createOtherThread := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis)
	otherThread, otherThreadErr := createOtherThread.Save(ctx)
	s.Require().NoError(otherThreadErr)

	baseTime := time.Date(2026, 6, 4, 9, 30, 0, 0, time.UTC)
	comments := make([]*ent.DiscussionComment, 3)
	// Insert out of chronological order so insertion order cannot satisfy the assertion.
	for _, index := range []int{2, 0, 1} {
		createComment := client.DiscussionComment.Create().
			SetThread(thread).
			SetUser(user).
			SetContent(uuid.NewString()).
			SetCreatedAt(baseTime.Add(time.Duration(index) * time.Minute))
		comment, commentErr := createComment.Save(ctx)
		s.Require().NoError(commentErr)
		comments[index] = comment
	}
	createOtherComment := client.DiscussionComment.Create().
		SetThread(otherThread).
		SetUser(user).
		SetContent("other thread").
		SetCreatedAt(baseTime.Add(30 * time.Second))
	otherComment, otherCommentErr := createOtherComment.Save(ctx)
	s.Require().NoError(otherCommentErr)

	service := NewDiscussionService(db)
	params := rez.ListDiscussionCommentsParams{
		ThreadID: thread.ID,
		ListParams: ent.ListParams{
			Page:     1,
			PageSize: 2,
		},
	}

	firstPage, firstPageErr := service.ListComments(ctx, params)
	s.Require().NoError(firstPageErr)
	s.Equal(3, firstPage.Total)
	s.Require().Len(firstPage.Data, 2)
	s.Equal(comments[0].ID, firstPage.Data[0].ID)
	s.Equal(comments[1].ID, firstPage.Data[1].ID)

	params.Page = 2
	secondPage, secondPageErr := service.ListComments(ctx, params)
	s.Require().NoError(secondPageErr)
	s.Equal(3, secondPage.Total)
	s.Require().Len(secondPage.Data, 1)
	s.Equal(comments[2].ID, secondPage.Data[0].ID)
	for _, page := range []*ent.ListResult[ent.DiscussionComment]{firstPage, secondPage} {
		for _, comment := range page.Data {
			s.Equal(thread.ID, comment.ThreadID)
			s.NotEqual(otherComment.ID, comment.ID)
		}
	}
}
