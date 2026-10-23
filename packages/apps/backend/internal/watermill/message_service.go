package watermill

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	"github.com/google/uuid"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/sync/errgroup"

	"github.com/ThreeDotsLabs/watermill"
	wsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
)

// outboxTopic names the forwarder topic. MessageOutbox maps every topic to its one table, so it is only a name.
const outboxTopic = "outbox"

type MessageQueue struct {
	cfg    rez.MessageQueueConfig
	logger watermill.LoggerAdapter

	transport         messages.Transport
	outboxDB          wsql.Beginner
	outboxBacklog     metric.Registration
	router            *message.Router
	handlerMiddleware []message.HandlerMiddleware

	bus       *cqrs.EventBus
	liveBus   *cqrs.EventBus
	processor *cqrs.EventProcessor
}

func NewMessageQueue(cfg rez.MessageQueueConfig, transport messages.Transport, outboxConn wsql.Conn) (*MessageQueue, error) {
	// Watermill logs routine operation at Info; keep it at Debug, below its warnings and errors.
	logger := slog.Default().With("component", "watermill")
	ms := &MessageQueue{
		cfg:       cfg,
		transport: transport,
		outboxDB:  wsql.BeginnerFromPgx(outboxConn),
		logger:    watermill.NewSlogLoggerWithLevelMapping(logger, map[slog.Level]slog.Level{slog.LevelInfo: slog.LevelDebug}),
	}

	rcfg := message.RouterConfig{
		CloseTimeout: time.Second * 5,
	}
	router, routerErr := message.NewRouter(rcfg, ms.logger)
	if routerErr != nil {
		return nil, fmt.Errorf("create router: %w", routerErr)
	}
	ms.router = router

	handlerMiddleware, middlewareErr := ms.makeHandlerMiddleware()
	if middlewareErr != nil {
		return nil, fmt.Errorf("handler middleware: %w", middlewareErr)
	}
	ms.handlerMiddleware = handlerMiddleware

	proc, procErr := ms.makeEventProcessor()
	if procErr != nil {
		return nil, fmt.Errorf("event processor: %w", procErr)
	}
	ms.processor = proc

	transportPub, transportPubErr := ms.makeCountedTransportPublisher()
	if transportPubErr != nil {
		return nil, fmt.Errorf("transport publisher: %w", transportPubErr)
	}

	if fwdErr := ms.makeForwarder(transportPub); fwdErr != nil {
		return nil, fmt.Errorf("message outbox forwarder: %w", fwdErr)
	}

	bus, busErr := ms.makeOutboxEventBus()
	if busErr != nil {
		return nil, fmt.Errorf("event bus: %w", busErr)
	}
	ms.bus = bus

	liveBus, liveBusErr := ms.makeLiveEventBus(transportPub)
	if liveBusErr != nil {
		return nil, fmt.Errorf("live event bus: %w", liveBusErr)
	}
	ms.liveBus = liveBus

	backlog, backlogErr := ms.registerOutboxBacklogGauge()
	if backlogErr != nil {
		return nil, fmt.Errorf("outbox backlog gauge: %w", backlogErr)
	}
	ms.outboxBacklog = backlog

	return ms, nil
}

func (ms *MessageQueue) Register(def messages.Definition) error {
	for _, handler := range def.Handlers {
		routerHandler, addErr := ms.processor.AddHandler(handler)
		if addErr != nil {
			return fmt.Errorf("add message handler: %w", addErr)
		}
		routerHandler.AddMiddleware(ms.handlerMiddleware...)
	}
	return nil
}

func (ms *MessageQueue) Run(ctx context.Context, ready chan<- struct{}) (runErr error) {
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

func (ms *MessageQueue) Shutdown(ctx context.Context) error {
	return errors.Join(ms.router.Close(), ms.transport.Close(), ms.outboxBacklog.Unregister())
}

// makeHandlerMiddleware builds the middleware every event handler shares. The instances are shared, so the
// throttle limits all handlers together.
func (ms *MessageQueue) makeHandlerMiddleware() ([]message.HandlerMiddleware, error) {
	poison, poisonErr := ms.makePoisonQueue()
	if poisonErr != nil {
		return nil, fmt.Errorf("poison queue: %w", poisonErr)
	}
	return []message.HandlerMiddleware{
		poison,
		ms.makeRetryMiddleware(),
		ms.makeThrottleMiddleware(),
		ms.attemptTimeout,
		ms.restoreMessageAccessScope,
		wotelfloss.ExtractRemoteParentSpanContext(),
		wotel.Trace(),
		middleware.Recoverer,
	}, nil
}

// makeForwarder adds a handler to the router that moves committed events from the outbox to the transport.
func (ms *MessageQueue) makeForwarder(transportPub message.Publisher) error {
	outboxSub, outboxSubErr := ms.makeOutboxSubscriber()
	if outboxSubErr != nil {
		return fmt.Errorf("outbox subscriber: %w", outboxSubErr)
	}
	cfg := forwarder.Config{
		Router:              ms.router,
		ForwarderTopic:      outboxTopic,
		AckWhenCannotUnwrap: false,
		Middlewares:         []message.HandlerMiddleware{ms.awaitRouterRunning},
	}
	_, fwdErr := forwarder.NewForwarder(outboxSub, transportPub, ms.logger, cfg)
	return fwdErr
}

// awaitRouterRunning holds a forwarded event until every handler has subscribed. The router starts handlers in
// no fixed order, and the transport drops an event with no subscribers, so the outbox row would be deleted unhandled.
func (ms *MessageQueue) awaitRouterRunning(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		select {
		case <-ms.router.Running():
			return next(msg)
		case <-msg.Context().Done():
			return nil, msg.Context().Err()
		}
	}
}

// makePoisonQueue sends a message whose handler still fails after its retries to the poison topic, where it
// is logged and counted. The message is then acked: otherwise the transport would redeliver it forever.
func (ms *MessageQueue) makePoisonQueue() (message.HandlerMiddleware, error) {
	topic := ms.cfg.Namespace + ":poison"
	poison, poisonErr := middleware.PoisonQueue(ms.transport.Publisher(), topic)
	if poisonErr != nil {
		return nil, fmt.Errorf("poison queue middleware: %w", poisonErr)
	}

	poisonSub, poisonSubErr := ms.transport.Subscriber("poison")
	if poisonSubErr != nil {
		return nil, fmt.Errorf("poison subscriber: %w", poisonSubErr)
	}
	poisoned, poisonedErr := otel.Meter("github.com/rezible/rezible").Int64Counter("backend.messages.poisoned",
		metric.WithDescription("Watermill messages whose handler failed after its retries"))
	if poisonedErr != nil {
		return nil, fmt.Errorf("messages poisoned counter: %w", poisonedErr)
	}
	ms.router.AddConsumerHandler("PoisonQueueLogger", topic, poisonSub, func(msg *message.Message) error {
		handler := msg.Metadata.Get(middleware.PoisonedHandlerKey)
		poisoned.Add(msg.Context(), 1, metric.WithAttributes(attribute.String("handler", handler)))
		slog.ErrorContext(msg.Context(), "message handler failed after its retries",
			"handler", handler,
			"topic", msg.Metadata.Get(middleware.PoisonedTopicKey),
			"error", msg.Metadata.Get(middleware.ReasonForPoisonedKey),
			"message_uuid", msg.UUID,
		)
		return nil
	})
	return poison, nil
}

func (ms *MessageQueue) eventTopic(eventName string) string {
	return ms.cfg.Namespace + ":events:" + eventName
}

func (ms *MessageQueue) Publish(ctx context.Context, ev any) error {
	return ms.publish(ctx, ms.bus, ev)
}

func (ms *MessageQueue) PublishLive(ctx context.Context, ev any) error {
	if entTx := ent.TxFromContext(ctx); entTx != nil {
		return fmt.Errorf("cannot use PublishLive inside a transaction")
	}
	return ms.publish(ctx, ms.liveBus, ev)
}

func (ms *MessageQueue) publish(ctx context.Context, bus *cqrs.EventBus, ev any) error {
	if !execution.ContextExists(ctx) {
		return fmt.Errorf("publish: execution context required")
	}
	if _, encodeErr := execution.GetContext(ctx).Encode(); encodeErr != nil {
		return fmt.Errorf("encode execution context: %w", encodeErr)
	}
	return bus.Publish(ctx, ev)
}

func (ms *MessageQueue) Subscribe(ctx context.Context, opts *rez.MessageEventSubscriptionOpts, handlers ...rez.MessageEventHandler) error {
	if len(handlers) == 0 {
		return fmt.Errorf("subscribe: at least one event handler is required")
	}

	var scopes rez.MessageEventScopes
	if opts != nil {
		scopes = opts.Scopes
	}

	group, subCtx := errgroup.WithContext(ctx)
	for _, handler := range handlers {
		group.Go(func() error {
			return ms.subscribe(subCtx, handler, scopes)
		})
	}
	return group.Wait()
}

func (ms *MessageQueue) subscribe(ctx context.Context, handler rez.MessageEventHandler, scopes rez.MessageEventScopes) error {
	eventName := eventMarshaller.Name(handler.NewEvent())
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
		if jsonErr := eventMarshaller.Unmarshal(msg, event); jsonErr != nil {
			return fmt.Errorf("unmarshal %s: %w", eventName, jsonErr)
		}

		if ctxErr := ms.restoreMessageContext(msg); ctxErr != nil {
			return fmt.Errorf("message context: %w", ctxErr)
		}

		if !ms.verifyMessageTenantId(ctx, msg) {
			return nil
		}

		if !ms.verifyMessageMatchesAnyScopes(msg, scopes) {
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

const msgMetadataScopesKey = "scopes"

func (ms *MessageQueue) setMessageEventScopesMetadata(ev any, msg *message.Message) {
	if scopedEvent, hasScopes := ev.(messages.EventWithScopes); hasScopes {
		msg.Metadata.Set(msgMetadataScopesKey, strings.Join(scopedEvent.MessageScopes(), ","))
	}
}

func (ms *MessageQueue) verifyMessageMatchesAnyScopes(msg *message.Message, scopes rez.MessageEventScopes) bool {
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

func (ms *MessageQueue) verifyMessageTenantId(ctx context.Context, msg *message.Message) bool {
	subscriberTenantID, subscriberHasTenant := execution.GetContext(ctx).TenantID()
	producerTenantID, producerHasTenant := execution.GetContext(msg.Context()).TenantID()
	if !subscriberHasTenant || !producerHasTenant {
		return false
	}
	return subscriberTenantID == producerTenantID
}
