package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ifo "github.com/rezible/rezible/ent/incidentfieldoption"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type IncidentManagementServiceSuite struct {
	test.Suite
}

func TestIncidentManagementServiceSuite(t *testing.T) {
	suite.Run(t, &IncidentManagementServiceSuite{Suite: test.NewSuite()})
}

func (s *IncidentManagementServiceSuite) newService(tdb rez.Database) *IncidentManagementService {
	msgs := mocks.NewMockMessageQueue(s.T())
	msgs.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Maybe()
	situations := mocks.NewMockSituationService(s.T())
	situations.EXPECT().SyncIncidentLinks(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	retrospectives, retrospectivesErr := NewRetrospectiveService(tdb, nil)
	s.Require().NoError(retrospectivesErr)
	details, detailsErr := NewIncidentDetailsService(tdb, msgs, situations)
	s.Require().NoError(detailsErr)
	orgs, orgsErr := NewOrganizationService(tdb, nil)
	s.Require().NoError(orgsErr)
	svc, svcErr := NewIncidentManagementService(tdb, msgs, orgs, situations, retrospectives, details)
	s.Require().NoError(svcErr)
	return svc
}

func (s *IncidentManagementServiceSuite) enable(ctx context.Context, tdb rez.Database) {
	organizationID, organizationErr := tdb.Client(ctx).Organization.Query().OnlyID(ctx)
	s.Require().NoError(organizationErr)
	orgs, orgsErr := NewOrganizationService(tdb, mocks.NewMockJobService(s.T()))
	s.Require().NoError(orgsErr)
	_, enableErr := orgs.SetPreferences(ctx, organizationID, func(m *ent.OrganizationPreferencesMutation) {
		m.SetEnableIncidentManagement(true)
	})
	s.Require().NoError(enableErr)
}

func (s *IncidentManagementServiceSuite) declare(ctx context.Context, tdb rez.Database, svc *IncidentManagementService) *ent.Incident {
	declarerID, declarerErr := tdb.Client(ctx).User.Query().FirstID(ctx)
	s.Require().NoError(declarerErr)
	declared, declareErr := svc.Declare(ctx, rez.DeclareIncidentParams{Title: "Checkout outage", DeclaredByID: declarerID})
	s.Require().NoError(declareErr)
	return declared
}

func (s *IncidentManagementServiceSuite) TestDeclareNeedsIncidentManagement() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.newService(tdb)
	declarerID, declarerErr := tdb.Client(ctx).User.Query().FirstID(ctx)
	s.Require().NoError(declarerErr)

	_, declareErr := svc.Declare(ctx, rez.DeclareIncidentParams{Title: "Checkout outage", DeclaredByID: declarerID})
	s.Equal(errs.CodeConflict, errs.CodeOf(declareErr))
	s.Equal("Incident management is off for this organization.", errs.PublicMessage(declareErr))
	s.Zero(tdb.Client(ctx).Incident.Query().CountX(ctx))
}

func (s *IncidentManagementServiceSuite) TestDeclareRecordsDetailsAndDeclarer() {
	ctx, tdb := s.SetupTestDatabase()
	s.enable(ctx, tdb)
	svc := s.newService(tdb)
	client := tdb.Client(ctx)

	createTag := client.IncidentTag.Create().
		SetKey("service").
		SetValue("api")
	tag, tagErr := createTag.Save(ctx)
	s.Require().NoError(tagErr)
	field, fieldErr := client.IncidentField.Create().SetName("Environment").Save(ctx)
	s.Require().NoError(fieldErr)
	createOption := client.IncidentFieldOption.Create().
		SetIncidentFieldID(field.ID).
		SetType(ifo.TypeCustom).
		SetValue("production")
	option, optionErr := createOption.Save(ctx)
	s.Require().NoError(optionErr)
	declarerID, declarerErr := client.User.Query().FirstID(ctx)
	s.Require().NoError(declarerErr)

	params := rez.DeclareIncidentParams{
		Title:          "API outage",
		Summary:        "Customer requests are failing",
		TagIDs:         []uuid.UUID{tag.ID},
		FieldOptionIDs: []uuid.UUID{option.ID},
		DeclaredByID:   declarerID,
	}
	declared, declareErr := svc.Declare(ctx, params)
	s.Require().NoError(declareErr)

	s.Equal("API outage", declared.Title)
	s.Nil(declared.SeverityID, "severity is optional")
	s.Nil(declared.TypeID, "type is optional")
	s.Require().NotNil(declared.DeclaredByID)
	s.Equal(declarerID, *declared.DeclaredByID)
	s.Require().Len(declared.Edges.TagAssignments, 1)
	s.Equal(tag.ID, declared.Edges.TagAssignments[0].ID)
	s.Require().Len(declared.Edges.FieldSelections, 1)
	s.Equal(option.ID, declared.Edges.FieldSelections[0].ID)
	s.Equal(incident.ResponseStateStarted, declared.ResponseState)
}

func (s *IncidentManagementServiceSuite) TestResponseState() {
	svc := &IncidentManagementService{}
	at := func(minutes int) time.Time {
		return time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC).Add(time.Duration(minutes) * time.Minute)
	}
	milestone := func(id string, kind im.Kind, minutes int) *ent.IncidentMilestone {
		return &ent.IncidentMilestone{ID: uuid.MustParse(id), Kind: kind, Timestamp: at(minutes)}
	}
	const (
		lowID  = "00000000-0000-0000-0000-000000000001"
		highID = "00000000-0000-0000-0000-000000000002"
	)
	for _, tc := range []struct {
		name       string
		milestones ent.IncidentMilestones
		want       incident.ResponseState
	}{
		{"none", nil, incident.ResponseStateStarted},
		{"only detection", ent.IncidentMilestones{milestone(lowID, im.KindDetection, 0)}, incident.ResponseStateStarted},
		{"impact", ent.IncidentMilestones{milestone(lowID, im.KindImpact, 0)}, incident.ResponseStateStarted},
		{"mitigation", ent.IncidentMilestones{
			milestone(lowID, im.KindImpact, 0),
			milestone(highID, im.KindMitigation, 10),
		}, incident.ResponseStateMitigated},
		{"an impact after a mitigation", ent.IncidentMilestones{
			milestone(highID, im.KindImpact, 20),
			milestone(lowID, im.KindMitigation, 10),
		}, incident.ResponseStateStarted},
		{"an impact and a mitigation at the same time, by id", ent.IncidentMilestones{
			milestone(highID, im.KindMitigation, 10),
			milestone(lowID, im.KindImpact, 10),
		}, incident.ResponseStateMitigated},
		{"a resolution", ent.IncidentMilestones{
			milestone(lowID, im.KindMitigation, 10),
			milestone(highID, im.KindResolution, 20),
		}, incident.ResponseStateResolved},
	} {
		s.Run(tc.name, func() {
			state, _ := svc.responseState(tc.milestones)
			s.Equal(tc.want, state)
		})
	}
}

func (s *IncidentManagementServiceSuite) TestImpactAfterMitigationStoresStarted() {
	ctx, tdb := s.SetupTestDatabase()
	s.enable(ctx, tdb)
	svc := s.newService(tdb)
	declared := s.declare(ctx, tdb, svc)
	add := func(kind im.Kind, minutes int) {
		params := rez.AddIncidentMilestoneParams{IncidentID: declared.ID, Kind: kind, Timestamp: declared.OpenedAt.Add(time.Duration(minutes) * time.Minute), Source: "web"}
		_, addErr := svc.AddMilestone(ctx, params)
		s.Require().NoError(addErr)
	}

	add(im.KindMitigation, 10)
	s.Equal(incident.ResponseStateMitigated, tdb.Client(ctx).Incident.GetX(ctx, declared.ID).ResponseState)
	add(im.KindImpact, 20)
	s.Equal(incident.ResponseStateStarted, tdb.Client(ctx).Incident.GetX(ctx, declared.ID).ResponseState)
}

func (s *IncidentManagementServiceSuite) TestResolvedIsFinal() {
	ctx, tdb := s.SetupTestDatabase()
	s.enable(ctx, tdb)
	svc := s.newService(tdb)
	declared := s.declare(ctx, tdb, svc)
	openedAt := declared.OpenedAt

	add := func(kind im.Kind, at time.Time) error {
		_, addErr := svc.AddMilestone(ctx, rez.AddIncidentMilestoneParams{IncidentID: declared.ID, Kind: kind, Timestamp: at, Source: "web"})
		return addErr
	}
	s.Require().NoError(add(im.KindImpact, openedAt))
	s.Require().NoError(add(im.KindMitigation, openedAt.Add(10*time.Minute)))
	s.Require().NoError(add(im.KindResolution, openedAt.Add(20*time.Minute)))
	s.Require().NoError(add(im.KindImpact, openedAt.Add(20*time.Minute)), "an impact at the resolution's time is allowed")

	secondResolutionErr := add(im.KindResolution, openedAt.Add(30*time.Minute))
	s.Equal("This incident is resolved. Its response can't change.", errs.PublicMessage(secondResolutionErr))
	laterImpactErr := add(im.KindImpact, openedAt.Add(30*time.Minute))
	s.Equal("The resolution can't be before an impact or mitigation.", errs.PublicMessage(laterImpactErr))

	resolved, getErr := svc.details.GetIncident(ctx, incident.ID(declared.ID))
	s.Require().NoError(getErr)
	s.Len(resolved.Edges.Milestones, 4)
	s.Equal(incident.ResponseStateResolved, resolved.ResponseState)
	s.Require().NotNil(resolved.ResolvedAt)
	s.True(resolved.ResolvedAt.Equal(openedAt.Add(20*time.Minute)), "the resolution's time")
	s.NotNil(resolved.Edges.Retrospective)
}

func (s *IncidentManagementServiceSuite) TestRoleLimit() {
	ctx, tdb := s.SetupTestDatabase()
	s.enable(ctx, tdb)
	svc := s.newService(tdb)
	client := tdb.Client(ctx)
	declared := s.declare(ctx, tdb, svc)

	createCommander := client.IncidentRole.Create().
		SetName("Commander").
		SetMaxAssignees(1)
	commander, commanderErr := createCommander.Save(ctx)
	s.Require().NoError(commanderErr)
	responder, responderErr := client.IncidentRole.Create().SetName("Responder").Save(ctx)
	s.Require().NoError(responderErr)
	first, firstErr := client.User.Query().FirstID(ctx)
	s.Require().NoError(firstErr)
	createSecond := client.User.Create().
		SetEmail("second+" + uuid.NewString() + "@example.com").
		SetName("Second Responder")
	second, secondErr := createSecond.Save(ctx)
	s.Require().NoError(secondErr)

	assigned, assignErr := svc.SetRoleAssignment(ctx, declared.ID, commander.ID, first)
	s.Require().NoError(assignErr)
	again, againErr := svc.SetRoleAssignment(ctx, declared.ID, commander.ID, first)
	s.Require().NoError(againErr)
	s.Equal(assigned.ID, again.ID, "assigning a held role changes nothing")
	_, responderAssignErr := svc.SetRoleAssignment(ctx, declared.ID, responder.ID, first)
	s.Require().NoError(responderAssignErr, "a person can hold several roles")

	_, limitErr := svc.SetRoleAssignment(ctx, declared.ID, commander.ID, second.ID)
	s.Equal(errs.CodeConflict, errs.CodeOf(limitErr))
	s.Equal("Commander can be held by 1 person in an incident.", errs.PublicMessage(limitErr))
}
