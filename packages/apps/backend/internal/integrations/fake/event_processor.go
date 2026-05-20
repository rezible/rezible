package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/projections"
)

func (i *integration) MakeProviderEventProcessor() rez.ProviderEventProcessor {
	return &eventProcessor{}
}

type eventProcessor struct {
}

const (
	sourceAlerts    = "alerts"
	sourceIncidents = "incidents"
)

func (p *eventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourceAlerts:
		return p.processAlert(prov)
	case sourceIncidents:
		return p.processIncident(prov)
	default:
		return nil, fmt.Errorf("unknown provider source: %s", prov.ProviderSource)
	}
}

func (p *eventProcessor) processAlert(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload alertObservedPayload
	if jsonErr := json.Unmarshal(prov.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal alert observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = prov.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.AlertObservedAttributes{
		Title:       payload.Title,
		Description: payload.Description,
		Definition:  payload.Definition,
	}

	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceAlerts,
		Kind:             ne.KindAlertObserved,
		SubjectKind:      "alert",
		SubjectRef:       prov.SubjectRef,
		ProviderEventRef: prov.ProviderEventRef,
		OccurredAt:       occurredAt,
		ReceivedAt:       prov.ReceivedAt,
		Attributes:       attrs.Encode(),
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processIncident(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload incidentObservedPayload
	if jsonErr := json.Unmarshal(prov.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal incident observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = prov.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.IncidentObservedAttributes{
		Title:        payload.Title,
		Summary:      payload.Summary,
		SeverityName: payload.SeverityName,
		SeverityRank: payload.SeverityRank,
		TypeName:     payload.TypeName,
	}

	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceIncidents,
		ProviderEventRef: prov.ProviderEventRef,
		Kind:             ne.KindIncidentObserved,
		SubjectKind:      "incident",
		SubjectRef:       prov.SubjectRef,
		OccurredAt:       occurredAt,
		ReceivedAt:       prov.ReceivedAt,
		Attributes:       attrs.Encode(),
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}
