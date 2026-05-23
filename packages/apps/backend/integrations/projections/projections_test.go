package projections

import (
	"testing"

	"github.com/rezible/rezible/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleEventAttributes struct {
	FooBar string `attr:"foo_bar" validate:"required"`
}

func TestEncodeAttributesUsesAttributeTags(t *testing.T) {
	attrs := ExampleEventAttributes{
		FooBar: "baz",
	}
	encoded, encodeErr := EncodeAttributes(attrs)
	require.NoError(t, encodeErr)
	assert.Equal(t, attrs.FooBar, encoded["foo_bar"])
	assert.NotContains(t, encoded, "FooBar")
}

func TestDecodeIncidentObservedEvent(t *testing.T) {
	attrs := IncidentSubjectAttributes{
		ExternalRef: "foo",
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
	}
	encAttrs, encErr := EncodeAttributes(attrs)
	require.NoError(t, encErr)
	ev := &ent.NormalizedEvent{
		Attributes: encAttrs,
	}

	incEv, err := DecodeSubjectAttributes[IncidentSubjectAttributes](ev)
	require.NoError(t, err)
	assert.Equal(t, "Checkout search lookups timing out", incEv.Attributes.Title)
	assert.Equal(t, attrs.SeverityRef, incEv.Attributes.SeverityRef)
	assert.Equal(t, attrs.TypeRef, incEv.Attributes.TypeRef)
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
