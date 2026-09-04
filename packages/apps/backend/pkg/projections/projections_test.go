package projections

import (
	"testing"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleEventAttributes struct {
	FooBar string `json:"foo_bar" validate:"required"`
}

func TestDecodeIncidentObservedEvent(t *testing.T) {
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	attrs := IncidentEventAttributes{
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	}
	encAttrs, encErr := EncodeAttributes(attrs)
	require.NoError(t, encErr)
	ev := &ent.NormalizedEvent{Attributes: encAttrs}
	incEv, err := DecodeEventAttributes[IncidentEventAttributes](ev)
	require.NoError(t, err)
	assert.Equal(t, "Checkout search lookups timing out", incEv.Attributes.Title)
	assert.Equal(t, attrs.SeverityRef, incEv.Attributes.SeverityRef)
	assert.Equal(t, attrs.TypeRef, incEv.Attributes.TypeRef)
	assert.True(t, openedAt.Equal(incEv.Attributes.OpenedAt))
}

func TestDecodeWithRejectsMissingRequiredAttributes(t *testing.T) {
	ev := &ent.NormalizedEvent{
		Attributes: []byte("{}"),
	}
	_, err := DecodeEventAttributes[ExampleEventAttributes](ev)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed on the 'required' tag")
}

func TestSortEntityObservations(t *testing.T) {
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

	assert.Equal(t, "demo:component:elasticsearch_catalog", sortedRefs[0].Ref.ResourceRef)
	assert.Equal(t, "demo:component:search_api", sortedRefs[1].Ref.ResourceRef)
	assert.Equal(t, "demo:component:search_api", refs[0].Ref.ResourceRef)
}

func TestUserEntityLinkingAttributesNormalizeEmail(t *testing.T) {
	attrs := UserEntityLinkingAttributes{Email: "  Alice@Example.COM "}
	assert.Equal(t, map[string]string{"user.email": "alice@example.com"}, attrs.Values())
	assert.Nil(t, (UserEntityLinkingAttributes{Email: "   "}).Values())
}

func TestDerivedRelationshipRefUsesCompleteDirectionalIdentity(t *testing.T) {
	source := rez.ProviderResourceRef{Provider: "github", ProviderNamespace: "org-1", ResourceRef: "change-1"}
	target := rez.ProviderResourceRef{Provider: "github", ProviderNamespace: "org-1", ResourceRef: "repo-1"}

	base := DerivedRelationshipRef(knr.PredicateTouches, source, target)
	assert.Equal(t, "rezible", base.Provider)
	assert.Empty(t, base.ProviderNamespace)
	assert.Contains(t, base.ResourceRef, "relationship:v1:")
	assert.Equal(t, base, DerivedRelationshipRef(knr.PredicateTouches, source, target))

	differentNamespace := target
	differentNamespace.ProviderNamespace = "org-2"
	assert.NotEqual(t, base, DerivedRelationshipRef(knr.PredicateTouches, source, differentNamespace))
	differentProvider := target
	differentProvider.Provider = "gitlab"
	assert.NotEqual(t, base, DerivedRelationshipRef(knr.PredicateTouches, source, differentProvider))
	differentResource := target
	differentResource.ResourceRef = "repo-2"
	assert.NotEqual(t, base, DerivedRelationshipRef(knr.PredicateTouches, source, differentResource))
	assert.NotEqual(t, base, DerivedRelationshipRef(knr.PredicateImpacts, source, target))
	assert.NotEqual(t, base, DerivedRelationshipRef(knr.PredicateTouches, target, source))
}
