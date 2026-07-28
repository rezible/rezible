package messages

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"

	rez "github.com/rezible/rezible"
)

func NewEventHandler[T any](name string, handleFn func(context.Context, *T) error) rez.MessageEventHandler {
	return cqrs.NewEventHandler[T](name, handleFn)
}

func NewGroupEventHandler[T any](handleFn func(context.Context, *T) error) cqrs.GroupEventHandler {
	return cqrs.NewGroupEventHandler[T](handleFn)
}
