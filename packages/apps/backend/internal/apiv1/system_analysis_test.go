package apiv1

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/systemanalysisnode"
	"github.com/rezible/rezible/ent/topologysnapshot"
	oapi "github.com/rezible/rezible/openapi/v1"
	"github.com/rezible/rezible/testkit"
)

type SystemAnalysisHandlerSuite struct {
	testkit.Suite
}

func TestSystemAnalysisHandlerSuite(t *testing.T) {
	suite.Run(t, &SystemAnalysisHandlerSuite{Suite: testkit.NewSuite()})
}

func (s *SystemAnalysisHandlerSuite) createAnalysis() *ent.SystemAnalysis {
	ctx := s.SeedTenantContext()
	retro := s.createRetrospective()
	analysis, err := s.Client().SystemAnalysis.Create().
		SetRetrospective(retro).
		Save(ctx)
	s.Require().NoError(err)
	return analysis
}

func (s *SystemAnalysisHandlerSuite) createSnapshotBackedAnalysis() (*ent.SystemAnalysis, *ent.TopologySnapshotEntity) {
	ctx := s.SeedTenantContext()
	retro := s.createRetrospective()
	snapshot, err := s.Client().TopologySnapshot.Create().
		SetScope(topologysnapshot.ScopeAnalysis).
		SetScopeProperties(map[string]any{"test": true}).
		Save(ctx)
	s.Require().NoError(err)
	analysis, err := s.Client().SystemAnalysis.Create().
		SetRetrospective(retro).
		SetTopologySnapshotID(snapshot.ID).
		Save(ctx)
	s.Require().NoError(err)
	snapshotEntity, err := s.Client().TopologySnapshotEntity.Create().
		SetSnapshotID(snapshot.ID).
		SetEntityKind("service").
		SetDisplayName("Existing service").
		SetProperties(map[string]any{}).
		SetAliases([]map[string]any{}).
		Save(ctx)
	s.Require().NoError(err)
	return analysis, snapshotEntity
}

func (s *SystemAnalysisHandlerSuite) createRetrospective() *ent.Retrospective {
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

	doc, err := s.Client().Document.Create().
		SetContent([]byte("")).
		SetAccessRestricted(false).
		Save(ctx)
	s.Require().NoError(err)

	retro, err := s.Client().Retrospective.Create().
		SetIncident(inc).
		SetDocument(doc).
		SetKind(retrospective.KindFull).
		SetState(retrospective.StateDraft).
		Save(ctx)
	s.Require().NoError(err)
	return retro
}

func (s *SystemAnalysisHandlerSuite) createKnowledgeEntity() *ent.KnowledgeEntity {
	entity, err := s.Client().KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("Payments API " + uuid.NewString()).
		SetDescription("Handles payments").
		SetProperties(map[string]any{"tier": "backend"}).
		Save(s.SeedTenantContext())
	s.Require().NoError(err)
	return entity
}

func (s *SystemAnalysisHandlerSuite) TestAddSystemAnalysisNodeCopiesKnowledgeEntityIntoSnapshot() {
	ctx := s.SeedTenantContext()
	handler := newSystemAnalysisHandler(s.Client())
	analysis := s.createAnalysis()
	entity := s.createKnowledgeEntity()

	request := &oapi.AddSystemAnalysisNodeRequest{Id: analysis.ID}
	request.Body.Attributes = oapi.AddSystemAnalysisNodeAttributes{
		KnowledgeEntityId: &entity.ID,
		Position:          oapi.SystemAnalysisDiagramPosition{X: 10, Y: 20},
		Description:       "Important service",
	}
	resp, err := handler.AddSystemAnalysisNode(ctx, request)
	s.Require().NoError(err)
	s.Equal(entity.DisplayName, resp.Body.Data.Attributes.SnapshotEntity.Attributes.DisplayName)
	s.Require().NotNil(resp.Body.Data.Attributes.SnapshotEntity.Attributes.KnowledgeEntityId)
	s.Equal(entity.ID, *resp.Body.Data.Attributes.SnapshotEntity.Attributes.KnowledgeEntityId)

	analysisNodes, err := s.Client().SystemAnalysisNode.Query().
		Where(systemanalysisnode.AnalysisID(analysis.ID)).
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(analysisNodes, 1)
}

func (s *SystemAnalysisHandlerSuite) TestAddSystemAnalysisNodeAcceptsExistingSnapshotEntityID() {
	ctx := s.SeedTenantContext()
	handler := newSystemAnalysisHandler(s.Client())
	analysis, snapshotEntity := s.createSnapshotBackedAnalysis()

	request := &oapi.AddSystemAnalysisNodeRequest{Id: analysis.ID}
	request.Body.Attributes = oapi.AddSystemAnalysisNodeAttributes{
		SnapshotEntityId: &snapshotEntity.ID,
		Position:         oapi.SystemAnalysisDiagramPosition{X: 1, Y: 2},
	}
	resp, err := handler.AddSystemAnalysisNode(ctx, request)
	s.Require().NoError(err)
	s.Equal(snapshotEntity.ID, resp.Body.Data.Attributes.SnapshotEntity.Id)
}
