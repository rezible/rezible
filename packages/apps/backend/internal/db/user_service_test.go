package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	entuser "github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/projections"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/testkit"
)

type UserServiceSuite struct {
	testkit.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: testkit.NewSuite()})
}

func (s *UserServiceSuite) newUserService() *UserService {
	svc, err := NewUserService(s.Database(), nil, NewKnowledgeService(s.Database()))
	s.Require().NoError(err)
	return svc
}

func (s *UserServiceSuite) createUserProjectionEvent(subjectRef string, attrs projections.UserSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("users").
		SetProviderEventRef("user-event-" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindUser.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *UserServiceSuite) TestUserProjectionCreatesAndLinksKnowledgeEntity() {
	ctx := s.SeedTenantContext()
	svc := s.newUserService()
	ev := s.createUserProjectionEvent("user-1", projections.UserSubjectAttributes{
		Name:     "Projected User",
		Email:    "projected+" + uuid.NewString() + "@example.com",
		ChatId:   "U123",
		Timezone: "Australia/Perth",
	})

	s.Require().NoError(svc.HandleEventProjection(ctx, ev))

	created, err := s.Client(ctx).User.Query().
		Where(entuser.Email(ev.Attributes["email"].(string))).
		Only(ctx)
	s.Require().NoError(err)
	s.NotNil(created.KnowledgeEntityID)
	s.Equal("Projected User", created.Name)
	s.Equal("U123", created.ChatID)
}

func (s *UserServiceSuite) TestUserProjectionReusesExistingEmailUser() {
	ctx := s.SeedTenantContext()
	svc := s.newUserService()
	email := "existing+" + uuid.NewString() + "@example.com"
	existing, err := s.Client(ctx).User.Create().
		SetEmail(email).
		SetName("Existing").
		Save(ctx)
	s.Require().NoError(err)
	ev := s.createUserProjectionEvent("user-2", projections.UserSubjectAttributes{
		Name:  "Existing Updated",
		Email: email,
	})

	s.Require().NoError(svc.HandleEventProjection(ctx, ev))

	users, err := s.Client(ctx).User.Query().
		Where(entuser.Email(email)).
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(users, 1)
	s.Equal(existing.ID, users[0].ID)
	s.NotNil(users[0].KnowledgeEntityID)
	s.Equal("Existing Updated", users[0].Name)
}

func (s *UserServiceSuite) TestUserProjectionFailsWhenKnowledgeLinkConflictsWithEmailOwner() {
	ctx := s.SeedTenantContext()
	svc := s.newUserService()
	firstEmail := "linked+" + uuid.NewString() + "@example.com"
	first := s.createUserProjectionEvent("user-3", projections.UserSubjectAttributes{
		Name:  "Linked User",
		Email: firstEmail,
	})
	s.Require().NoError(svc.HandleEventProjection(ctx, first))

	conflictEmail := "conflict+" + uuid.NewString() + "@example.com"
	_, err := s.Client(ctx).User.Create().
		SetEmail(conflictEmail).
		SetName("Email Owner").
		Save(ctx)
	s.Require().NoError(err)
	conflict := s.createUserProjectionEvent("user-3", projections.UserSubjectAttributes{
		Name:  "Linked User",
		Email: conflictEmail,
	})

	s.Require().Error(svc.HandleEventProjection(ctx, conflict))
}
