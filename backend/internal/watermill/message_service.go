package watermill

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/rs/zerolog/log"
)

var (
	logger = watermill.NewStdLogger(false, false)
)

type MessageService struct {
	router     *message.Router
	publisher  message.Publisher
	subscriber message.Subscriber
}

func NewMessageService() (*MessageService, error) {
	router, routerErr := message.NewRouter(message.RouterConfig{}, logger)
	if routerErr != nil {
		return nil, fmt.Errorf("faile initializing message router: %w", routerErr)
	}

	router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,
		middleware.Recoverer,
	)

	ms := &MessageService{
		router: router,
	}

	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	ms.publisher = pubSub
	ms.subscriber = pubSub

	/*
		handler := router.AddHandler(
			"struct_handler",          // handler name, must be unique
			"incoming_messages_topic", // topic from which we will read events
			pubSub,
			"outgoing_messages_topic", // topic to which we will publish events
			pubSub,
			structHandler{}.Handler,
		)

		// Handler level middleware is only executed for a specific handler
		// Such middleware can be added the same way the router level ones
		handler.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
			return func(message *message.Message) ([]*message.Message, error) {
				//log.Println("executing handler specific middleware for ", message.UUID)

				return h(message)
			}
		})
	*/

	printMessages := func(msg *message.Message) error {
		log.Debug().
			Str("uuid", msg.UUID).
			Str("payload", string(msg.Payload)).
			Interface("metadata", msg.Metadata).
			Msg("received message")
		return nil
	}

	router.AddConsumerHandler(
		"print_incoming_messages",
		"incoming_messages_topic",
		pubSub,
		printMessages,
	)

	router.AddConsumerHandler(
		"print_outgoing_messages",
		"outgoing_messages_topic",
		pubSub,
		printMessages,
	)

	return ms, nil
}

func (ms *MessageService) Start(ctx context.Context) error {
	log.Debug().Msg("starting message service")
	return ms.router.Run(ctx)
}

func (ms *MessageService) Publish(topic string, msgs ...*message.Message) error {
	return ms.publisher.Publish(topic, msgs...)
}

func (ms *MessageService) AddHandler(name, subTopic, pubTopic string, fn message.HandlerFunc, mw ...message.HandlerMiddleware) {
	handler := ms.router.AddHandler(name, subTopic, ms.subscriber, pubTopic, ms.publisher, fn)
	handler.AddMiddleware(mw...)
}

func (ms *MessageService) AddConsumerHandler(name, subTopic string, fn message.NoPublishHandlerFunc, mw ...message.HandlerMiddleware) {
	handler := ms.router.AddConsumerHandler(name, subTopic, ms.subscriber, fn)
	handler.AddMiddleware(mw...)
}

func (ms *MessageService) Stop(ctx context.Context) error {
	log.Debug().Msg("stopping message service")
	if !ms.router.IsRunning() {
		return nil
	}
	return ms.router.Close()
}
