package watermill

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

type MessageService struct {
	router     *message.Router
	publisher  message.Publisher
	subscriber message.Subscriber
}

func NewMessageService() (*MessageService, error) {
	slogOpts := slogzerolog.Option{
		Level:  slog.LevelDebug,
		Logger: zerolog.DefaultContextLogger,
	}

	logger := watermill.NewSlogLogger(slog.New(slogOpts.NewZerologHandler()))
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	cfg := message.RouterConfig{
		CloseTimeout: time.Second * 5,
	}
	router, routerErr := message.NewRouter(cfg, logger)
	if routerErr != nil {
		return nil, fmt.Errorf("failed initializing message router: %w", routerErr)
	}

	if mwErr := addMiddleware(router, logger, pubSub, pubSub); mwErr != nil {
		return nil, fmt.Errorf("failed adding middleware: %w", mwErr)
	}

	ms := &MessageService{
		router:     router,
		publisher:  pubSub,
		subscriber: pubSub,
	}

	return ms, nil
}

func (ms *MessageService) Start(ctx context.Context) error {
	log.Debug().Msg("starting message service")
	return ms.router.Run(ctx)
}

func (ms *MessageService) NewMessage(ctx context.Context, payload []byte) *message.Message {
	return message.NewMessageWithContext(ctx, uuid.NewString(), payload)
}

func (ms *MessageService) Publish(topic string, msgs ...*message.Message) error {
	for _, msg := range msgs {
		if mdErr := setMessageMetadata(msg); mdErr != nil {
			return mdErr
		}
	}
	return ms.publisher.Publish(topic, msgs...)
}

type MessageHandler = func(ctx context.Context, msg *message.Message) ([]*message.Message, error)
type MessageConsumerHandler = func(ctx context.Context, msg *message.Message) error

func (ms *MessageService) AddHandler(name, subTopic, pubTopic string, fn MessageHandler) {
	ms.router.AddHandler(name, subTopic, ms.subscriber, pubTopic, ms.publisher, func(m *message.Message) ([]*message.Message, error) {
		return fn(m.Context(), m)
	})
}

func (ms *MessageService) AddConsumerHandler(name, subTopic string, fn MessageConsumerHandler) {
	ms.router.AddConsumerHandler(name, subTopic, ms.subscriber, func(m *message.Message) error {
		return fn(m.Context(), m)
	})
}

func (ms *MessageService) Stop(ctx context.Context) error {
	log.Debug().Msg("stopping message service")
	if !ms.router.IsRunning() {
		return nil
	}
	return ms.router.Close()
}
