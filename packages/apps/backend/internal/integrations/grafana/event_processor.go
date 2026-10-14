package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

// unnamedServicePrefix is what OpenTelemetry reports for a service that sets no name.
const unnamedServicePrefix = "unknown_service"

// EventProcessor normalizes Grafana provider events. It has no dependencies.
type EventProcessor struct{}

// ProcessProviderEvent turns a service listing into one service observation per value. Each value's resource
// ref names the signal, the label and the value, so the alias states which label value a service's telemetry
// carries.
func (p EventProcessor) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	if prov.ProviderEventSource != SourceServices {
		return nil, fmt.Errorf("unknown provider event source: %s", prov.ProviderEventSource)
	}
	var listing serviceListing
	if decodeErr := json.Unmarshal(prov.Attributes, &listing); decodeErr != nil {
		return nil, fmt.Errorf("decode service listing: %w", decodeErr)
	}
	events := make(ent.NormalizedEvents, 0, len(listing.Values))
	for _, value := range listing.Values {
		name := projections.NormalizeServiceName(value)
		if name == "" || strings.HasPrefix(value, unnamedServicePrefix) {
			continue
		}
		attrs := projections.SystemComponentEventAttributes{
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: value,
			LinkingAttributes: projections.LinkingAttributes{
				projections.LinkingAttributeServiceName: name,
			},
		}
		encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
		if encodeErr != nil {
			return nil, fmt.Errorf("encode service %q: %w", value, encodeErr)
		}
		event := &ent.NormalizedEvent{
			Provider:            ProviderName,
			ProviderNamespace:   prov.ProviderNamespace,
			ProviderResourceRef: serviceResourceRef(listing.Signal, listing.Label, value),
			ProviderEventSource: prov.ProviderEventSource,
			ProviderEventRef:    prov.ProviderEventRef,
			Kind:                projections.KindSystemComponent,
			OccurredAt:          prov.ReceivedAt,
			ReceivedAt:          prov.ReceivedAt,
			Attributes:          encodedAttrs,
		}
		events = append(events, event)
	}
	return events, nil
}

// serviceResourceRef is a service's resource ref from a listing: <signal>:<label>:<value>, for example
// logs:service_name:checkout-api.
func serviceResourceRef(signal, label, value string) string {
	return signal + ":" + label + ":" + value
}
