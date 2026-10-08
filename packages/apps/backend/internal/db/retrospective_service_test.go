package db

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/pkg/errs"

	"github.com/google/uuid"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
)

type RetrospectiveServiceSuite struct {
	test.Suite
}

func TestRetrospectiveServiceSuite(t *testing.T) {
	suite.Run(t, &RetrospectiveServiceSuite{Suite: test.NewSuite()})
}

func (s *RetrospectiveServiceSuite) createIncident(ctx context.Context, client *ent.Client) *ent.Incident {
	severity, err := client.IncidentSeverity.Create().
		SetName("SEV-1 " + uuid.NewString()).
		SetRank(1).
		SetDescription("Critical").
		Save(ctx)
	s.Require().NoError(err)

	incidentType, err := client.IncidentType.Create().
		SetName("Customer Impact " + uuid.NewString()).
		Save(ctx)
	s.Require().NoError(err)

	inc, err := client.Incident.Create().
		SetSlug("incident-" + uuid.NewString()).
		SetTitle("API outage").
		SetResponseState(incident.ResponseStateResolved).
		SetSeverityID(severity.ID).
		SetTypeID(incidentType.ID).
		Save(ctx)
	s.Require().NoError(err)
	return inc
}

func (s *RetrospectiveServiceSuite) TestCreateFullRetrospective() {
	ctx, tdb := s.SetupTestDatabase()
	svc := &RetrospectiveService{db: tdb}

	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)

	retro, err := svc.CreateForIncident(ctx, inc.ID)
	s.Require().NoError(err)
	s.NotNil(retro)
	s.NotEqual(uuid.Nil, retro.DocumentID)
	s.NotEqual(uuid.Nil, retro.SystemAnalysisID)
	_, analysisErr := client.SystemAnalysis.Get(ctx, retro.SystemAnalysisID)
	s.Require().NoError(analysisErr)
}

func (s *RetrospectiveServiceSuite) TestRejectUnresolvedIncident() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	updateIncident := client.Incident.UpdateOneID(inc.ID).SetResponseState(incident.ResponseStateStarted)
	s.Require().NoError(updateIncident.Exec(ctx))
	svc := &RetrospectiveService{db: tdb}
	retro, createErr := svc.CreateForIncident(ctx, inc.ID)
	s.ErrorIs(createErr, errs.ErrConflict)
	s.Nil(retro)
	s.Zero(client.Retrospective.Query().CountX(ctx))
	s.Zero(client.Document.Query().CountX(ctx))
	s.Zero(client.SystemAnalysis.Query().CountX(ctx))
}

func (s *RetrospectiveServiceSuite) TestConcurrentCreationReusesWorkspace() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	svc := &RetrospectiveService{db: tdb}
	const callers = 8
	results := make([]*ent.Retrospective, callers)
	failures := make([]error, callers)
	var callersDone sync.WaitGroup
	start := make(chan struct{})
	for i := range callers {
		callersDone.Go(func() {
			<-start
			results[i], failures[i] = svc.CreateForIncident(ctx, inc.ID)
		})
	}
	close(start)
	callersDone.Wait()
	for i := range callers {
		s.Require().NoError(failures[i])
		s.Require().NotNil(results[i])
		s.Equal(results[0].ID, results[i].ID)
		s.Equal(retrospective.StateDraft, results[i].State)
	}
	s.Equal(1, client.Retrospective.Query().CountX(ctx))
	s.Equal(1, client.Document.Query().CountX(ctx))
	s.Equal(1, client.SystemAnalysis.Query().CountX(ctx))
}

func (s *RetrospectiveServiceSuite) TestCreationRollsBackWithTransaction() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	svc := &RetrospectiveService{db: tdb}
	abort := fmt.Errorf("abort incident update")
	txErr := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		retro, createErr := svc.CreateForIncident(txCtx, inc.ID)
		s.Require().NoError(createErr)
		s.Require().NotNil(retro)
		return abort
	})
	s.ErrorIs(txErr, abort)
	s.Zero(client.Retrospective.Query().CountX(ctx))
	s.Zero(client.Document.Query().CountX(ctx))
	s.Zero(client.SystemAnalysis.Query().CountX(ctx))
}
