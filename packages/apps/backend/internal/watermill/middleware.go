package watermill

import (
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"

	"github.com/rezible/rezible/execution"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	messageMetadataKeyExecutionContext = "ec"
)

func (ms *MessageService) telemetryMiddleware(fn message.HandlerFunc) message.HandlerFunc {
	m := ms.telemetry.DefaultMeter()
	messagesHandled, messagesHandledErr := m.Int64Counter("backend.messages.handled", metric.WithDescription("Watermill messages handled"))
	messageHandlingSeconds, messageHandlingSecondsErr := m.Float64Histogram("backend.messages.handle_duration", metric.WithDescription("Watermill message handling duration"), metric.WithUnit("s"))
	if telErr := errors.Join(messagesHandledErr, messageHandlingSecondsErr); telErr != nil {
		panic("telemetry error: " + telErr.Error())
	}

	return func(msg *message.Message) ([]*message.Message, error) {
		name := ms.marshaller.NameFromMessage(msg)
		topic := message.SubscribeTopicFromCtx(msg.Context())

		start := time.Now()
		out, err := fn(msg)

		attrsOpts := metric.WithAttributes(
			attribute.String("message.name", name),
			attribute.String("topic", topic),
			attribute.Bool("success", err == nil),
		)
		ctx := msg.Context()
		messagesHandled.Add(ctx, 1, attrsOpts)
		messageHandlingSeconds.Record(ctx, time.Since(start).Seconds(), attrsOpts)

		return out, err
	}
}

func (ms *MessageService) addPublisherDecorations(base message.Publisher) (message.Publisher, error) {
	m := ms.telemetry.DefaultMeter()
	messagesPublished, messagesPublishedErr := m.Int64Counter("backend.messages.published", metric.WithDescription("Watermill messages published"))
	if messagesPublishedErr != nil {
		return nil, fmt.Errorf("telemetry error: %w", messagesPublishedErr)
	}

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
	numPoisonedMetric, metricErr := ms.telemetry.DefaultMeter().Int64Counter("backend.messages.poisoned",
		metric.WithDescription("Watermill messages sent to the poison queue"))
	if metricErr != nil {
		return nil, fmt.Errorf("telemetry error: %w", metricErr)
	}

	poisonFilter := func(err error) bool {
		return true
	}

	poisonQueueTopic := "poison.queue"
	poison, poisonErr := middleware.PoisonQueueWithFilter(pub, poisonQueueTopic, poisonFilter)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed initializing poison queue: %w", poisonErr)
	}
	router.AddConsumerHandler("PoisonQueueLogger", poisonQueueTopic, sub, func(msg *message.Message) error {
		numPoisonedMetric.Add(msg.Context(), 1)
		router.Logger().Info("message sent to poison queue", watermill.LogFields{"uuid": msg.UUID})
		return nil
	})

	return poison, nil
}

func (ms *MessageService) setMessageExecutionContext(msg *message.Message) {
	encodedExec, encodeErr := execution.GetContext(msg.Context()).Encode()
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

		exec, decodeErr := execution.RestoreFrom([]byte(encodedExec))
		if decodeErr != nil {
			return nil, fmt.Errorf("restoring execution context: %w", decodeErr)
		}
		msg.SetContext(execution.SetContext(msg.Context(), exec))
		return fn(msg)
	}
}
