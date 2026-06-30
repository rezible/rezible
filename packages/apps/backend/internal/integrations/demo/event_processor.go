package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return (&eventProcessor{event: &prov}).process()
}

type eventProcessor struct {
	event *rez.ProviderEvent
}

const (
	sourceAlerts          = "alerts"
	sourceIncidents       = "incidents"
	sourceTopology        = "system_topology"
	sourceUsers           = "users"
	sourceCodeRepos       = "code_repositories"
	sourceCodeChanges     = "code_changes"
	sourceChatMessages    = "chat_messages"
	sourcePlaybooks       = "playbooks"
	sourceIncidentImpacts = "incident_impacts"
)

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderSource {
	case sourceAlerts:
		return p.processAlert()
	case sourceIncidents:
		return p.processIncident()
	case sourceTopology:
		return p.processTopology()
	case sourceUsers:
		return p.processUser()
	case sourceCodeRepos:
		return p.processCodeRepository()
	case sourceCodeChanges:
		return p.processCodeChange()
	case sourceChatMessages:
		return p.processChatMessage()
	case sourcePlaybooks:
		return p.processPlaybook()
	case sourceIncidentImpacts:
		return p.processIncidentImpact()
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
		Title:           payload.Title,
		Description:     payload.Description,
		Definition:      payload.Definition,
		RelatedEntities: payload.RelatedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert observed attributes: %w", encodeErr)
	}

	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceAlerts,
		Kind:               ne.KindObserved,
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
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
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

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceCodeChanges,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindCodeChange.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}}, nil
}

func (p *eventProcessor) processChatMessage() (ent.NormalizedEvents, error) {
	var payload chatMessageObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal chat message observed payload: %w", jsonErr)
	}

	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.ChatMessageAttributes{
		ConversationExternalRef: payload.ConversationExternalRef,
		Body:                    payload.Body,
		SenderExternalRef:       payload.SenderExternalRef,
		ThreadExternalRef:       payload.ThreadExternalRef,
		RelatedEntities:         payload.RelatedEntities,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode chat message attributes: %w", encodeErr)
	}

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceChatMessages,
		Kind:               ne.KindReceived,
		SubjectKind:        projections.SubjectKindChatMessage.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}}, nil
}

func (p *eventProcessor) processPlaybook() (ent.NormalizedEvents, error) {
	var payload playbookObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal playbook observed payload: %w", jsonErr)
	}

	occurredAt := payload.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.PlaybookSubjectAttributes{
		Title:         payload.Title,
		Content:       payload.Content,
		RelatedAlerts: payload.RelatedAlertExternalRefs,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode playbook attributes: %w", encodeErr)
	}

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourcePlaybooks,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindPlaybook.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}}, nil
}

func (p *eventProcessor) processIncidentImpact() (ent.NormalizedEvents, error) {
	var payload incidentImpactObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal incident impact observed payload: %w", jsonErr)
	}

	occurredAt := payload.ObservedAt
	if occurredAt.IsZero() {
		occurredAt = p.event.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	attrs := projections.IncidentImpactSubjectAttributes{
		IncidentExternalRef: payload.IncidentExternalRef,
		EntityExternalRef:   payload.EntityExternalRef,
		EntityKind:          payload.EntityKind,
		EntityDisplayName:   payload.EntityDisplayName,
		Source:              payload.Source,
		Note:                payload.Note,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode incident impact attributes: %w", encodeErr)
	}

	return ent.NormalizedEvents{&ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceIncidentImpacts,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindIncidentImpact.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}}, nil
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
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceTopology,
		ProviderEventRef:   p.event.ProviderEventRef,
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		Kind:               ne.KindObserved,
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
