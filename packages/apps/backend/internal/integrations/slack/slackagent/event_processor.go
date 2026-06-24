package slackagent

import (
	"context"
	"encoding/json"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/projections"
)

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	p := &eventProcessor{event: &prov}
	return p.process()
}

type eventProcessor struct {
	event *rez.ProviderEvent
}

const (
	sourceUsers             = "users"
	sourceEventsApiCallback = "events_api/callback"
)

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderSource {
	case sourceUsers:
		return p.processUserObserved()
	case sourceEventsApiCallback:
		return p.processEventsApiCallback()
	default:
		return nil, fmt.Errorf("unknown provider source: %s", p.event.ProviderSource)
	}
}

func (p *eventProcessor) processUserObserved() (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if jsonErr := json.Unmarshal(p.event.Payload, &payload); jsonErr != nil {
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
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindUser.String(),
		ProviderSubjectRef: p.event.ProviderSubjectRef,
		ProviderEventRef:   p.event.ProviderEventRef,
		OccurredAt:         payload.UpdatedAt.Time(),
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

var processCallbackEventTypes = mapset.NewSet(
	slackevents.AppMention,
	slackevents.Message,
)

func (p *eventProcessor) processEventsApiCallback() (ent.NormalizedEvents, error) {
	ev, parseErr := slackevents.ParseEvent(p.event.Payload, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return nil, fmt.Errorf("parse event: %w", parseErr)
	}

	providerEventRef := p.event.ProviderEventRef
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

	occurredAt := slackintegration.TryConvertSlackTs(ts, slackintegration.TryConvertSlackTs(eventTS, p.event.ReceivedAt))

	receivedAt := p.event.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}

	ProviderSubjectRef := p.event.ProviderSubjectRef
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
		ActivityKind:       ne.ActivityKindReceived,
		SubjectKind:        projections.SubjectKindChatMessage.String(),
		ProviderSubjectRef: ProviderSubjectRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         receivedAt,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}
