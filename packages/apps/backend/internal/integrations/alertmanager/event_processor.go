package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

// EventProcessor normalizes Alertmanager provider events. It has no dependencies.
type EventProcessor struct{}

func (p EventProcessor) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	if prov.ProviderEventSource != sourceAlerts {
		return nil, fmt.Errorf("unknown provider event source: %s", prov.ProviderEventSource)
	}
	var alert alertEventAttributes
	if decodeErr := json.Unmarshal(prov.Attributes, &alert); decodeErr != nil {
		return nil, fmt.Errorf("decode alert event: %w", decodeErr)
	}
	alertname := alert.Labels["alertname"]
	severityRef := alert.Labels["severity"]
	attrs := projections.AlertInstanceEventAttributes{
		Title:            alertname,
		Summary:          alert.Annotations["summary"],
		Description:      alert.Annotations["description"],
		Definition:       alert.GeneratorURL,
		State:            alert.Status,
		InstanceID:       alert.Fingerprint,
		Labels:           alert.Labels,
		Severity:         string(mapSeverity(severityRef)),
		SeverityRef:      severityRef,
		StartedAt:        alert.StartsAt,
		ObservedEntities: []projections.EntityObservation{},
	}
	occurredAt := prov.ReceivedAt
	if alert.Status == projections.AlertStateResolved && alert.EndsAt != nil {
		attrs.EndedAt = alert.EndsAt
		occurredAt = *alert.EndsAt
	}
	if alert.Service != nil {
		attrs.ObservedEntities = append(attrs.ObservedEntities, *alert.Service)
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert instance attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            ProviderName,
		ProviderNamespace:   prov.ProviderNamespace,
		ProviderResourceRef: alertDefinitionRef(alertname, alert.ServiceName),
		ProviderEventSource: prov.ProviderEventSource,
		ProviderEventRef:    prov.ProviderEventRef,
		Kind:                projections.KindAlertInstance,
		OccurredAt:          occurredAt,
		ReceivedAt:          prov.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

// alertDefinitionRef identifies a Rezible alert definition: one per alert name and service, so one upstream
// alerting rule applied to two services is two definitions. JSON null is the explicit no-service value.
func alertDefinitionRef(alertname, serviceName string) string {
	var service any
	if serviceName != "" {
		service = serviceName
	}
	return "alert:" + identityDigest(alertname, service)
}

func mapSeverity(raw string) schematypes.SignalSeverity {
	switch strings.ToLower(raw) {
	case "critical", "page", "p1", "sev1", "high":
		return schematypes.SignalSeverityCritical
	case "warning", "warn", "p2", "p3", "medium":
		return schematypes.SignalSeverityWarning
	case "info", "none", "low", "p4", "p5":
		return schematypes.SignalSeverityInfo
	default:
		return schematypes.SignalSeverityUnknown
	}
}
