package eventprojection

import (
	"context"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

var linkingTestTime = time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

// alertmanagerServiceObservation is the service observation an Alertmanager installation makes for a label value.
func alertmanagerServiceObservation(installation string, label string) projections.EntityObservation {
	name := projections.NormalizeServiceName(label)
	return projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "alertmanager",
			ProviderNamespace: installation,
			ResourceRef:       "service:" + name,
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: label,
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: name,
		},
	}
}

func (s *ProjectionServiceSuite) projectServiceAlert(ctx context.Context, tdb rez.Database, service *ProjectionService, installation string, label string) {
	attrs := projections.AlertInstanceEventAttributes{
		Title:            "HighErrorRate",
		State:            projections.AlertStateFiring,
		Severity:         "critical",
		StartedAt:        linkingTestTime,
		ObservedEntities: []projections.EntityObservation{alertmanagerServiceObservation(installation, label)},
	}
	event := s.createAlertProjectionEvent(ctx, tdb, installation+":HighErrorRate", linkingTestTime, attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)
}

// githubRepositoryObservation is the repository observation GitHub makes for a repository ID and full name.
func githubRepositoryObservation(repositoryID string, fullName string, linkingValue string) projections.EntityObservation {
	return projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "github",
			ProviderNamespace: "installation-1",
			ResourceRef:       repositoryID,
		},
		Category:    kne.CategoryCode,
		Kind:        knowledgeEntityKindRepository,
		DisplayName: fullName,
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeRepositoryFullName: linkingValue,
		},
	}
}

func (s *ProjectionServiceSuite) projectGithubChange(ctx context.Context, tdb rez.Database, service *ProjectionService, changeRef string, repository projections.EntityObservation) error {
	attrs := projections.CodeChangeEventAttributes{
		Repository:  repository,
		DisplayName: "refs/heads/main",
	}
	event := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, changeRef, linkingTestTime, attrs)

	_, projectErr := runProjection(ctx, service, event)
	return projectErr
}

// projectFixtureRepository observes a repository from the test provider, as a provider other than GitHub would.
func (s *ProjectionServiceSuite) projectFixtureRepository(ctx context.Context, tdb rez.Database, service *ProjectionService, fullName string) {
	attrs := projections.CodeForgeEventAttributes{
		DisplayName: fullName,
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeRepositoryFullName: fullName,
		},
	}
	event := s.createNormalizedEvent(ctx, tdb, projections.KindCodeForge, "repository:"+fullName, linkingTestTime, attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)
}

func (s *ProjectionServiceSuite) entitiesOfKind(ctx context.Context, tdb rez.Database, category kne.Category, kind string) []*ent.KnowledgeEntity {
	queryEntities := tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.CategoryEQ(category), kne.Kind(kind))
	entities, queryErr := queryEntities.All(ctx)
	s.Require().NoError(queryErr)
	return entities
}

func (s *ProjectionServiceSuite) relationshipsOf(ctx context.Context, tdb rez.Database, predicate knr.Predicate) []*ent.KnowledgeRelationship {
	queryRelationships := tdb.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(predicate))
	relationships, queryErr := queryRelationships.All(ctx)
	s.Require().NoError(queryErr)
	return relationships
}

func (s *ProjectionServiceSuite) TestAlertsFromTwoInstallationsObserveOneService() {
	s.Run("Checkout_API first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.projectServiceAlert(ctx, tdb, service, "installation-a", "Checkout_API")
		s.projectServiceAlert(ctx, tdb, service, "installation-b", "checkout-api")

		s.assertAlertsObserveOneService(ctx, tdb)
	})

	s.Run("checkout-api first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.projectServiceAlert(ctx, tdb, service, "installation-b", "checkout-api")
		s.projectServiceAlert(ctx, tdb, service, "installation-a", "Checkout_API")

		s.assertAlertsObserveOneService(ctx, tdb)
	})
}

func (s *ProjectionServiceSuite) assertAlertsObserveOneService(ctx context.Context, tdb rez.Database) {
	services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
	s.Require().Len(services, 1)
	observes := s.relationshipsOf(ctx, tdb, knr.PredicateObserves)
	s.Require().Len(observes, 2, "each installation's alert observes the service")
	for _, relationship := range observes {
		s.Equal(services[0].ID, relationship.TargetEntityID)
	}
}

// topologyObservation is an entity observed by a topology fixture provider, linked when a service name is given.
func topologyObservation(resourceRef string, kind string, serviceName string) projections.EntityObservation {
	observation := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          "test",
			ProviderNamespace: "topology",
			ResourceRef:       resourceRef,
		},
		Category:    kne.CategoryContainer,
		Kind:        kind,
		DisplayName: resourceRef,
	}
	if serviceName != "" {
		observation.LinkingAttributes = projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: serviceName,
		}
	}
	return observation
}

func (s *ProjectionServiceSuite) projectUsesRelationship(ctx context.Context, tdb rez.Database, service *ProjectionService, source projections.EntityObservation, target projections.EntityObservation) {
	relationship := projections.SystemRelationshipEventAttributes{
		Predicate: knr.PredicateUses,
		Source:    source,
		Target:    target,
	}
	relationshipRef := source.Ref.ResourceRef + "-uses-" + target.Ref.ResourceRef
	event := s.createNormalizedEvent(ctx, tdb, projections.KindSystemRelationship, relationshipRef, linkingTestTime, relationship)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)
}

func (s *ProjectionServiceSuite) TestSystemRelationshipEndpointsLandOnLinkedService() {
	s.Run("source", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)
		s.projectServiceAlert(ctx, tdb, service, "installation-a", "checkout-api")

		checkout := topologyObservation("checkout", "service", "checkout-api")
		database := topologyObservation("checkout-db", "database", "")
		s.projectUsesRelationship(ctx, tdb, service, checkout, database)

		services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
		s.Require().Len(services, 1)
		uses := s.relationshipsOf(ctx, tdb, knr.PredicateUses)
		s.Require().Len(uses, 1)
		s.Equal(services[0].ID, uses[0].SourceEntityID)
	})

	s.Run("target", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)
		s.projectServiceAlert(ctx, tdb, service, "installation-a", "checkout-api")

		storefront := topologyObservation("storefront", "web_app", "")
		checkout := topologyObservation("checkout", "service", "checkout-api")
		s.projectUsesRelationship(ctx, tdb, service, storefront, checkout)

		services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
		s.Require().Len(services, 1)
		uses := s.relationshipsOf(ctx, tdb, knr.PredicateUses)
		s.Require().Len(uses, 1)
		s.Equal(services[0].ID, uses[0].TargetEntityID)
	})
}

func (s *ProjectionServiceSuite) TestCodeChangeImpactLandsOnLinkedService() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	s.projectServiceAlert(ctx, tdb, service, "installation-a", "checkout-api")

	change := projections.CodeChangeEventAttributes{
		Repository:       githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout"),
		DisplayName:      "refs/heads/main",
		ImpactedEntities: []projections.EntityObservation{topologyObservation("checkout", "service", "checkout-api")},
	}
	event := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, "change:42:abc123", linkingTestTime, change)
	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
	s.Require().Len(services, 1)
	impacts := s.relationshipsOf(ctx, tdb, knr.PredicateImpacts)
	s.Require().Len(impacts, 1)
	s.Equal(services[0].ID, impacts[0].TargetEntityID)
}

func (s *ProjectionServiceSuite) TestRepositoryObservationsFromTwoProvidersResolveToOneRepository() {
	githubCheckout := githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout")

	s.Run("fixture first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.projectFixtureRepository(ctx, tdb, service, "acme/checkout")
		s.Require().NoError(s.projectGithubChange(ctx, tdb, service, "change:42:abc123", githubCheckout))

		s.assertChangeTouchesOneRepository(ctx, tdb)
	})

	s.Run("GitHub first", func() {
		ctx, tdb := s.SetupTestDatabase()
		service := s.projectionService(tdb)

		s.Require().NoError(s.projectGithubChange(ctx, tdb, service, "change:42:abc123", githubCheckout))
		s.projectFixtureRepository(ctx, tdb, service, "acme/checkout")

		s.assertChangeTouchesOneRepository(ctx, tdb)
	})
}

func (s *ProjectionServiceSuite) assertChangeTouchesOneRepository(ctx context.Context, tdb rez.Database) {
	repositories := s.entitiesOfKind(ctx, tdb, kne.CategoryCode, knowledgeEntityKindRepository)
	s.Require().Len(repositories, 1)
	touches := s.relationshipsOf(ctx, tdb, knr.PredicateTouches)
	s.Require().Len(touches, 1)
	s.Equal(repositories[0].ID, touches[0].TargetEntityID)
}

func (s *ProjectionServiceSuite) TestAliasBoundToAnotherEntityThanItsNameConflicts() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	checkout := githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout")
	s.Require().NoError(s.projectGithubChange(ctx, tdb, service, "change:42:abc123", checkout))
	s.projectFixtureRepository(ctx, tdb, service, "acme/payments")

	// GitHub repository 42 is renamed to a name another provider already observed.
	renamed := githubRepositoryObservation("42", "Acme/Payments", "acme/payments")
	projectErr := s.projectGithubChange(ctx, tdb, service, "change:42:def456", renamed)

	s.ErrorIs(projectErr, rez.ErrConflict)
}
