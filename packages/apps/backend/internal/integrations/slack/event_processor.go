package slack

import (
	"context"
	"encoding/json"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/eventprojections"
	"github.com/slack-go/slack/slackevents"
)

type eventProcessor struct {
	services *rez.Services
}

const (
	sourceUsers             = "users"
	sourceEventsApiCallback = "events_api/callback"
)

func (p *eventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourceUsers:
		return p.processUserObserved(prov)
	case sourceEventsApiCallback:
		return p.processEventsApiCallback(prov)
	}
	return nil, fmt.Errorf("unknown provider source: %s", prov.ProviderSource)
}

func (p *eventProcessor) processUserObserved(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload userObservedPayload
	if err := json.Unmarshal(prov.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal repository observed event: %w", err)
	}

	attrs := eventprojections.ChatMessageAttributes{
		ConversationExternalRef: "",
		Body:                    "",
		SenderExternalRef:       "",
		ThreadExternalRef:       "",
	}

	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceUsers,
		Kind:             ne.KindUserObserved,
		SubjectKind:      "user",
		SubjectRef:       prov.SubjectRef,
		ProviderEventRef: prov.ProviderEventRef,
		OccurredAt:       payload.UpdatedAt.Time(),
		ReceivedAt:       prov.ReceivedAt,
		Attributes:       attrs.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}

var processCallbackEventTypes = mapset.NewSet(
	slackevents.AppMention,
	slackevents.Message,
)

func (p *eventProcessor) processEventsApiCallback(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	ev, parseErr := slackevents.ParseEvent(prov.Payload, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return nil, fmt.Errorf("parse event: %w", parseErr)
	}

	result := &ent.NormalizedEvent{
		Provider:       integrationName,
		ProviderSource: sourceEventsApiCallback,
		Kind:           ne.KindChatMessage,
		SubjectKind:    "message",
	}
	if cb, ok := ev.Data.(*slackevents.EventsAPICallbackEvent); ok {
		result.ProviderEventRef = cb.EventID
	}

	attrs := eventprojections.ChatMessageAttributes{
		ConversationExternalRef: "",
		Body:                    "",
		SenderExternalRef:       "",
		ThreadExternalRef:       "",
	}

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

	occurredAt := tryConvertTs(ts, tryConvertTs(eventTS, prov.ReceivedAt))
	receivedAt := prov.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}
	result.OccurredAt = occurredAt
	result.ReceivedAt = receivedAt

	result.SubjectRef = prov.SubjectRef
	if result.SubjectRef == "" {
		result.SubjectRef = fmt.Sprintf("slack:%s:%s:%s", ev.TeamID, attrs.ConversationExternalRef, ts)
	}
	if result.ProviderEventRef == "" {
		result.ProviderEventRef = result.SubjectRef
	}

	result.Attributes = attrs.Encode()

	return ent.NormalizedEvents{result}, nil
}
