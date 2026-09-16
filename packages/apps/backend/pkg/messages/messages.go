package messages

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	rez "github.com/rezible/rezible"
)

type (
	Transport interface {
		Publisher() message.Publisher
		Subscriber(name string) (message.Subscriber, error)
		Close() error
	}

	TransportWithAdhocSubscriber interface {
		AdhocSubscriber(handlerName string) (message.Subscriber, error)
	}

	EventHandler = rez.MessageEventHandler

	MessageHandlerProvider interface {
		MessageHandlers() []EventHandler
	}

	Definition struct {
		Handlers []EventHandler
	}

	Registrar interface {
		Register(Definition) error
	}
)

func NewEventHandler[T any](name string, handleFn func(context.Context, *T) error) rez.MessageEventHandler {
	return cqrs.NewEventHandler[T](name, handleFn)
}

func NewGroupEventHandler[T any](handleFn func(context.Context, *T) error) cqrs.GroupEventHandler {
	return cqrs.NewGroupEventHandler[T](handleFn)
}

type EventWithScopes interface {
	MessageScopes() rez.MessageEventScopes
}
