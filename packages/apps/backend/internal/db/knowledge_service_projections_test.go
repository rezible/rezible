package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knfh "github.com/rezible/rezible/ent/knowledgefacthistory"
	knfp "github.com/rezible/rezible/ent/knowledgefactprovenance"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/projections"
	"github.com/rezible/rezible/testkit"
)

type KnowledgeServiceEventProjectionSuite struct {
	testkit.Suite
}

func TestKnowledgeServiceEventProjectionSuite(t *testing.T) {
	suite.Run(t, &KnowledgeServiceEventProjectionSuite{Suite: testkit.NewSuite()})
}

func (s *KnowledgeServiceEventProjectionSuite) service() *KnowledgeService {
	return NewKnowledgeService(s.Client())
}

func (s *KnowledgeServiceEventProjectionSuite) createNormalizedEvent(fn func(*ent.NormalizedEventCreate)) *ent.NormalizedEvent {
	now := time.Now().UTC()
	create := s.Client().NormalizedEvent.Create().
		SetOccurredAt(now).
		SetReceivedAt(now).
		SetProcessingVersion("test.v1")
	fn(create)
	ev, saveErr := create.Save(s.SeedTenantContext())
	s.Require().NoError(saveErr)
	return ev
}

func (s *KnowledgeServiceEventProjectionSuite) TestRepositoryObservedCreatesAliasProvenanceAndHistory() {
	ctx := s.SeedTenantContext()
	dbc := s.Client()

	repoRef := "rezible/repository-" + uuid.NewString()
	attrs := projections.RepositoryObservedAttributes{
		DisplayName: repoRef,
		URL:         "https://github.com/rezible/rezible",
	}
	evKind := ne.KindRepositoryObserved
	provider := "github"
	providerSource := "code_checkout"
	ev := s.createNormalizedEvent(func(c *ent.NormalizedEventCreate) {
		c.SetProvider(provider).
			SetProviderSource(providerSource).
			SetKind(evKind).
			SetSubjectKind(string(evKind)).
			SetSubjectRef(repoRef).
			SetProviderEventRef(evKind.String() + ":" + repoRef + ":" + uuid.NewString()).
			SetAttributes(attrs.Encode())
	})

	validated, validationErr := projections.ValidateEvent(ev)
	s.Require().NoError(validationErr)

	projectedEvent, castOk := validated.(projections.RepositoryObserved)
	s.Require().True(castOk)

	ps := newEventProjectionProcessor(ev)
	result, projectionErr := ps.projectRepositoryObserved(ctx, projectedEvent)
	s.Require().NoError(projectionErr)
	s.Require().NoError(ps.saveProjectionResult(ctx, dbc, result), "saving projection result")

	queryAlias := dbc.KnowledgeEntityAlias.Query().
		Where(knea.Provider(provider), knea.ProviderSource(providerSource)).
		Where(knea.SubjectKind("repository"), knea.SubjectRef(repoRef))
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)
	s.Equal(ev.ID, *alias.NormalizedEventID)

	entity, entityErr := dbc.KnowledgeEntity.Get(ctx, alias.EntityID)
	s.Require().NoError(entityErr)
	s.Equal(kne.KindRepository, entity.Kind)
	s.Equal(repoRef, entity.DisplayName)

	queryFactProvenance := dbc.KnowledgeFactProvenance.Query().
		Where(knfp.AliasID(alias.ID))
	provCount, provCountErr := queryFactProvenance.Count(ctx)
	s.Require().NoError(provCountErr)
	s.Equal(1, provCount)

	queryFactHistory := dbc.KnowledgeFactHistory.Query().
		Where(knfh.AliasID(alias.ID))
	historyCount, historyCountErr := queryFactHistory.Count(ctx)
	s.Require().NoError(historyCountErr)
	s.Equal(1, historyCount)
}

func (s *KnowledgeServiceEventProjectionSuite) TestCodeChangeEventLinksToRepositoryIdempotently() {
	ctx := s.SeedTenantContext()
	dbc := s.Client()

	subjectRef := "github:pull_request:" + uuid.NewString()
	attrs := projections.ChangeEventObservedAttributes{
		RepositoryExternalRef: "rezible/change-repository-" + uuid.NewString(),
		DisplayName:           "Fix payments timeout",
	}
	evKind := ne.KindChangeEventObserved
	ev := s.createNormalizedEvent(func(c *ent.NormalizedEventCreate) {
		c.SetProvider("github").
			SetProviderSource("test").
			SetKind(evKind).
			SetSubjectKind(string(evKind)).
			SetSubjectRef(subjectRef).
			SetProviderEventRef(evKind.String() + ":" + subjectRef + ":" + uuid.NewString()).
			SetAttributes(attrs.Encode())
	})

	validated, validationErr := projections.ValidateEvent(ev)
	s.Require().NoError(validationErr)
	projectedEvent, castOk := validated.(projections.ChangeEventObserved)
	s.Require().True(castOk)

	numRuns := 2
	for i := 1; i <= numRuns; i++ {
		ps := newEventProjectionProcessor(ev)
		result, projErr := ps.projectCodeChangeEventObserved(ctx, projectedEvent)
		s.Require().NoError(projErr, "creating projection #%d", i)
		s.Require().NoError(ps.saveProjectionResult(ctx, dbc, result), "failed to save projection #%d", i)
	}

	queryChangeAlias := dbc.KnowledgeEntityAlias.Query().
		Where(knea.SubjectKind("change_event"), knea.SubjectRef(ev.SubjectRef))
	changeAlias, changeAliasErr := queryChangeAlias.Only(ctx)
	s.Require().NoError(changeAliasErr)

	changeEntity, changeEntityErr := dbc.KnowledgeEntity.Get(ctx, changeAlias.EntityID)
	s.Require().NoError(changeEntityErr)
	s.Equal(kne.KindChangeEvent, changeEntity.Kind)

	queryRepoAlias := dbc.KnowledgeEntityAlias.Query().
		Where(knea.SubjectKind("repository"), knea.SubjectRef(attrs.RepositoryExternalRef))
	repoAlias, repoAliasErr := queryRepoAlias.Only(ctx)
	s.Require().NoError(repoAliasErr)

	repoEntity, repoEntityErr := dbc.KnowledgeEntity.Get(ctx, repoAlias.EntityID)
	s.Require().NoError(repoEntityErr)
	s.Equal(kne.KindRepository, repoEntity.Kind)

	queryEntityAliases := dbc.KnowledgeEntityAlias.Query().
		Where(knea.IDIn(changeAlias.ID, repoAlias.ID))
	aliasCount, aliasCountErr := queryEntityAliases.Count(ctx)
	s.Require().NoError(aliasCountErr)
	s.Equal(2, aliasCount, "expected only 2 aliases to be created")

	queryRelationships := dbc.KnowledgeRelationship.Query().
		Where(knr.Kind("changes_repository")).
		Where(knr.SourceEntityID(changeEntity.ID), knr.TargetEntityID(repoEntity.ID))
	relCount, relCountErr := queryRelationships.Count(ctx)
	s.Require().NoError(relCountErr)
	s.Equal(1, relCount, "expected only one relationship to be created")

	queryProvs := dbc.KnowledgeFactProvenance.Query().
		Where(knfp.ProviderEventRef(ev.ProviderEventRef))
	provCount, provCountErr := queryProvs.Count(ctx)
	s.Require().NoError(provCountErr)
	s.Equal(3, provCount, "expected 3 fact provenance records")

	queryHistory := dbc.KnowledgeFactHistory.Query().
		Where(knfh.ProviderEventRef(ev.ProviderEventRef))
	historyCount, historyCountErr := queryHistory.Count(ctx)
	s.Require().NoError(historyCountErr)
	s.Equal(3, historyCount, "expected 3 knowledge facts")
}

func (s *KnowledgeServiceEventProjectionSuite) TestCodeChangeEventDoesNotOverwriteExistingRepositoryDetails() {
	ctx := s.SeedTenantContext()
	dbc := s.Client()

	provider := "github"
	providerSource := "test"
	repoRef := "rezible/rich-repository-" + uuid.NewString()
	repoDisplayName := "Rich Repository"
	repoURL := "https://github.com/rezible/rich-repository"

	repoAttrs := projections.RepositoryObservedAttributes{
		DisplayName: repoDisplayName,
		URL:         repoURL,
	}
	repoEventKind := ne.KindRepositoryObserved
	repoEvent := s.createNormalizedEvent(func(c *ent.NormalizedEventCreate) {
		c.SetProvider(provider).
			SetProviderSource(providerSource).
			SetKind(repoEventKind).
			SetSubjectKind(string(repoEventKind)).
			SetSubjectRef(repoRef).
			SetProviderEventRef(repoEventKind.String() + ":" + repoRef + ":" + uuid.NewString()).
			SetAttributes(repoAttrs.Encode())
	})

	validatedRepoEvent, repoValidationErr := projections.ValidateEvent(repoEvent)
	s.Require().NoError(repoValidationErr)
	projectedRepoEvent, repoCastOk := validatedRepoEvent.(projections.RepositoryObserved)
	s.Require().True(repoCastOk)

	repoProjection := newEventProjectionProcessor(repoEvent)
	repoResult, repoProjectionErr := repoProjection.projectRepositoryObserved(ctx, projectedRepoEvent)
	s.Require().NoError(repoProjectionErr)
	s.Require().NoError(repoProjection.saveProjectionResult(ctx, dbc, repoResult), "saving repository projection result")

	changeAttrs := projections.ChangeEventObservedAttributes{
		RepositoryExternalRef: repoRef,
		DisplayName:           "Fix repository sync",
	}
	changeEventKind := ne.KindChangeEventObserved
	changeSubjectRef := "github:pull_request:" + uuid.NewString()
	changeEvent := s.createNormalizedEvent(func(c *ent.NormalizedEventCreate) {
		c.SetProvider(provider).
			SetProviderSource(providerSource).
			SetKind(changeEventKind).
			SetSubjectKind(string(changeEventKind)).
			SetSubjectRef(changeSubjectRef).
			SetProviderEventRef(changeEventKind.String() + ":" + changeSubjectRef + ":" + uuid.NewString()).
			SetAttributes(changeAttrs.Encode())
	})

	validatedChangeEvent, changeValidationErr := projections.ValidateEvent(changeEvent)
	s.Require().NoError(changeValidationErr)
	projectedChangeEvent, changeCastOk := validatedChangeEvent.(projections.ChangeEventObserved)
	s.Require().True(changeCastOk)

	changeProjection := newEventProjectionProcessor(changeEvent)
	changeResult, changeProjectionErr := changeProjection.projectCodeChangeEventObserved(ctx, projectedChangeEvent)
	s.Require().NoError(changeProjectionErr)
	s.Require().NoError(changeProjection.saveProjectionResult(ctx, dbc, changeResult), "saving change event projection result")

	repoAlias, repoAliasErr := dbc.KnowledgeEntityAlias.Query().
		Where(knea.Provider(provider), knea.ProviderSource(providerSource)).
		Where(knea.SubjectKind("repository"), knea.SubjectRef(repoRef)).
		Only(ctx)
	s.Require().NoError(repoAliasErr)

	repoEntity, repoEntityErr := dbc.KnowledgeEntity.Get(ctx, repoAlias.EntityID)
	s.Require().NoError(repoEntityErr)
	s.Equal(kne.KindRepository, repoEntity.Kind)
	s.Equal(repoDisplayName, repoEntity.DisplayName)
	s.Equal(repoURL, repoEntity.Properties["url"])
	s.Equal(repoRef, repoEntity.Properties["external_ref"])
}
