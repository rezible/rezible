package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemanalysis"
)

type RetrospectiveServiceSuite struct {
	test.Suite
}

func TestRetrospectiveServiceSuite(t *testing.T) {
	suite.Run(t, &RetrospectiveServiceSuite{Suite: test.NewSuite()})
}

func (s *RetrospectiveServiceSuite) createIncident() *ent.Incident {
	ctx := s.SeedTenantContext()

	client := s.Database().Client(ctx)

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
		SetSeverityID(severity.ID).
		SetTypeID(incidentType.ID).
		Save(ctx)
	s.Require().NoError(err)
	return inc
}

func (s *RetrospectiveServiceSuite) TestCreateFullRetrospectiveCreatesSnapshotBackedAnalysis() {
	ctx := s.SeedTenantContext()
	svc := &RetrospectiveService{db: s.Database()}
	inc := s.createIncident()

	retro, err := svc.createForIncident(ctx, inc)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, retro.SystemAnalysisID)

	analysis, err := s.Database().Client(ctx).SystemAnalysis.Query().
		Where(systemanalysis.ID(retro.SystemAnalysisID)).
		WithKnowledgeGraphSnapshot().
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(analysis.KnowledgeGraphSnapshotID)

	//snapshot := analysis.Edges.KnowledgeGraphSnapshot
	//s.Require().NotNil(snapshot)
	//s.Equal("incident", snapshot.ScopeKind)
	//s.Equal(inc.ID.String(), snapshot.ScopeProperties["incidentId"])
}
