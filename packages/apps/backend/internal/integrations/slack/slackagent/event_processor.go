package slackagent

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	sourceUsers             = "users"
	sourceTeams             = "teams"
	sourceTeamMemberships   = "team_memberships"
	sourceEventsApiCallback = "events_api/callback"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	_ = ctx
	switch prov.ProviderEventSource {
	case sourceUsers:
		return i.processUserObservedEvent(prov)
	case sourceTeams:
		return i.processTeamObservedEvent(prov)
	case sourceTeamMemberships:
		return i.processTeamMembershipObservedEvent(prov)
	case sourceEventsApiCallback:
		return ent.NormalizedEvents{}, nil
	default:
		return nil, fmt.Errorf("unknown provider event source: %s", prov.ProviderEventSource)
	}
}

func (i *Integration) processTeamObservedEvent(ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload teamObservedPayload
	if jsonErr := json.Unmarshal(ev.Attributes, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal team observed payload: %w", jsonErr)
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(payload.makeEventAttributes())
	if encodeErr != nil {
		return nil, fmt.Errorf("encode team observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            slackintegration.ProviderName,
		ProviderNamespace:   ev.ProviderNamespace,
		ProviderResourceRef: payload.SlackID,
		ProviderEventSource: ev.ProviderEventSource,
		Kind:                projections.KindTeam,
		ProviderEventRef:    ev.ProviderEventRef,
		ReceivedAt:          ev.ReceivedAt,
		OccurredAt:          payload.UpdatedAt.Time(),
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (i *Integration) processTeamMembershipObservedEvent(ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload teamMembershipObservedPayload
	if jsonErr := json.Unmarshal(ev.Attributes, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal team membership observed payload: %w", jsonErr)
	}
	teamResourceRef := rez.ProviderResourceRef{
		Provider:          slackintegration.ProviderName,
		ProviderNamespace: ev.ProviderNamespace,
		ResourceRef:       payload.Team.SlackID,
	}
	userResourceRef := rez.ProviderResourceRef{
		Provider:          slackintegration.ProviderName,
		ProviderNamespace: ev.ProviderNamespace,
		ResourceRef:       payload.User.SlackID,
	}
	attrs := projections.TeamMembershipEventAttributes{
		Team: projections.TeamMembershipTeamAttributes{
			ProviderResourceRef: teamResourceRef,
			Name:                payload.Team.Name,
			Slug:                payload.Team.Slug,
			ChatChannelId:       payload.Team.ChatChannelID,
		},
		User: projections.TeamMembershipUserAttributes{
			ProviderResourceRef: userResourceRef,
			Name:                payload.User.Name,
			Email:               payload.User.Email,
			ChatId:              payload.User.SlackID,
			Timezone:            payload.User.Timezone,
		},
		Role: "member",
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode team membership attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            slackintegration.ProviderName,
		ProviderNamespace:   ev.ProviderNamespace,
		ProviderResourceRef: fmt.Sprintf("membership:%s:%s", payload.Team.SlackID, payload.User.SlackID),
		ProviderEventSource: ev.ProviderEventSource,
		Kind:                projections.KindTeamMembership,
		ProviderEventRef:    ev.ProviderEventRef,
		ReceivedAt:          ev.ReceivedAt,
		OccurredAt:          payload.Team.UpdatedAt.Time(),
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (i *Integration) processUserObservedEvent(ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if jsonErr := json.Unmarshal(ev.Attributes, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal user observed payload: %w", jsonErr)
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(payload.makeEventAttributes())
	if encodeErr != nil {
		return nil, fmt.Errorf("encode user observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            slackintegration.ProviderName,
		ProviderNamespace:   ev.ProviderNamespace,
		ProviderResourceRef: payload.SlackID,
		ProviderEventSource: ev.ProviderEventSource,
		Kind:                projections.KindUser,
		ProviderEventRef:    ev.ProviderEventRef,
		ReceivedAt:          ev.ReceivedAt,
		OccurredAt:          payload.UpdatedAt.Time(),
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}
