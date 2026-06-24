package projections

import (
	"testing"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleEventAttributes struct {
	FooBar string `json:"foo_bar" validate:"required"`
}

func TestEncodeAttributesUsesJSONTags(t *testing.T) {
	attrs := ExampleEventAttributes{
		FooBar: "baz",
	}
	encoded, encodeErr := EncodeAttributes(attrs)
	require.NoError(t, encodeErr)
	assert.Equal(t, attrs.FooBar, encoded["foo_bar"])
	assert.NotContains(t, encoded, "FooBar")
}

func TestDecodeIncidentObservedEvent(t *testing.T) {
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	attrs := IncidentSubjectAttributes{
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	}
	encAttrs, encErr := EncodeAttributes(attrs)
	require.NoError(t, encErr)
	assert.Equal(t, openedAt.Format(time.RFC3339Nano), encAttrs["opened_at"])
	ev := &ent.NormalizedEvent{
		Attributes: encAttrs,
	}
	incEv, err := DecodeSubjectAttributes[IncidentSubjectAttributes](ev)
	require.NoError(t, err)
	assert.Equal(t, "Checkout search lookups timing out", incEv.Attributes.Title)
	assert.Equal(t, attrs.SeverityRef, incEv.Attributes.SeverityRef)
	assert.Equal(t, attrs.TypeRef, incEv.Attributes.TypeRef)
	assert.True(t, openedAt.Equal(incEv.Attributes.OpenedAt))
}

func TestDecodeWithRejectsMissingRequiredAttributes(t *testing.T) {
	ev := &ent.NormalizedEvent{
		Attributes: map[string]any{
			"foo_bar": "",
		},
	}
	_, err := DecodeSubjectAttributes[ExampleEventAttributes](ev)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed on the 'required' tag")
}

func TestDecodeWithRejectsUnknownAttributes(t *testing.T) {
	ev := &ent.NormalizedEvent{
		Attributes: map[string]any{
			"foo_bar":    "baz",
			"unexpected": true,
		},
	}
	_, err := DecodeSubjectAttributes[ExampleEventAttributes](ev)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid keys")
}

func TestSortRelatedEntityRefs(t *testing.T) {
	refs := []RelatedEntityRef{
		{ExternalRef: "demo:component:search_api", Kind: "service", DisplayName: "Search API"},
		{ExternalRef: "demo:component:elasticsearch_catalog", Kind: "search_cluster", DisplayName: "Elasticsearch Catalog"},
	}

	sortedRefs := SortRelatedEntityRefs(refs)

	assert.Equal(t, "demo:component:elasticsearch_catalog", sortedRefs[0].ExternalRef)
	assert.Equal(t, "demo:component:search_api", sortedRefs[1].ExternalRef)
	assert.Equal(t, "demo:component:search_api", refs[0].ExternalRef)
}
