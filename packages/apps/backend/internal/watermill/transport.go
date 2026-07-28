package watermill

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type (
	Transport interface {
		Publisher() message.Publisher
		Subscriber(name string) (Subscriber, error)
		AdhocSubscriber(name string) (Subscriber, error)
		Close() error
	}

	Subscriber interface {
		message.Subscriber
		Close() error
	}
)

type goChannelTransport struct {
	pubsub *gochannel.GoChannel
}

func newGoChannelTransport(logger watermill.LoggerAdapter) Transport {
	return &goChannelTransport{
		pubsub: gochannel.NewGoChannel(gochannel.Config{PreserveContext: false}, logger),
	}
}

func (t *goChannelTransport) Publisher() message.Publisher {
	return t.pubsub
}

func (t *goChannelTransport) Subscriber(name string) (Subscriber, error) {
	return t.pubsub, nil
}

type wrappedSubscriber struct {
	message.Subscriber
}

func (s *wrappedSubscriber) Close() error {
	return nil
}

func (t *goChannelTransport) AdhocSubscriber(name string) (Subscriber, error) {
	return &wrappedSubscriber{Subscriber: t.pubsub}, nil
}

func (t *goChannelTransport) Close() error {
	return t.pubsub.Close()
}
