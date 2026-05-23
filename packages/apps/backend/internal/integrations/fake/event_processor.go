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
		ProviderSubjectRef: prov.ProviderSubjectRef,
		ProviderEventRef:   prov.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         prov.ReceivedAt,
		Attributes:         encodedAttrs,
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

	attrs := projections.IncidentSubjectAttributes{
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
		Provider:           integrationName,
		ProviderSource:     sourceIncidents,
		ProviderEventRef:   prov.ProviderEventRef,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindIncident.String(),
		ProviderSubjectRef: prov.ProviderSubjectRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         prov.ReceivedAt,
		Attributes:         encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processTopology(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceTopology,
		ProviderEventRef:   prov.ProviderEventRef,
		ProviderSubjectRef: prov.ProviderSubjectRef,
		ActivityKind:       ne.ActivityKindObserved,
		ReceivedAt:         prov.ReceivedAt,
		OccurredAt:         prov.ReceivedAt,
	}

	eventErr := fmt.Errorf("unknown topology subject ref: %s", prov.ProviderSubjectRef)
	if strings.HasPrefix(prov.ProviderSubjectRef, componentRefPrefix) {
		result.SubjectKind = projections.SubjectKindSystemComponent.String()
		result.Attributes, eventErr = getTopologyComponentAttributes(prov.Payload)
	} else if strings.HasPrefix(prov.ProviderSubjectRef, relationshipRefPrefix) {
		result.SubjectKind = projections.SubjectKindSystemRelationship.String()
		result.Attributes, eventErr = getTopologyRelationshipAttributes(prov.Payload)
	}
	return ent.NormalizedEvents{result}, eventErr
}
