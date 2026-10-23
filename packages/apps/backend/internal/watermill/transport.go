package watermill

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	"github.com/rezible/rezible/pkg/messages"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// decorateEventsPublisher carries the execution context and trace context in the message metadata, inside a
// named publish span.
func decorateEventsPublisher(pub message.Publisher) (message.Publisher, error) {
	traced := wotel.NewNamedPublisherDecorator("pub", wotelfloss.NewTracePropagatingPublisherDecorator(pub))
	return &messageExecutionContextPublisher{Publisher: traced}, nil
}

// makeCountedTransportPublisher counts the messages handed to the transport: forwarded and live events.
func (ms *MessageQueue) makeCountedTransportPublisher() (message.Publisher, error) {
	messagesPublished, messagesPublishedErr := otel.Meter("github.com/rezible/rezible").Int64Counter("backend.messages.published",
		metric.WithDescription("Watermill messages published"))
	if messagesPublishedErr != nil {
		return nil, fmt.Errorf("messages published counter: %w", messagesPublishedErr)
	}
	countMessage := func(msg *message.Message) {
		messagesPublished.Add(msg.Context(), 1)
	}
	return message.MessageTransformPublisherDecorator(countMessage)(ms.transport.Publisher())
}

type messageExecutionContextPublisher struct {
	message.Publisher
}

func (p messageExecutionContextPublisher) Publish(topic string, messages ...*message.Message) error {
	for _, msg := range messages {
		if ctxErr := setMessageExecutionContext(msg); ctxErr != nil {
			return ctxErr
		}
	}
	return p.Publisher.Publish(topic, messages...)
}

type goChannelTransport struct {
	pubsub *gochannel.GoChannel
}

func NewGoChannelTransport() messages.Transport {
	cfg := gochannel.Config{
		PreserveContext: false,
	}
	pubsub := gochannel.NewGoChannel(cfg, watermill.NewSlogLogger(slog.Default()))
	return &goChannelTransport{pubsub: pubsub}
}

func (t *goChannelTransport) Publisher() message.Publisher {
	return t.pubsub
}

func (t *goChannelTransport) Subscriber(name string) (message.Subscriber, error) {
	return &wrappedSubscriber{Subscriber: t.pubsub}, nil
}

func (t *goChannelTransport) AdhocSubscriber(name string) (message.Subscriber, error) {
	return &wrappedSubscriber{Subscriber: t.pubsub}, nil
}

func (t *goChannelTransport) Close() error {
	return t.pubsub.Close()
}

type wrappedSubscriber struct {
	message.Subscriber
}

func (s *wrappedSubscriber) Close() error {
	return nil
}
