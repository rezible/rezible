package db

import (
	"testing"

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

func (s *RetrospectiveServiceSuite) createIncident(client *ent.Client) *ent.Incident {
	ctx := s.SeedTenantContext()

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

func (s *RetrospectiveServiceSuite) TestCreateFullRetrospective() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := &RetrospectiveService{db: tdb, incidents: &IncidentService{db: tdb}}

	client := tdb.Client(ctx)
	inc := s.createIncident(client)

	retro, err := svc.CreateForIncident(ctx, inc.ID)
	s.Require().NoError(err)
	s.NotNil(retro)
	s.NotEqual(uuid.Nil, retro.DocumentID)
	s.NotEqual(uuid.Nil, retro.SystemAnalysisID)
	_, analysisErr := client.SystemAnalysis.Get(ctx, retro.SystemAnalysisID)
	s.Require().NoError(analysisErr)
}
