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
	sourceTopology  = "system_topology"
)

func (p *eventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourceAlerts:
		return p.processAlert(prov)
	case sourceIncidents:
		return p.processIncident(prov)
	case sourceTopology:
		return p.processTopology(prov)
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
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert observed attributes: %w", encodeErr)
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
		Attributes:       encodedAttrs,
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
		ExternalRef: payload.ExternalID,
		Title:       payload.Title,
		Summary:     payload.Summary,
		SeverityRef: payload.SeverityRef,
		TypeRef:     payload.TypeRef,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode incident observed attributes: %w", encodeErr)
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
		Attributes:       encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processTopology(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload topologyObservedPayload
	if jsonErr := json.Unmarshal(prov.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal topology observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = prov.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceTopology,
		ProviderEventRef: prov.ProviderEventRef,
		SubjectRef:       prov.SubjectRef,
		OccurredAt:       occurredAt,
		ReceivedAt:       prov.ReceivedAt,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	payloadErr := fmt.Errorf("unknown topology observation type: %s", payload.ObservationType)
	switch payload.ObservationType {
	case topologyObservationComponent:
		{
			result.Kind = ne.KindSystemComponentObserved
			result.SubjectKind = "system_component"
			result.Attributes, payloadErr = payload.Component.encodeAttributes()
		}
	case topologyObservationRelationship:
		{
			result.Kind = ne.KindSystemRelationshipObserved
			result.SubjectKind = "system_relationship"
			result.Attributes, payloadErr = payload.Relationship.encodeAttributes()
		}
	}
	if payloadErr != nil {
		return nil, fmt.Errorf("topology component payload invalid: %w", payloadErr)
	}

	return ent.NormalizedEvents{result}, nil
}
