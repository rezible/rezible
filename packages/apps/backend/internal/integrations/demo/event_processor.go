package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	sourceAlerts      = "alerts"
	sourceIncidents   = "incidents"
	sourceUsers       = "users"
	sourceCodeRepos   = "code_repositories"
	sourceCodeChanges = "code_changes"
	sourceTopology    = "system_topology"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return (&eventProcessor{event: &prov}).process()
}

type eventProcessor struct {
	event *rez.ProviderEvent
}

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderSource {
	case sourceAlerts:
		return p.processAlert()
	case sourceIncidents:
		return p.processIncident()
	case sourceUsers:
		return p.processUser()
	case sourceCodeRepos:
		return p.processCodeRepository()
	case sourceCodeChanges:
		return p.processCodeChange()
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

	attrs := projections.AlertInstanceSubjectAttributes{
		Title:           payload.Title,
		Description:     payload.Description,
		Definition:      payload.Definition,
		ExternalRef:     payload.ExternalRef,
		RelatedEntities: payload.RelatedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert observed attributes: %w", encodeErr)
	}

	result := &ent.NormalizedEvent{
		Kind:               ne.KindObserved,
		Provider:           integrationName,
		ProviderSource:     sourceAlerts,
		ProviderEventRef:   p.event.ProviderEventRef,
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		SubjectKind:        projections.SubjectKindAlertInstance.String(),
		Attributes:         encodedAttrs,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processUser() (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal user observed payload: %w", jsonErr)
	}

	occurredAt := payload.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.UserSubjectAttributes{
		Name:     payload.Name,
		Email:    payload.Email,
		ChatId:   payload.ChatID,
		Timezone: payload.Timezone,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode user observed attributes: %w", encodeErr)
	}

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceUsers,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindUser.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}}, nil
}

func (p *eventProcessor) processCodeRepository() (ent.NormalizedEvents, error) {
	var payload codeRepositoryObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal code repository observed payload: %w", jsonErr)
	}

	occurredAt := payload.ObservedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.CodeForgeSubjectAttributes{
		DisplayName: payload.FullName,
		URL:         payload.URL,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode code repository attributes: %w", encodeErr)
	}

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceCodeRepos,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindCodeForge.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		ReceivedAt:         p.event.ReceivedAt,
		OccurredAt:         occurredAt,
		Attributes:         encodedAttrs,
	}}, nil
}

func (p *eventProcessor) processCodeChange() (ent.NormalizedEvents, error) {
	var payload codeChangeObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal code change observed payload: %w", jsonErr)
	}

	occurredAt := payload.MergedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: payload.RepositoryExternalRef,
		DisplayName:           payload.Title,
		RelatedEntities:       payload.RelatedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode code change attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceCodeChanges,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindCodeChange.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
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
		ExternalRef: payload.ExternalRef,
		Title:       payload.Title,
		Summary:     payload.Summary,
		SeverityRef: payload.SeverityRef,
		TypeRef:     payload.TypeRef,
		OpenedAt:    occurredAt,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode incident observed attributes: %w", encodeErr)
	}

	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceIncidents,
		ProviderEventRef:   p.event.ProviderEventRef,
		Kind:               ne.KindObserved,
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
	return ent.NormalizedEvents{}, nil
}
