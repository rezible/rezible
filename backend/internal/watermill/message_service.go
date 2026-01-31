package watermill

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

type MessageService struct {
	logger watermill.LoggerAdapter
	router *message.Router

	cmdBus  *cqrs.CommandBus
	cmdProc *cqrs.CommandProcessor

	eventBus  *cqrs.EventBus
	eventProc *cqrs.EventProcessor
}

func NewMessageService() (*MessageService, error) {
	var ms MessageService

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: zerolog.DefaultContextLogger,
	}
	if rez.Config.DataSyncMode() {
		slogOpts.Level = slog.LevelWarn
	}

	ms.logger = watermill.NewSlogLogger(slog.New(slogOpts.NewZerologHandler()))

	cfg := message.RouterConfig{
		CloseTimeout: time.Second * 5,
	}
	var routerErr error
	ms.router, routerErr = message.NewRouter(cfg, ms.logger)
	if routerErr != nil {
		return nil, fmt.Errorf("failed initializing message router: %w", routerErr)
	}

	gcCfg := gochannel.Config{
		PreserveContext: false,
	}
	gcPubSub := gochannel.NewGoChannel(gcCfg, ms.logger)

	pubWithMetadata := message.MessageTransformPublisherDecorator(ms.setMessageMetadataPublisherTransform)
	mdPub, pubErr := pubWithMetadata(gcPubSub)
	if pubErr != nil {
		return nil, fmt.Errorf("failed to decorate publisher: %w", pubErr)
	}

	retry := middleware.Retry{
		MaxRetries:      1,
		InitialInterval: time.Second,
		Logger:          ms.logger,
	}

	poison, poisonErr := ms.setupPoisonQueue(mdPub, gcPubSub)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed to setup poison queue: %w", poisonErr)
	}

	throttle := middleware.NewThrottle(10, time.Second)

	ms.router.AddMiddleware(
		middleware.CorrelationID,
		throttle.Middleware,
		setMessageAccessContextMiddleware,
		// send errors to different queue
		poison,
		// catch errors & retry up to 1 time, then bubble up
		retry.Middleware,
		// catch panics and return as error
		middleware.Recoverer,
	)

	if setupErr := ms.setup(mdPub, gcPubSub); setupErr != nil {
		return nil, fmt.Errorf("setup: %w", setupErr)
	}

	return &ms, nil
}

func (ms *MessageService) Start(ctx context.Context) error {
	log.Debug().Msg("starting message service")
	return ms.router.Run(ctx)
}

func (ms *MessageService) handlePoisonQueueMessageAdded(msg *message.Message) error {
	log.Error().Msg("message sent to poison queue")
	return nil
}

func (ms *MessageService) Stop(ctx context.Context) error {
	log.Debug().Msg("stopping message service")
	if !ms.router.IsRunning() {
		return nil
	}
	return ms.router.Close()
}

func (ms *MessageService) setup(pub message.Publisher, sub message.Subscriber) error {
	cmdsPub := pub
	cmdsSub := sub

	eventsPub := pub
	eventsSub := sub

	jsonMarshaller := cqrs.JSONMarshaler{GenerateName: cqrs.FullyQualifiedStructName}

	generateEventsTopic := func(eventName string) string {
		return "events." + eventName
	}

	generateCommandsTopic := func(commandName string) string {
		return "commands." + commandName
	}

	cmdBusCfg := cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return generateCommandsTopic(params.CommandName), nil
		},
		OnSend: func(params cqrs.CommandBusOnSendParams) error {
			log.Debug().Str("name", params.CommandName).Msg("sending command")
			return nil
		},
		Marshaler: jsonMarshaller,
		Logger:    ms.logger,
	}

	cmdProcessorCfg := cqrs.CommandProcessorConfig{
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return cmdsSub, nil
		},
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateCommandsTopic(params.CommandName), nil
		},
		Marshaler: jsonMarshaller,
		Logger:    ms.logger,
	}

	eventBusCfg := cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return generateEventsTopic(params.EventName), nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			log.Debug().Str("name", params.EventName).Msg("sending event")
			return nil
		},
		Marshaler: jsonMarshaller,
		Logger:    ms.logger,
	}

	eventProcCfg := cqrs.EventProcessorConfig{
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return eventsSub, nil
		},
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateEventsTopic(params.EventName), nil
		},
		Marshaler: jsonMarshaller,
		Logger:    ms.logger,
	}

	cmdBus, cmdBusErr := cqrs.NewCommandBusWithConfig(cmdsPub, cmdBusCfg)
	if cmdBusErr != nil {
		return fmt.Errorf("failed creating command bus: %w", cmdBusErr)
	}
	ms.cmdBus = cmdBus

	cmdProc, cmdProcErr := cqrs.NewCommandProcessorWithConfig(ms.router, cmdProcessorCfg)
	if cmdProcErr != nil {
		return fmt.Errorf("failed creating command processor: %w", cmdProcErr)
	}
	ms.cmdProc = cmdProc

	eventBus, eventBusErr := cqrs.NewEventBusWithConfig(eventsPub, eventBusCfg)
	if eventBusErr != nil {
		return fmt.Errorf("failed creating event bus: %w", eventBusErr)
	}
	ms.eventBus = eventBus

	eventProc, eventProcErr := cqrs.NewEventProcessorWithConfig(ms.router, eventProcCfg)
	if eventProcErr != nil {
		return fmt.Errorf("failed creating event processor: %w", eventProcErr)
	}
	ms.eventProc = eventProc

	return nil
}

func (ms *MessageService) SendCommand(ctx context.Context, cmd any) error {
	return ms.cmdBus.Send(ctx, cmd)
}

func (ms *MessageService) AddCommandHandlers(handlers ...cqrs.CommandHandler) error {
	return ms.cmdProc.AddHandlers(handlers...)
}

func (ms *MessageService) PublishEvent(ctx context.Context, ev any) error {
	return ms.eventBus.Publish(ctx, ev)
}

func (ms *MessageService) AddEventHandlers(handlers ...cqrs.EventHandler) error {
	return ms.eventProc.AddHandlers(handlers...)
}
