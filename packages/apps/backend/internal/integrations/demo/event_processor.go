package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	sourceAlerts      = "alerts"
	sourceIncidents   = "incidents"
	sourceUsers       = "users"
	sourceTeams       = "teams"
	sourceTeamMembers = "team_memberships"
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
	switch p.event.ProviderEventSource {
	case sourceAlerts:
		return p.processAlert()
	case sourceIncidents:
		return p.processIncident()
	case sourceUsers:
		return p.processUser()
	case sourceTeams:
		return p.processTeam()
	case sourceTeamMembers:
		return p.processTeamMembership()
	case sourceCodeRepos:
		return p.processCodeRepository()
	case sourceCodeChanges:
		return p.processCodeChange()
	case sourceTopology:
		return p.processTopology()
	default:
		return nil, fmt.Errorf("unknown provider event source: %s", p.event.ProviderEventSource)
	}
}

func (p *eventProcessor) processTeam() (ent.NormalizedEvents, error) {
	var payload teamObservedPayload
	if unmarshalErr := json.Unmarshal(p.event.Attributes, &payload); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal team observed payload: %w", unmarshalErr)
	}
	occurredAt := payload.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	attrs := projections.TeamEventAttributes{
		Name:          payload.Name,
		Slug:          payload.Slug,
		ChatChannelId: payload.ChatChannelID,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode team observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindTeam,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processTeamMembership() (ent.NormalizedEvents, error) {
	var payload teamMembershipObservedPayload
	if unmarshalErr := json.Unmarshal(p.event.Attributes, &payload); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal team membership observed payload: %w", unmarshalErr)
	}
	attrs := projections.TeamMembershipEventAttributes{
		Team: projections.TeamMembershipTeamAttributes{
			ProviderResourceRef: rez.ProviderResourceRef{Provider: providerName, ProviderNamespace: integrationName, ResourceRef: "demo:team:" + payload.TeamResourceID},
			Name:                "Search Platform",
			Slug:                "search-platform",
			ChatChannelId:       "CSEARCH123",
		},
		User: projections.TeamMembershipUserAttributes{
			ProviderResourceRef: rez.ProviderResourceRef{Provider: providerName, ProviderNamespace: integrationName, ResourceRef: "demo:user:" + payload.UserExternalID},
			Name:                "Ava Patel",
			Email:               "ava.patel@rezible.example",
			ChatId:              "UAVA123",
			Timezone:            "Australia/Sydney",
		},
		Role: payload.Role,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode team membership observed attributes: %w", encodeErr)
	}
	occurredAt := p.event.ReceivedAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindTeamMembership,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processAlert() (ent.NormalizedEvents, error) {
	var payload alertObservedPayload
	if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal alert observed payload: %w", err)
	}
	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	attrs := projections.AlertInstanceEventAttributes{
		Title:            payload.Title,
		Description:      payload.Description,
		Definition:       payload.Definition,
		ObservedEntities: payload.ObservedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert observed attributes: %w", encodeErr)
	}
	receivedAt := p.event.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindAlertInstance,
		OccurredAt:          occurredAt,
		ReceivedAt:          receivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processUser() (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal user observed payload: %w", err)
	}
	occurredAt := payload.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	attrs := projections.UserEventAttributes{
		Name:     payload.Name,
		Email:    payload.Email,
		ChatId:   payload.ChatID,
		Timezone: payload.Timezone,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode user observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindUser,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processCodeRepository() (ent.NormalizedEvents, error) {
	var payload codeRepositoryObservedPayload
	if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal code repository observed payload: %w", err)
	}
	occurredAt := payload.ObservedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	attrs := projections.CodeForgeEventAttributes{
		DisplayName: payload.FullName,
		URL:         payload.URL,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode code repository attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindCodeForge,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processCodeChange() (ent.NormalizedEvents, error) {
	var payload codeChangeObservedPayload
	if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal code change observed payload: %w", err)
	}
	occurredAt := payload.MergedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	repository := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          providerName,
			ProviderNamespace: p.event.ProviderNamespace,
			ResourceRef:       payload.RepositoryRef,
		},
		Category:    kne.CategoryCode,
		Kind:        "repository",
		DisplayName: payload.RepositoryRef,
	}
	attrs := projections.CodeChangeEventAttributes{
		Repository:       repository,
		DisplayName:      payload.Title,
		ImpactedEntities: payload.ImpactedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode code change attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindCodeChange,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processIncident() (ent.NormalizedEvents, error) {
	var payload incidentObservedPayload
	if unmarshalErr := json.Unmarshal(p.event.Attributes, &payload); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal incident observed payload: %w", unmarshalErr)
	}
	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	var sourceUpdatedAt *time.Time
	if !payload.SourceUpdatedAt.IsZero() {
		sourceUpdatedAt = &payload.SourceUpdatedAt
	}
	var sourceURL *string
	if payload.SourceURL != "" {
		sourceURL = &payload.SourceURL
	}
	attrs := projections.IncidentEventAttributes{
		Title:           payload.Title,
		Summary:         payload.Summary,
		SeverityRef:     payload.SeverityRef,
		TypeRef:         payload.TypeRef,
		OpenedAt:        occurredAt,
		ResponseState:   payload.ResponseState,
		ResolvedAt:      payload.ResolvedAt,
		SourceUpdatedAt: sourceUpdatedAt,
		SourceURL:       sourceURL,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode incident observed attributes: %w", encodeErr)
	}
	receivedAt := p.event.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: payload.resourceRef(),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindIncident,
		OccurredAt:          occurredAt,
		ReceivedAt:          receivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processTopology() (ent.NormalizedEvents, error) {
	var envelope struct {
		Source topologyRelationshipObservedPayloadComponent `json:"source"`
	}
	if err := json.Unmarshal(p.event.Attributes, &envelope); err != nil {
		return nil, fmt.Errorf("inspect topology payload: %w", err)
	}
	var kind string
	var attributes any
	var resourceRef string
	if envelope.Source.ResourceRef != "" {
		var payload topologyRelationshipObservedPayload
		if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
			return nil, fmt.Errorf("unmarshal topology relationship payload: %w", err)
		}
		kind, attributes = projections.KindSystemRelationship, payload.getAttributes(p.event.ProviderNamespace)
		resourceRef = payload.ResourceRef
	} else {
		var payload topologyComponentObservedPayload
		if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
			return nil, fmt.Errorf("unmarshal topology component payload: %w", err)
		}
		kind, attributes = projections.KindSystemComponent, payload.getAttributes()
		resourceRef = payload.ResourceRef
	}
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode topology attributes: %w", encodeErr)
	}
	occurredAt := p.event.ReceivedAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	result := &ent.NormalizedEvent{
		Provider:            providerName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: resourceRef,
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                kind,
		OccurredAt:          occurredAt,
		ReceivedAt:          occurredAt,
		Attributes:          encodedAttributes,
	}
	return ent.NormalizedEvents{result}, nil
}
