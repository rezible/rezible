package watermill

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	mapset "github.com/deckarep/golang-set/v2"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	"github.com/google/uuid"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
	"golang.org/x/sync/errgroup"

	rez "github.com/rezible/rezible"
)

type MessageService struct {
	telemetry rez.TelemetryService
	logger    watermill.LoggerAdapter

	transport  Transport
	publisher  message.Publisher
	router     *message.Router
	marshaller cqrs.CommandEventMarshaler

	eventBus  *cqrs.EventBus
	eventProc *cqrs.EventProcessor
}

func NewMessageService(ts rez.TelemetryService, transport Transport) (*MessageService, error) {
	ms := MessageService{
		telemetry: ts,
		transport: transport,
		logger: watermill.NewSlogLogger(ts.NewLogger(rez.NewLoggerOptions{
			Name:  "watermill",
			Level: slog.LevelWarn,
		})),
		marshaller: cqrs.JSONMarshaler{GenerateName: cqrs.FullyQualifiedStructName},
	}
	if ms.transport == nil {
		ms.transport = newGoChannelTransport(ms.logger)
	}

	pub, pubErr := ms.addPublisherDecorations(ms.transport.Publisher())
	if pubErr != nil {
		return nil, fmt.Errorf("decorate publisher: %w", pubErr)
	}
	ms.publisher = pub

	router, routerErr := message.NewRouter(message.RouterConfig{CloseTimeout: time.Second * 5}, ms.logger)
	if routerErr != nil {
		return nil, fmt.Errorf("failed initializing message router: %w", routerErr)
	}
	ms.router = router

	poison, poisonErr := ms.makePoisonQueue()
	if poisonErr != nil {
		return nil, fmt.Errorf("failed to setup poison queue: %w", poisonErr)
	}

	retry := middleware.Retry{
		MaxRetries:      1,
		InitialInterval: time.Second,
		Logger:          ms.logger,
	}

	ms.router.AddMiddleware(
		middleware.NewThrottle(10, time.Second).Middleware,
		ms.restoreMessageAccessScope,
		wotelfloss.ExtractRemoteParentSpanContext(),
		wotel.Trace(),
		poison,               // send caught errors to a dedicated queue
		retry.Middleware,     // catch errors & retry up to 1 time, then bubble up
		middleware.Recoverer, // catch panics and return as error
	)

	if eventsErr := ms.setupEventProcessor(); eventsErr != nil {
		return nil, fmt.Errorf("event processor: %w", eventsErr)
	}

	return &ms, nil
}

func (ms *MessageService) Run(ctx context.Context, ready chan<- struct{}) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-ms.router.Running():
			close(ready)
		case <-runCtx.Done():
		}
	}()
	if err := ms.router.Run(runCtx); err != nil {
		return fmt.Errorf("message router: %w", err)
	}
	return nil
}

func (ms *MessageService) Shutdown(context.Context) error {
	return errors.Join(ms.router.Close(), ms.transport.Close())
}

func (ms *MessageService) eventTopic(eventName string) string {
	return "events." + eventName
}

const msgMetadataScopesKey = "scopes"

func (ms *MessageService) setupEventProcessor() error {
	eventBusCfg := cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return ms.eventTopic(params.EventName), nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			if ev, hasScopes := params.Event.(rez.MessageEventWithScopes); hasScopes {
				params.Message.Metadata.Set(msgMetadataScopesKey, strings.Join(ev.MessageScopes(), ","))
			}
			return nil
		},
		Marshaler: ms.marshaller,
		Logger:    ms.logger,
	}
	eventBus, eventBusErr := cqrs.NewEventBusWithConfig(ms.publisher, eventBusCfg)
	if eventBusErr != nil {
		return fmt.Errorf("failed creating event bus: %w", eventBusErr)
	}
	ms.eventBus = eventBus

	eventProcCfg := cqrs.EventProcessorConfig{
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return ms.transport.Subscriber(params.HandlerName)
		},
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return ms.eventTopic(params.EventName), nil
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

func (ms *MessageService) AddHandlers(handlers ...rez.MessageEventHandler) error {
	names := mapset.NewSetFromMapKeys(ms.router.Handlers())
	for _, handler := range handlers {
		if handler == nil {
			return errors.New("cannot add nil message handler")
		}
		handlerName := handler.HandlerName()
		if handlerName == "" {
			return errors.New("cannot add message handler with empty name")
		}
		if !names.Add(handlerName) {
			return fmt.Errorf("message handler %q is duplicated", handlerName)
		}
		if _, hErr := ms.eventProc.AddHandler(handler); hErr != nil {
			return fmt.Errorf("failed adding handler: %w", hErr)
		}
	}
	return nil
}

func (ms *MessageService) Publish(ctx context.Context, ev any) error {
	return ms.eventBus.Publish(ctx, ev)
}

func (ms *MessageService) matchesScopes(msg *message.Message, scopes []string) bool {
	if len(scopes) == 0 {
		return true
	}

	if msgScopes := msg.Metadata.Get(msgMetadataScopesKey); msgScopes != "" {
		for _, want := range scopes {
			for scope := range strings.SplitSeq(msgScopes, ",") {
				if scope == want {
					return true
				}
			}
		}
	}

	return false
}

func (ms *MessageService) Subscribe(ctx context.Context, opts *rez.MessageEventSubscriptionOpts, handlers ...rez.MessageEventHandler) error {
	if len(handlers) == 0 {
		return fmt.Errorf("subscribe: at least one event handler is required")
	}

	var scopes []string
	if opts != nil {
		scopes = opts.Scopes
	}

	group, subscriptionCtx := errgroup.WithContext(ctx)
	for _, handler := range handlers {
		group.Go(func() error {
			return ms.subscribe(subscriptionCtx, handler, scopes)
		})
	}
	return group.Wait()
}

func (ms *MessageService) subscribe(ctx context.Context, handler rez.MessageEventHandler, scopes []string) error {
	eventName := ms.marshaller.Name(handler.NewEvent())
	topic := ms.eventTopic(eventName)

	subscriberName := handler.HandlerName() + "-" + uuid.NewString()
	sub, subErr := ms.transport.AdhocSubscriber(subscriberName)
	if subErr != nil {
		return fmt.Errorf("subscriber %q: %w", subscriberName, subErr)
	}
	defer func() {
		if sub != nil {
			if closeErr := sub.Close(); closeErr != nil {
				slog.Warn("failed to close adhoc subscriber", "err", closeErr)
			}
		}
	}()

	msgs, subscribeErr := sub.Subscribe(ctx, topic)
	if subscribeErr != nil {
		return fmt.Errorf("subscribe topic %q: %w", topic, subscribeErr)
	}

	handleMsg := func(msg *message.Message) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("subscriber handler panic: %v", r)
			}
		}()

		event := handler.NewEvent()
		if jsonErr := ms.marshaller.Unmarshal(msg, event); jsonErr != nil {
			return fmt.Errorf("unmarshal %s: %w", eventName, jsonErr)
		}

		if ctxErr := ms.restoreMessageContext(msg); ctxErr != nil {
			return fmt.Errorf("message context: %w", ctxErr)
		}

		if !ms.matchesScopes(msg, scopes) {
			return nil
		}

		handlerCtx, cancelHandler := context.WithCancel(msg.Context())
		stopSubscriptionCancel := context.AfterFunc(ctx, cancelHandler)
		defer func() {
			stopSubscriptionCancel()
			cancelHandler()
		}()

		return handler.Handle(handlerCtx, event)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-msgs:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("subscription topic %q closed unexpectedly", topic)
			}
			handlerErr := handleMsg(msg)

			// subscribe is live delivery - do not retry failed messages
			msg.Ack()
			if handlerErr != nil {
				return fmt.Errorf("subscribe handler: %w", handlerErr)
			}
		}
	}
}
