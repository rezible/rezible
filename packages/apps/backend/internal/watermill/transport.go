package watermill

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/messages"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
	"go.opentelemetry.io/otel/metric"
)

type publisherWrapper struct {
	telemetry rez.TelemetryService
}

func newPublisherWrapper(tel rez.TelemetryService) *publisherWrapper {
	return &publisherWrapper{telemetry: tel}
}

func (w *publisherWrapper) wrap(pub message.Publisher) (message.Publisher, error) {
	telPub, wrapTelErr := w.wrapPublisherTelemetry(pub)
	if wrapTelErr != nil {
		return nil, fmt.Errorf("telemetry wrapper: %w", wrapTelErr)
	}
	pub = telPub

	pub = &messageExecutionContextPublisher{Publisher: pub}

	return pub, nil
}

func (w *publisherWrapper) wrapPublisherTelemetry(base message.Publisher) (message.Publisher, error) {
	m := w.telemetry.DefaultMeter()
	messagesPublished, messagesPublishedErr := m.Int64Counter("backend.messages.published",
		metric.WithDescription("Watermill messages published"))
	if messagesPublishedErr != nil {
		return nil, fmt.Errorf("messages published counter: %w", messagesPublishedErr)
	}
	msgMetricsFn := func(msg *message.Message) {
		messagesPublished.Add(msg.Context(), 1)
	}

	wrapped, wrapErr := message.MessageTransformPublisherDecorator(msgMetricsFn)(base)
	if wrapErr != nil {
		return nil, fmt.Errorf("wrap metric decorator: %w", wrapErr)
	}

	wrapped = wotel.NewNamedPublisherDecorator("pub", wotelfloss.NewTracePropagatingPublisherDecorator(wrapped))

	return wrapped, nil
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
