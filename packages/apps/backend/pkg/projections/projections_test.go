package projections

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/test"
)

type ProjectionsSuite struct {
	test.Suite
}

func TestProjectionsSuite(t *testing.T) {
	suite.Run(t, &ProjectionsSuite{Suite: test.NewSuite()})
}

type ExampleEventAttributes struct {
	FooBar string `json:"foo_bar" validate:"required"`
}

func (s *ProjectionsSuite) TestDecodeIncidentObservedEvent() {
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	attrs := IncidentEventAttributes{
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	}
	encAttrs, encErr := EncodeAttributes(attrs)
	s.Require().NoError(encErr)
	ev := &ent.NormalizedEvent{Attributes: encAttrs}
	incEv, decodeErr := DecodeEventAttributes[IncidentEventAttributes](ev)
	s.Require().NoError(decodeErr)
	s.Equal("Checkout search lookups timing out", incEv.Attributes.Title)
	s.Equal(attrs.SeverityRef, incEv.Attributes.SeverityRef)
	s.Equal(attrs.TypeRef, incEv.Attributes.TypeRef)
	s.True(openedAt.Equal(incEv.Attributes.OpenedAt))
}

func (s *ProjectionsSuite) TestDecodeWithRejectsMissingRequiredAttributes() {
	ev := &ent.NormalizedEvent{
		Attributes: []byte("{}"),
	}
	_, decodeErr := DecodeEventAttributes[ExampleEventAttributes](ev)
	s.Require().Error(decodeErr)
	s.ErrorContains(decodeErr, "failed on the 'required' tag")
}

func (s *ProjectionsSuite) TestSortEntityObservations() {
	searchRef := rez.ProviderResourceRef{
		Provider:          "demo",
		ProviderNamespace: "demo",
		ResourceRef:       "demo:component:search_api",
	}
	elasticsearchRef := rez.ProviderResourceRef{
		Provider:          "demo",
		ProviderNamespace: "demo",
		ResourceRef:       "demo:component:elasticsearch_catalog",
	}
	refs := []EntityObservation{
		{
			Ref:         searchRef,
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: "Search API",
		},
		{
			Ref:         elasticsearchRef,
			Category:    kne.CategoryContainer,
			Kind:        "search_cluster",
			DisplayName: "Elasticsearch Catalog",
		},
	}

	sortedRefs := SortEntityObservations(refs)

	s.Equal("demo:component:elasticsearch_catalog", sortedRefs[0].Ref.ResourceRef)
	s.Equal("demo:component:search_api", sortedRefs[1].Ref.ResourceRef)
	s.Equal("demo:component:search_api", refs[0].Ref.ResourceRef)
}

func (s *ProjectionsSuite) TestUserEntityLinkingAttributesNormalizeEmail() {
	attrs := UserEntityLinkingAttributes{Email: "  Alice@Example.COM "}
	s.Equal(map[string]string{"user.email": "alice@example.com"}, attrs.Values())
	s.Nil((UserEntityLinkingAttributes{Email: "   "}).Values())
}

func (s *ProjectionsSuite) TestNormalizeServiceName() {
	cases := map[string]string{
		"checkout-api":    "checkout-api",
		"Checkout_API":    "checkout-api",
		"  Search..API  ": "search-api",
		"cart":            "cart",
		"payments/v2":     "payments-v2",
		"---":             "",
		"":                "",
		"Überprüfung-1":   "berpr-fung-1",
	}
	for value, expected := range cases {
		s.Equal(expected, NormalizeServiceName(value), "value %q", value)
	}
}

func (s *ProjectionsSuite) TestDerivedRelationshipRefUsesCompleteDirectionalIdentity() {
	source := rez.ProviderResourceRef{Provider: "github", ProviderNamespace: "org-1", ResourceRef: "change-1"}
	target := rez.ProviderResourceRef{Provider: "github", ProviderNamespace: "org-1", ResourceRef: "repo-1"}

	base := DerivedRelationshipRef(knr.PredicateTouches, source, target)
	s.Equal("rezible", base.Provider)
	s.Empty(base.ProviderNamespace)
	s.Contains(base.ResourceRef, "relationship:v1:")
	s.Equal(base, DerivedRelationshipRef(knr.PredicateTouches, source, target))

	differentNamespace := target
	differentNamespace.ProviderNamespace = "org-2"
	s.NotEqual(base, DerivedRelationshipRef(knr.PredicateTouches, source, differentNamespace))
	differentProvider := target
	differentProvider.Provider = "gitlab"
	s.NotEqual(base, DerivedRelationshipRef(knr.PredicateTouches, source, differentProvider))
	differentResource := target
	differentResource.ResourceRef = "repo-2"
	s.NotEqual(base, DerivedRelationshipRef(knr.PredicateTouches, source, differentResource))
	s.NotEqual(base, DerivedRelationshipRef(knr.PredicateImpacts, source, target))
	s.NotEqual(base, DerivedRelationshipRef(knr.PredicateTouches, target, source))
}

func (s *ProjectionsSuite) TestRepositoryLinkingAttributes() {
	checkout := LinkingAttributes{LinkingAttributeRepositoryFullName: "acme/checkout"}
	cases := map[string]LinkingAttributes{
		"Acme/Checkout":     checkout,
		"  Acme/Checkout  ": checkout,
		"":                  nil,
		"   ":               nil,
	}
	for fullName, expected := range cases {
		s.Require().Equal(expected, RepositoryLinkingAttributes(fullName), "full name %q", fullName)
	}
}
