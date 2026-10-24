package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/internal/db"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type IncidentsHandlerSuite struct {
	test.Suite
}

func TestIncidentsHandlerSuite(t *testing.T) {
	suite.Run(t, &IncidentsHandlerSuite{Suite: test.NewSuite()})
}

// noJudge fails the test if a situation is judged.
type noJudge struct{ t *testing.T }

func (j noJudge) Run(context.Context, rezai.SituationJudgeInput) (rezai.SituationJudgeOutput, error) {
	j.t.Error("unexpected situation judgment")
	return rezai.SituationJudgeOutput{}, nil
}

func (s *IncidentsHandlerSuite) createResolvedIncident(ctx context.Context, client *ent.Client) *ent.Incident {
	severity := client.IncidentSeverity.Create().SetName("SEV-1").SetRank(1).SaveX(ctx)
	return client.Incident.Create().
		SetSlug(uuid.NewString()).
		SetTitle("API outage").
		SetSeverityID(severity.ID).
		SetResponseState(incident.ResponseStateResolved).
		SetResolvedAt(time.Now()).
		SaveX(ctx)
}

func (s *IncidentsHandlerSuite) TestReferencesToAnotherTenantAreNotFound() {
	_, tdb := s.SetupTestDatabase()
	ctx, identity := s.NewIdentity(tdb, "a")
	otherCtx, otherIdentity := s.NewIdentity(tdb, "b")
	client := tdb.Client(ctx)
	otherClient := tdb.Client(otherCtx)

	jobService := mocks.NewMockJobService(s.T())
	msgs := mocks.NewMockMessageQueue(s.T())
	orgs, orgsErr := db.NewOrganizationService(tdb, jobService)
	s.Require().NoError(orgsErr)
	situations, situationsErr := db.NewSituationService(rez.SituationsConfig{}, tdb, test.NewClock(time.Now()), jobService, msgs, nil, nil, nil, noJudge{t: s.T()})
	s.Require().NoError(situationsErr)
	retros, retrosErr := db.NewRetrospectiveService(tdb, jobService)
	s.Require().NoError(retrosErr)
	details, detailsErr := db.NewIncidentDetailsService(tdb, msgs, situations)
	s.Require().NoError(detailsErr)
	management, managementErr := db.NewIncidentManagementService(tdb, msgs, orgs, situations, retros, details)
	s.Require().NoError(managementErr)

	orgID := client.Organization.Query().OnlyIDX(ctx)
	_, enableErr := orgs.SetPreferences(ctx, orgID, func(m *ent.OrganizationPreferencesMutation) {
		m.SetEnableIncidentManagement(true)
	})
	s.Require().NoError(enableErr)
	inc := s.createResolvedIncident(ctx, client)
	role := client.IncidentRole.Create().SetName("Commander").SaveX(ctx)
	retro, retroErr := retros.CreateForIncident(ctx, inc.ID)
	s.Require().NoError(retroErr)
	_, ownerErr := retros.SetOwner(ctx, retro.ID, identity.Session.UserID)
	s.Require().NoError(ownerErr)

	otherInc := s.createResolvedIncident(otherCtx, otherClient)
	otherSeverity := otherClient.IncidentSeverity.Create().SetName("SEV-2").SetRank(2).SaveX(otherCtx)
	seed := otherClient.KnowledgeEntity.Create().SetCategory(knowledgeentity.CategoryContainer).SetKind("service").SaveX(otherCtx)
	otherSituation := otherClient.Situation.Create().SetTitle("errors").SetOpenedAt(time.Now()).SetSeedEntityID(seed.ID).SaveX(otherCtx)
	otherUserID := otherIdentity.Session.UserID

	_, api := humatest.New(s.T())
	api.UseMiddleware(func(hctx huma.Context, next func(huma.Context)) {
		next(huma.WithContext(hctx, execution.NewUserContext(hctx.Context(), identity.Session)))
	})
	incidents := newIncidentsHandler(nil, details, management)
	huma.Register(api, oapi.GetIncident, incidents.GetIncident)
	huma.Register(api, oapi.UpdateIncident, incidents.UpdateIncident)
	huma.Register(api, oapi.CreateIncidentRoleAssignment, incidents.CreateIncidentRoleAssignment)
	huma.Register(api, oapi.LinkIncidentSituation, incidents.LinkIncidentSituation)
	huma.Register(api, oapi.SetRetrospectiveOwner, newRetrospectivesHandler(retros).SetRetrospectiveOwner)

	attributes := func(attrs map[string]any) map[string]any { return map[string]any{"attributes": attrs} }
	for name, call := range map[string]func() *httptest.ResponseRecorder{
		"read other incident": func() *httptest.ResponseRecorder { return api.Get("/incidents/" + otherInc.ID.String()) },
		"link other situation": func() *httptest.ResponseRecorder {
			return api.Post("/incidents/"+inc.ID.String()+"/situations", attributes(map[string]any{"situationId": otherSituation.ID}))
		},
		"set other severity": func() *httptest.ResponseRecorder {
			return api.Patch("/incidents/"+inc.ID.String(), attributes(map[string]any{"severityId": otherSeverity.ID}))
		},
		"assign other user": func() *httptest.ResponseRecorder {
			return api.Post("/incidents/"+inc.ID.String()+"/role-assignments", attributes(map[string]any{"roleId": role.ID, "userId": otherUserID}))
		},
		"own with other user": func() *httptest.ResponseRecorder {
			return api.Put("/retrospectives/"+retro.ID.String()+"/owner", attributes(map[string]any{"userId": otherUserID}))
		},
	} {
		response := call()
		s.Equal(http.StatusNotFound, response.Code, "%s: %s", name, response.Body.String())
		s.NotContains(response.Body.String(), otherUserID.String(), name)
	}

	systemCtx := execution.NewSystemContext(s.T().Context())
	systemClient := tdb.Client(systemCtx)
	s.Equal(otherInc.SeverityID, systemClient.Incident.GetX(systemCtx, otherInc.ID).SeverityID)
	s.Zero(systemClient.Situation.QueryIncidents(otherSituation).CountX(systemCtx))
	s.Zero(systemClient.IncidentRoleAssignment.Query().Where(incidentroleassignment.UserID(otherUserID)).CountX(systemCtx))
	s.Equal(identity.Session.UserID, *client.Retrospective.GetX(ctx, retro.ID).OwnerID)
}
