package eventprojection

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	entuser "github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) createUserProjectionEvent(tdb rez.Database, resourceRef string, attrs projections.UserEventAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	r := s.Require()
	encoded, attrsErr := projections.EncodeAttributes(attrs)
	r.NoError(attrsErr)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("projection-tests").
		SetProviderResourceRef(resourceRef).
		SetProviderEventSource("users").
		SetProviderEventRef("user-event-" + uuid.NewString()).
		SetKind(projections.KindUser).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded)

	ev, eventErr := createEvent.Save(ctx)
	r.NoError(eventErr)
	return ev
}

func (s *ProjectionServiceSuite) TestUserProjectionCreatesAndLinksKnowledgeEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	projector := s.projectionService(tdb)

	attrs := projections.UserEventAttributes{
		Name:     "Projected User",
		Email:    "projected+" + uuid.NewString() + "@example.com",
		ChatId:   "U123",
		Timezone: "Australia/Sydney",
	}
	ev := s.createUserProjectionEvent(tdb, "user-1", attrs)

	r := s.Require()

	_, projErr := runProjection(ctx, projector, ev)
	r.NoError(projErr)

	findUser := tdb.Client(ctx).User.Query().
		Where(entuser.Email(attrs.Email))
	created, err := findUser.Only(ctx)
	r.NoError(err)
	r.NotNil(created.KnowledgeEntityID)
	r.Equal("Projected User", created.Name)
	r.Equal("U123", created.ChatID)
}

func (s *ProjectionServiceSuite) TestUserProjectionReusesExistingEmailUser() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	r := s.Require()

	userId := uuid.New()
	email := "existing+" + uuid.NewString() + "@example.com"
	createUser := tdb.Client(ctx).User.Create().
		SetID(userId).
		SetEmail(email).
		SetName("Existing")
	r.NoError(createUser.Exec(ctx))

	projector := s.projectionService(tdb)

	attrs := projections.UserEventAttributes{
		Name:  "Existing Updated",
		Email: email,
	}
	ev := s.createUserProjectionEvent(tdb, "user-2", attrs)

	_, projErr := runProjection(ctx, projector, ev)
	r.NoError(projErr)

	queryUser := tdb.Client(ctx).User.Query().
		Where(entuser.Email(email))
	emailUser, lookupUserErr := queryUser.Only(ctx)
	r.NoError(lookupUserErr)
	r.Equal(userId, emailUser.ID)
	r.NotNil(emailUser.KnowledgeEntityID)
	r.Equal("Existing Updated", emailUser.Name)
}

func (s *ProjectionServiceSuite) TestUserProjectionFailsWhenKnowledgeLinkConflictsWithEmailOwner() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	r := s.Require()

	userRef := "user-3"
	firstAttrs := projections.UserEventAttributes{
		Name:  "Linked User",
		Email: "linked+" + uuid.NewString() + "@example.com",
	}
	first := s.createUserProjectionEvent(tdb, userRef, firstAttrs)

	projector := s.projectionService(tdb)

	_, projErr := runProjection(ctx, projector, first)
	r.NoError(projErr)

	conflictEmail := "conflict+" + uuid.NewString() + "@example.com"
	createUser := tdb.Client(ctx).User.Create().
		SetEmail(conflictEmail).
		SetName("Email Owner")
	r.NoError(createUser.Exec(ctx))

	secondAttrs := projections.UserEventAttributes{
		Name:  "Linked User",
		Email: conflictEmail,
	}
	second := s.createUserProjectionEvent(tdb, userRef, secondAttrs)
	evidenceBefore, queryEvidenceErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	r.NoError(queryEvidenceErr)

	_, projConfErr := runProjection(ctx, projector, second)
	r.Error(projConfErr)

	evidenceAfter, queryEvidenceAfterErr := tdb.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	r.NoError(queryEvidenceAfterErr)
	r.Equal(evidenceBefore, evidenceAfter)
}
