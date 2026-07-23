package slackagent

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	sourceUsers             = "users"
	sourceEventsApiCallback = "events_api/callback"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourceUsers:
		return i.processUserObservedEvent(prov)
	case sourceEventsApiCallback:
		return i.processEventsApiCallbackEvent(prov)
	default:
		return nil, fmt.Errorf("unknown provider source: %s", prov.ProviderSource)
	}
}

func (i *Integration) processUserObservedEvent(ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if jsonErr := json.Unmarshal(ev.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal userObservedPayload: %w", jsonErr)
	}

	attrs := projections.UserSubjectAttributes{
		Name:     payload.Name,
		Email:    payload.Email,
		ChatId:   payload.SlackID,
		Timezone: payload.Timezone,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode user observed attributes: %w", encodeErr)
	}

	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceUsers,
		Kind:               ne.KindObserved,
		SubjectKind:        projections.SubjectKindUser.String(),
		ProviderSubjectRef: ev.ProviderSubjectRef,
		ProviderEventRef:   ev.ProviderEventRef,
		ReceivedAt:         ev.ReceivedAt,
		OccurredAt:         payload.UpdatedAt.Time(),
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

func (i *Integration) processEventsApiCallbackEvent(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return ent.NormalizedEvents{}, nil
	/*
		ev, parseErr := slackevents.ParseEvent(prov.Payload, slackevents.OptionNoVerifyToken())
		if parseErr != nil {
			return nil, fmt.Errorf("parse event: %w", parseErr)
		}

		providerEventRef := prov.ProviderEventRef
		if cb, ok := ev.Data.(*slackevents.EventsAPICallbackEvent); ok {
			providerEventRef = cb.EventID
		}

		var attrs projections.ChatMessageAttributes
		var ts string
		var eventTS string
		switch data := ev.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			attrs.ConversationExternalRef = data.Channel
			attrs.SenderExternalRef = data.User
			attrs.Body = data.Text
			attrs.ThreadExternalRef = data.ThreadTimeStamp

			ts = data.TimeStamp
			eventTS = data.EventTimeStamp
		case *slackevents.AppMentionEvent:
			attrs.ConversationExternalRef = data.Channel
			attrs.SenderExternalRef = data.User
			attrs.Body = data.Text
			attrs.ThreadExternalRef = data.ThreadTimeStamp

			ts = data.TimeStamp
			eventTS = data.EventTimeStamp
		default:
			return nil, nil
		}

		if attrs.ConversationExternalRef == "" || ts == "" {
			return nil, nil
		}

		occurredAt := slackintegration.TryConvertSlackTs(ts, slackintegration.TryConvertSlackTs(eventTS, prov.ReceivedAt))

		receivedAt := prov.ReceivedAt
		if receivedAt.IsZero() {
			receivedAt = occurredAt
		}

		ProviderSubjectRef := prov.ProviderSubjectRef
		if ProviderSubjectRef == "" {
			ProviderSubjectRef = fmt.Sprintf("slack:%s:%s:%s", ev.TeamID, attrs.ConversationExternalRef, ts)
		}
		if providerEventRef == "" {
			providerEventRef = ProviderSubjectRef
		}
		encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
		if encodeErr != nil {
			return nil, fmt.Errorf("encode chat message attributes: %w", encodeErr)
		}

		result := &ent.NormalizedEvent{
			Provider:           integrationName,
			ProviderSource:     sourceEventsApiCallback,
			ProviderEventRef:   providerEventRef,
			Kind:               ne.KindReceived,
			SubjectKind:        projections.SubjectKindChatMessage.String(),
			ProviderSubjectRef: ProviderSubjectRef,
			OccurredAt:         occurredAt,
			ReceivedAt:         receivedAt,
			Attributes:         encodedAttrs,
		}

		return ent.NormalizedEvents{result}, nil

	*/
}
