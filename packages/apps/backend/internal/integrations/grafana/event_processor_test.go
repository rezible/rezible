package grafana

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

func TestServicesProcessor(t *testing.T) {
	receivedAt := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	listing := serviceListing{
		Signal: "logs",
		Label:  "service_name",
		Values: []string{"", "---", "Checkout_API", "payments", "unknown_service:java"},
	}
	attributes, encodeErr := json.Marshal(listing)
	require.NoError(t, encodeErr)
	listingEvent := rez.ProviderEvent{
		Provider:            ProviderName,
		ProviderNamespace:   "installation-1",
		ProviderEventSource: SourceServices,
		ProviderEventRef:    listing.eventRef(),
		Attributes:          attributes,
		ReceivedAt:          receivedAt,
	}

	events, processErr := EventProcessor{}.ProcessProviderEvent(t.Context(), listingEvent)

	require.NoError(t, processErr)
	// Each resource ref's display name and linking name.
	expected := map[string][2]string{
		"logs:service_name:Checkout_API": {"Checkout_API", "checkout-api"},
		"logs:service_name:payments":     {"payments", "payments"},
	}
	require.Len(t, events, len(expected))
	for _, event := range events {
		names, found := expected[event.ProviderResourceRef]
		require.True(t, found, "unexpected resource ref %q", event.ProviderResourceRef)
		require.Equal(t, ProviderName, event.Provider)
		require.Equal(t, "installation-1", event.ProviderNamespace)
		require.Equal(t, SourceServices, event.ProviderEventSource)
		require.Equal(t, listingEvent.ProviderEventRef, event.ProviderEventRef)
		require.Equal(t, projections.KindSystemComponent, event.Kind)
		require.Equal(t, receivedAt, event.OccurredAt)

		decoded, decodeErr := projections.DecodeSystemComponentEvent(event)
		require.NoError(t, decodeErr)
		attrs := decoded.Attributes
		require.Equal(t, kne.CategoryContainer, attrs.Category)
		require.Equal(t, "service", attrs.Kind)
		require.Equal(t, names[0], attrs.DisplayName)
		require.Equal(t, projections.LinkingAttributes{projections.LinkingAttributeServiceName: names[1]}, attrs.LinkingAttributes)
	}
}
