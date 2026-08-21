package eventprojection

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	entuser "github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) createUserProjectionEvent(tdb rez.Database, subjectRef string, attrs projections.UserSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	ev, err := tdb.Client(ctx).NormalizedEvent.Create().
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

func (s *ProjectionServiceSuite) TestUserProjectionCreatesAndLinksKnowledgeEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	projector := s.projectionService(tdb)

	email := "projected+" + uuid.NewString() + "@example.com"
	attrs := projections.UserSubjectAttributes{
		Name:     "Projected User",
		Email:    email,
		ChatId:   "U123",
		Timezone: "Australia/Perth",
	}
	ev := s.createUserProjectionEvent(tdb, "user-1", attrs)

	_, projErr := runProjection(ctx, projector, ev)
	s.Require().NoError(projErr)

	created, err := tdb.Client(ctx).User.Query().
		Where(entuser.Email(email)).
		Only(ctx)
	s.Require().NoError(err)
	s.NotNil(created.KnowledgeEntityID)
	s.Equal("Projected User", created.Name)
	s.Equal("U123", created.ChatID)
}

func (s *ProjectionServiceSuite) TestUserProjectionReusesExistingEmailUser() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	projector := s.projectionService(tdb)

	email := "existing+" + uuid.NewString() + "@example.com"
	existing, err := tdb.Client(ctx).User.Create().
		SetEmail(email).
		SetName("Existing").
		Save(ctx)
	s.Require().NoError(err)
	ev := s.createUserProjectionEvent(tdb, "user-2", projections.UserSubjectAttributes{
		Name:  "Existing Updated",
		Email: email,
	})

	_, projErr := runProjection(ctx, projector, ev)
	s.Require().NoError(projErr)

	users, err := tdb.Client(ctx).User.Query().
		Where(entuser.Email(email)).
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(users, 1)
	s.Equal(existing.ID, users[0].ID)
	s.NotNil(users[0].KnowledgeEntityID)
	s.Equal("Existing Updated", users[0].Name)
}

func (s *ProjectionServiceSuite) TestUserProjectionFailsWhenKnowledgeLinkConflictsWithEmailOwner() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	projector := s.projectionService(tdb)

	firstEmail := "linked+" + uuid.NewString() + "@example.com"
	first := s.createUserProjectionEvent(tdb, "user-3", projections.UserSubjectAttributes{
		Name:  "Linked User",
		Email: firstEmail,
	})

	_, projErr := runProjection(ctx, projector, first)
	s.Require().NoError(projErr)

	conflictEmail := "conflict+" + uuid.NewString() + "@example.com"
	createUser := tdb.Client(ctx).User.Create().
		SetEmail(conflictEmail).
		SetName("Email Owner")
	s.Require().NoError(createUser.Exec(ctx))

	conflict := s.createUserProjectionEvent(tdb, "user-3", projections.UserSubjectAttributes{
		Name:  "Linked User",
		Email: conflictEmail,
	})
	evidenceBefore, queryEvidenceErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(queryEvidenceErr)

	_, projConfErr := runProjection(ctx, projector, conflict)
	s.Require().Error(projConfErr)

	evidenceAfter, queryEvidenceAfterErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(queryEvidenceAfterErr)
	s.Equal(evidenceBefore, evidenceAfter)
}
