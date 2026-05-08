package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/testkit"
)

type RetrospectiveServiceSuite struct {
	testkit.Suite
}

func TestRetrospectiveServiceSuite(t *testing.T) {
	suite.Run(t, &RetrospectiveServiceSuite{Suite: testkit.NewSuite()})
}

func (s *RetrospectiveServiceSuite) createIncident() *ent.Incident {
	ctx := s.SeedTenantContext()

	severity, err := s.Client().IncidentSeverity.Create().
		SetName("SEV-1 " + uuid.NewString()).
		SetRank(1).
		SetDescription("Critical").
		Save(ctx)
	s.Require().NoError(err)

	incidentType, err := s.Client().IncidentType.Create().
		SetName("Customer Impact " + uuid.NewString()).
		Save(ctx)
	s.Require().NoError(err)

	inc, err := s.Client().Incident.Create().
		SetSlug("incident-" + uuid.NewString()).
		SetTitle("API outage").
		SetSeverityID(severity.ID).
		SetTypeID(incidentType.ID).
		Save(ctx)
	s.Require().NoError(err)
	return inc
}

func (s *RetrospectiveServiceSuite) TestCreateFullRetrospectiveCreatesSnapshotBackedAnalysis() {
	ctx := s.SeedTenantContext()
	svc := &RetrospectiveService{db: s.Client()}
	inc := s.createIncident()

	retro, err := svc.createForIncident(ctx, inc)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, retro.SystemAnalysisID)

	analysis, err := s.Client().SystemAnalysis.Query().
		Where(systemanalysis.ID(retro.SystemAnalysisID)).
		WithTopologySnapshot().
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(analysis.TopologySnapshotID)
	s.Require().NotNil(analysis.Edges.TopologySnapshot)
	s.Equal("incident", analysis.Edges.TopologySnapshot.Scope.String())
	s.Equal(inc.ID.String(), analysis.Edges.TopologySnapshot.ScopeProperties["incidentId"])
}
