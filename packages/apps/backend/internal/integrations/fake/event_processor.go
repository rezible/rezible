package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/projections"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	p := &eventProcessor{event: &prov}
	return p.process()

}

type eventProcessor struct {
	event *rez.ProviderEvent
}

const (
	sourceAlerts    = "alerts"
	sourceIncidents = "incidents"
	sourceTopology  = "system_topology"
)

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderSource {
	case sourceAlerts:
		return p.processAlert()
	case sourceIncidents:
		return p.processIncident()
	case sourceTopology:
		return p.processTopology()
	default:
		return nil, fmt.Errorf("unknown provider source: %s", p.event.ProviderSource)
	}
}

func (p *eventProcessor) processAlert() (ent.NormalizedEvents, error) {
	var payload alertObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal alert observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.AlertSubjectAttributes{
		Title:       payload.Title,
		Description: payload.Description,
		Definition:  payload.Definition,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert observed attributes: %w", encodeErr)
	}

	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceAlerts,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindAlert.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processIncident() (ent.NormalizedEvents, error) {
	var payload incidentObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal incident observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.IncidentSubjectAttributes{
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
		Provider:           integrationName,
		ProviderSource:     sourceIncidents,
		ProviderEventRef:   p.event.ProviderEventRef,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindIncident.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processTopology() (ent.NormalizedEvents, error) {
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceTopology,
		ProviderEventRef:   p.event.ProviderEventRef,
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ActivityKind:       ne.ActivityKindObserved,
		ReceivedAt:         p.event.ReceivedAt,
		OccurredAt:         p.event.ReceivedAt,
	}

	eventErr := fmt.Errorf("unknown topology subject ref: %s", p.event.ProviderSubjectRef)
	if strings.HasPrefix(p.event.ProviderSubjectRef, componentRefPrefix) {
		result.SubjectKind = projections.SubjectKindSystemComponent.String()
		result.Attributes, eventErr = getTopologyComponentAttributes(p.event.Payload)
	} else if strings.HasPrefix(p.event.ProviderSubjectRef, relationshipRefPrefix) {
		result.SubjectKind = projections.SubjectKindSystemRelationship.String()
		result.Attributes, eventErr = getTopologyRelationshipAttributes(p.event.Payload)
	}
	return ent.NormalizedEvents{result}, eventErr
}
