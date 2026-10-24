package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/discussionthread"
	incidentpkg "github.com/rezible/rezible/ent/incident"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/ent/systemanalysisentry"
	"github.com/rezible/rezible/pkg/errs"
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
		SetSeverity(severity).
		SetType(kind).
		SetResponseState(incidentpkg.ResponseStateResolved).
		SetResolvedAt(time.Now())
	incident, incidentErr := createIncident.Save(ctx)
	s.Require().NoError(incidentErr)

	return incident
}

func (s *DiscussionSuite) TestThreadRequiresExactlyOneOwnerAndReverseEdges() {
	ctx, db := s.SetupTestDatabase()
	client := db.Client(ctx)
	user, userErr := client.User.Query().Only(ctx)
	s.Require().NoError(userErr)

	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)

	incident := s.createIncident(client, ctx)
	retrospectives, retrospectiveServiceErr := NewRetrospectiveService(db, nil)
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
	ctx, db := s.SetupTestDatabase()
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

// createThreadTargets creates one thing of each kind a thread on the analysis can be about.
func (s *DiscussionSuite) createThreadTargets(ctx context.Context, tdb rez.Database, analysis *ent.SystemAnalysis) map[rez.DiscussionThreadTargetKind]uuid.UUID {
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	service := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:"+uuid.NewString())
	alert := graph.entity(kne.CategorySignal, "alert", "provider-a", "alert:"+uuid.NewString())
	observes := graph.relationship(knr.PredicateObserves, alert, service)
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		graph.observeEntity(service, "checkout"),
		graph.observeEntity(alert, "Checkout errors"),
		graph.observeRelationship(observes),
	)
	createEvent := client.NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("tests").
		SetProviderResourceRef(uuid.NewString()).
		SetProviderEventSource("tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind("test").
		SetOccurredAt(time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC)).
		SetReceivedAt(time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC)).
		SetAttributes([]byte(`{}`))
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	createEntry := client.SystemAnalysisEntry.Create().
		SetAnalysisID(analysis.ID).
		SetKind(systemanalysisentry.KindObservation).
		SetTitle("Checkout errors rose")
	entry, entryErr := createEntry.Save(ctx)
	s.Require().NoError(entryErr)

	createSession := client.AgentSession.Create().
		SetAgentName("investigation").
		SetInput([]byte(`{}`))
	session, sessionErr := createSession.Save(ctx)
	s.Require().NoError(sessionErr)
	createInvestigation := client.Investigation.Create().
		SetAgentSessionID(session.ID).
		SetSystemAnalysisID(analysis.ID)
	investigation, investigationErr := createInvestigation.Save(ctx)
	s.Require().NoError(investigationErr)
	createFinding := client.InvestigationFinding.Create().
		SetInvestigationID(investigation.ID).
		SetKey("answer")
	finding, findingErr := createFinding.Save(ctx)
	s.Require().NoError(findingErr)

	return map[rez.DiscussionThreadTargetKind]uuid.UUID{
		rez.DiscussionThreadTargetFinding:               finding.ID,
		rez.DiscussionThreadTargetEntry:                 entry.ID,
		rez.DiscussionThreadTargetKnowledgeEntity:       graph.entityID(ctx, service),
		rez.DiscussionThreadTargetKnowledgeRelationship: graph.relationshipID(ctx, observes),
		rez.DiscussionThreadTargetNormalizedEvent:       event.ID,
	}
}

// threadTargetColumns are the columns holding each kind of target.
var threadTargetColumns = map[rez.DiscussionThreadTargetKind]string{
	rez.DiscussionThreadTargetFinding:               discussionthread.FieldTargetFindingID,
	rez.DiscussionThreadTargetEntry:                 discussionthread.FieldTargetEntryID,
	rez.DiscussionThreadTargetKnowledgeEntity:       discussionthread.FieldTargetKnowledgeEntityID,
	rez.DiscussionThreadTargetKnowledgeRelationship: discussionthread.FieldTargetKnowledgeRelationshipID,
	rez.DiscussionThreadTargetNormalizedEvent:       discussionthread.FieldTargetNormalizedEventID,
}

func (s *DiscussionSuite) TestThreadHasAtMostOneTarget() {
	ctx, db := s.SetupTestDatabase()
	client := db.Client(ctx)
	user, userErr := client.User.Query().Only(ctx)
	s.Require().NoError(userErr)
	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)
	targets := s.createThreadTargets(ctx, db, analysis)

	createTwoTargetThread := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis).
		SetTargetEntryID(targets[rez.DiscussionThreadTargetEntry]).
		SetTargetKnowledgeEntityID(targets[rez.DiscussionThreadTargetKnowledgeEntity])
	_, twoTargetsErr := createTwoTargetThread.Save(ctx)
	s.Require().Error(twoTargetsErr)
	s.True(ent.IsConstraintError(twoTargetsErr))
}

func (s *DiscussionSuite) TestListThreadsByTarget() {
	ctx, db := s.SetupTestDatabase()
	client := db.Client(ctx)
	user, userErr := client.User.Query().Only(ctx)
	s.Require().NoError(userErr)
	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)
	targets := s.createThreadTargets(ctx, db, analysis)

	threads := make(map[rez.DiscussionThreadTargetKind]uuid.UUID, len(targets))
	for kind, targetID := range targets {
		createThread := client.DiscussionThread.Create().
			SetUser(user).
			SetAnalysis(analysis)
		s.Require().NoError(createThread.Mutation().SetField(threadTargetColumns[kind], targetID))
		thread, createErr := createThread.Save(ctx)
		s.Require().NoError(createErr)
		threads[kind] = thread.ID
	}
	createUntargeted := client.DiscussionThread.Create().
		SetUser(user).
		SetAnalysis(analysis)
	s.Require().NoError(createUntargeted.Exec(ctx))

	service := NewDiscussionService(db)
	listThreadIDs := func(kind rez.DiscussionThreadTargetKind, targetID uuid.UUID) []uuid.UUID {
		params := rez.ListDiscussionThreadsParams{AnalysisID: analysis.ID, TargetKind: kind, TargetID: targetID}
		result, listErr := service.ListThreads(ctx, params)
		s.Require().NoError(listErr)
		ids := make([]uuid.UUID, len(result.Data))
		for i, thread := range result.Data {
			ids[i] = thread.ID
		}
		return ids
	}

	s.Len(listThreadIDs("", uuid.Nil), len(targets)+1)
	for kind, targetID := range targets {
		s.Run(string(kind), func() {
			s.Equal([]uuid.UUID{threads[kind]}, listThreadIDs(kind, uuid.Nil))
			s.Equal([]uuid.UUID{threads[kind]}, listThreadIDs(kind, targetID))
			s.Empty(listThreadIDs(kind, uuid.New()))
		})
	}

	s.Run("a target ID needs its kind", func() {
		params := rez.ListDiscussionThreadsParams{AnalysisID: analysis.ID, TargetID: targets[rez.DiscussionThreadTargetEntry]}
		_, listErr := service.ListThreads(ctx, params)
		s.ErrorIs(listErr, errs.ErrInvalidInput)
	})
	s.Run("an unknown kind is refused", func() {
		params := rez.ListDiscussionThreadsParams{AnalysisID: analysis.ID, TargetKind: "comment"}
		_, listErr := service.ListThreads(ctx, params)
		s.ErrorIs(listErr, errs.ErrInvalidInput)
	})
}
