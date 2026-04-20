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
)

type MessageService struct {
	logger     watermill.LoggerAdapter
	router     *message.Router
	marshaller cqrs.CommandEventMarshaler

	cmdBus  *cqrs.CommandBus
	cmdProc *cqrs.CommandProcessor

	eventBus  *cqrs.EventBus
	eventProc *cqrs.EventProcessor
}

func NewMessageService() (*MessageService, error) {
	ms := MessageService{
		logger:     watermill.NewSlogLogger(slog.Default().With("package", "watermill")),
		marshaller: cqrs.JSONMarshaler{GenerateName: cqrs.FullyQualifiedStructName},
	}

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

	retry := middleware.Retry{
		MaxRetries:      1,
		InitialInterval: time.Second,
		Logger:          ms.logger,
	}

	pubWithMetadata := message.MessageTransformPublisherDecorator(ms.setMessageAccessScope)
	mdPub, pubErr := pubWithMetadata(gcPubSub)
	if pubErr != nil {
		return nil, fmt.Errorf("failed to decorate publisher: %w", pubErr)
	}

	poison, poisonErr := ms.setupPoisonQueue(mdPub, gcPubSub)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed to setup poison queue: %w", poisonErr)
	}

	throttle := middleware.NewThrottle(10, time.Second)

	ms.router.AddMiddleware(
		middleware.CorrelationID,
		throttle.Middleware,
		ms.restoreMessageAccessScope,
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
	ms.logger.Debug("starting message service", nil)
	return ms.router.Run(ctx)
}

func (ms *MessageService) Stop(ctx context.Context) error {
	ms.logger.Debug("stopping message service", nil)
	if !ms.router.IsRunning() {
		return nil
	}
	return ms.router.Close()
}

func (ms *MessageService) setup(pub message.Publisher, sub message.Subscriber) error {
	if cmdsErr := ms.setupCommandProcessor(pub, sub); cmdsErr != nil {
		return fmt.Errorf("command processor: %w", cmdsErr)
	}

	if cmdsErr := ms.setupEventProcessor(pub, sub); cmdsErr != nil {
		return fmt.Errorf("event processor: %w", cmdsErr)
	}

	return nil
}

func (ms *MessageService) setupEventProcessor(pub message.Publisher, sub message.Subscriber) error {
	generateTopic := func(eventName string) string {
		return "events." + eventName
	}

	eventBusCfg := cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return generateTopic(params.EventName), nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			slog.Debug("sending event", "name", params.EventName)
			return nil
		},
		Marshaler: ms.marshaller,
		Logger:    ms.logger,
	}
	eventBus, eventBusErr := cqrs.NewEventBusWithConfig(pub, eventBusCfg)
	if eventBusErr != nil {
		return fmt.Errorf("failed creating event bus: %w", eventBusErr)
	}
	ms.eventBus = eventBus

	eventProcCfg := cqrs.EventProcessorConfig{
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return sub, nil
		},
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateTopic(params.EventName), nil
		},
		Marshaler: ms.marshaller,
		Logger:    ms.logger,
	}

	eventProc, eventProcErr := cqrs.NewEventProcessorWithConfig(ms.router, eventProcCfg)
	if eventProcErr != nil {
		return fmt.Errorf("failed creating event processor: %w", eventProcErr)
	}
	ms.eventProc = eventProc

	return nil
}

func (ms *MessageService) setupCommandProcessor(pub message.Publisher, sub message.Subscriber) error {
	generateTopic := func(eventName string) string {
		return "commands." + eventName
	}

	cmdBusCfg := cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return generateTopic(params.CommandName), nil
		},
		OnSend: func(params cqrs.CommandBusOnSendParams) error {
			ms.logger.Debug("sending command", watermill.LogFields{"name": params.CommandName})
			return nil
		},
		Marshaler: ms.marshaller,
		Logger:    ms.logger,
	}

	cmdProcessorCfg := cqrs.CommandProcessorConfig{
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return sub, nil
		},
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateTopic(params.CommandName), nil
		},
		Marshaler: ms.marshaller,
		Logger:    ms.logger,
	}

	cmdBus, cmdBusErr := cqrs.NewCommandBusWithConfig(pub, cmdBusCfg)
	if cmdBusErr != nil {
		return fmt.Errorf("failed creating command bus: %w", cmdBusErr)
	}
	ms.cmdBus = cmdBus

	cmdProc, cmdProcErr := cqrs.NewCommandProcessorWithConfig(ms.router, cmdProcessorCfg)
	if cmdProcErr != nil {
		return fmt.Errorf("failed creating command processor: %w", cmdProcErr)
	}
	ms.cmdProc = cmdProc

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
