package projections

import (
	"testing"
	"time"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleEventAttributes struct {
	FooBar string `json:"foo_bar" validate:"required"`
}

func TestDecodeIncidentObservedEvent(t *testing.T) {
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	attrs := IncidentSubjectAttributes{
		ExternalRef: "foo-bar",
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	}
	encAttrs, encErr := EncodeAttributes(attrs)
	require.NoError(t, encErr)
	ev := &ent.NormalizedEvent{Attributes: encAttrs}
	incEv, err := DecodeSubjectAttributes[IncidentSubjectAttributes](ev)
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
	_, err := DecodeSubjectAttributes[ExampleEventAttributes](ev)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed on the 'required' tag")
}

func TestSortRelatedEntityRefs(t *testing.T) {
	refs := []RelatedEntityRef{
		{ExternalRef: "demo:component:search_api", Kind: kne.KindContainer, Subkind: "service", DisplayName: "Search API"},
		{ExternalRef: "demo:component:elasticsearch_catalog", Kind: kne.KindContainer, Subkind: "search_cluster", DisplayName: "Elasticsearch Catalog"},
	}

	sortedRefs := SortRelatedEntityRefs(refs)

	assert.Equal(t, "demo:component:elasticsearch_catalog", sortedRefs[0].ExternalRef)
	assert.Equal(t, "demo:component:search_api", sortedRefs[1].ExternalRef)
	assert.Equal(t, "demo:component:search_api", refs[0].ExternalRef)
}
