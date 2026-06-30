package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	entuser "github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

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
		SetKind(ne.KindObserved).
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
	svc := s.newUserService(nil)
	ev := s.createUserProjectionEvent("user-1", projections.UserSubjectAttributes{
		Name:     "Projected User",
		Email:    "projected+" + uuid.NewString() + "@example.com",
		ChatId:   "U123",
		Timezone: "Australia/Perth",
	})

	_, projErr := svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(projErr)

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
	svc := s.newUserService(nil)
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

	_, projErr := svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(projErr)

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
	svc := s.newUserService(nil)
	firstEmail := "linked+" + uuid.NewString() + "@example.com"
	first := s.createUserProjectionEvent("user-3", projections.UserSubjectAttributes{
		Name:  "Linked User",
		Email: firstEmail,
	})

	_, projErr := svc.HandleEventProjection(ctx, first)
	s.Require().NoError(projErr)

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

	_, projConfErr := svc.HandleEventProjection(ctx, conflict)
	s.Require().Error(projConfErr)
}
