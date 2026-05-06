package watermill

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/telemetry"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
)

const (
	messageMetadataKeyExecutionContext = "ec"
)

func (ms *MessageService) telemetryMiddleware(fn message.HandlerFunc) message.HandlerFunc {
	m := telemetry.DefaultMeter()
	messagesHandled := telemetry.Int64CounterInstrument(m, "backend.messages.handled", "Watermill messages handled")
	messageHandlingSeconds := telemetry.Float64HistogramInstrument(m, "backend.messages.handle_duration", "Watermill message handling duration", "s")

	return func(msg *message.Message) ([]*message.Message, error) {
		name := ms.marshaller.NameFromMessage(msg)
		topic := message.SubscribeTopicFromCtx(msg.Context())

		start := time.Now()
		out, err := fn(msg)

		attrsOpts := telemetry.WithMetricAttributes(
			telemetry.StringAttr("message.name", telemetry.NormalizeLabel(name)),
			telemetry.StringAttr("topic", telemetry.NormalizeLabel(topic)),
			telemetry.ResultAttr(err),
		)
		ctx := msg.Context()
		messagesHandled.Add(ctx, 1, attrsOpts)
		messageHandlingSeconds.Record(ctx, time.Since(start).Seconds(), attrsOpts)

		return out, err
	}
}

func (ms *MessageService) addPublisherDecorations(base message.Publisher) (message.Publisher, error) {
	m := telemetry.DefaultMeter()
	messagesPublished := telemetry.Int64CounterInstrument(m, "backend.messages.published", "Watermill messages published")

	decorator := message.MessageTransformPublisherDecorator(func(msg *message.Message) {
		messagesPublished.Add(msg.Context(), 1)
		ms.setMessageExecutionContext(msg)
	})

	pub, pubErr := decorator(base)
	if pubErr != nil {
		return nil, fmt.Errorf("failed to decorate publisher: %w", pubErr)
	}

	pub = wotel.NewNamedPublisherDecorator("pub", wotelfloss.NewTracePropagatingPublisherDecorator(pub))

	return pub, nil
}

func (ms *MessageService) setupPoisonQueue(router *message.Router, pub message.Publisher, sub message.Subscriber) (message.HandlerMiddleware, error) {
	messagesPoisoned := telemetry.Int64CounterInstrument(telemetry.DefaultMeter(), "backend.messages.poisoned",
		"Watermill messages sent to the poison queue")

	poisonFilter := func(err error) bool {
		return true
	}

	poisonQueueTopic := "poison.queue"
	poison, poisonErr := middleware.PoisonQueueWithFilter(pub, poisonQueueTopic, poisonFilter)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed initializing poison queue: %w", poisonErr)
	}
	router.AddConsumerHandler("PoisonQueueLogger", poisonQueueTopic, sub, func(msg *message.Message) error {
		messagesPoisoned.Add(msg.Context(), 1)
		router.Logger().Info("message sent to poison queue", watermill.LogFields{"uuid": msg.UUID})
		return nil
	})

	return poison, nil
}

func (ms *MessageService) setMessageExecutionContext(msg *message.Message) {
	encodedExec, encodeErr := execution.FromContext(msg.Context()).Encode()
	if encodeErr != nil {
		ms.logger.Error("failed to marshal execution context", encodeErr, nil)
		return
	}
	msg.Metadata.Set(messageMetadataKeyExecutionContext, string(encodedExec))
}

func (ms *MessageService) restoreMessageAccessScope(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		encodedExec := msg.Metadata.Get(messageMetadataKeyExecutionContext)
		if encodedExec == "" {
			return fn(msg)
		}

		exec, decodeErr := execution.Decode([]byte(encodedExec))
		if decodeErr != nil {
			return nil, fmt.Errorf("restoring execution context: %w", decodeErr)
		}
		exec.Provenance.ParentKind = "message"
		exec.Provenance.ParentID = msg.UUID
		msg.SetContext(execution.StoreInContext(msg.Context(), exec))
		return fn(msg)
	}
}
