package watermill

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var eventMarshaller = cqrs.JSONMarshaler{
	GenerateName: func(v any) string {
		if val := reflect.ValueOf(v); val.Kind() == reflect.Ptr && val.IsNil() {
			return ""
		}
		if msgEvent, ok := v.(rez.MessageEvent); ok {
			if name := msgEvent.MessageName(); name != "" {
				return name
			}
		}
		return cqrs.FullyQualifiedStructName(v)
	},
}

func (ms *MessageQueue) makeEventBus(publisher message.Publisher) (*cqrs.EventBus, error) {
	return cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return ms.eventTopic(params.EventName), nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			ms.setMessageEventScopesMetadata(params.Event, params.Message)
			return nil
		},
		Marshaler: eventMarshaller,
		Logger:    ms.logger,
	})
}

func (ms *MessageQueue) makeEventProcessor() (*cqrs.EventProcessor, error) {
	eventProcCfg := cqrs.EventProcessorConfig{
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return ms.transport.Subscriber(params.HandlerName)
		},
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return ms.eventTopic(params.EventName), nil
		},
		Marshaler: eventMarshaller,
		Logger:    ms.logger,
	}
	return cqrs.NewEventProcessorWithConfig(ms.router, eventProcCfg)
}

func (ms *MessageQueue) makeTelemetryMiddleware(ts rez.TelemetryService) (message.HandlerMiddleware, error) {
	m := ts.DefaultMeter()
	messagesHandled, messagesHandledErr := m.Int64Counter("backend.messages.handled",
		metric.WithDescription("Watermill messages handled"))
	messageHandlingSeconds, messageHandlingSecondsErr := m.Float64Histogram("backend.messages.handle_duration",
		metric.WithDescription("Watermill message handling duration"),
		metric.WithUnit("s"))
	if telErr := errors.Join(messagesHandledErr, messageHandlingSecondsErr); telErr != nil {
		return nil, fmt.Errorf("failed to handle messages: %w", telErr)
	}

	return func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			name := eventMarshaller.NameFromMessage(msg)
			topic := message.SubscribeTopicFromCtx(msg.Context())

			start := time.Now()
			out, err := next(msg)

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
	}, nil
}

func (ms *MessageQueue) makeRetryMiddleware() message.HandlerMiddleware {
	retry := middleware.Retry{
		MaxRetries:      2,
		InitialInterval: time.Second,
		Logger:          ms.logger,
	}
	return retry.Middleware
}

func (ms *MessageQueue) makeThrottleMiddleware() message.HandlerMiddleware {
	throttle := middleware.NewThrottle(10, time.Second)
	return throttle.Middleware
}

func (ms *MessageQueue) attemptTimeout(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		attemptCtx, cancel := context.WithTimeout(msg.Context(), 30*time.Second)
		defer cancel()
		previousCtx := msg.Context()
		msg.SetContext(attemptCtx)
		out, handleErr := next(msg)
		msg.SetContext(previousCtx)
		return out, handleErr
	}
}

const messageMetadataKeyExecutionContext = "ec"

func setMessageExecutionContext(msg *message.Message) error {
	if !execution.ContextExists(msg.Context()) {
		return fmt.Errorf("no execution context")
	}
	encodedExec, encodeErr := execution.GetContext(msg.Context()).Encode()
	if encodeErr != nil {
		return fmt.Errorf("marshal execution context: %w", encodeErr)
	}
	msg.Metadata.Set(messageMetadataKeyExecutionContext, string(encodedExec))
	return nil
}

func (ms *MessageQueue) restoreMessageContext(msg *message.Message) error {
	encodedExec := msg.Metadata.Get(messageMetadataKeyExecutionContext)
	if encodedExec == "" {
		return fmt.Errorf("message execution context metadata %q is missing", messageMetadataKeyExecutionContext)
	}

	exec, decodeErr := execution.DecodeContext([]byte(encodedExec))
	if decodeErr != nil {
		return fmt.Errorf("restoring execution context: %w", decodeErr)
	}

	msg.SetContext(execution.SetContext(msg.Context(), exec))
	return nil
}

func (ms *MessageQueue) restoreMessageAccessScope(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		if err := ms.restoreMessageContext(msg); err != nil {
			return nil, err
		}
		return fn(msg)
	}
}
