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
	"golang.org/x/sync/errgroup"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	//"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
)

type MessageQueue struct {
	cfg    rez.MessageQueueConfig
	logger watermill.LoggerAdapter

	transport messages.Transport
	router    *message.Router

	bus       *cqrs.EventBus
	liveBus   *cqrs.EventBus
	processor *cqrs.EventProcessor
}

func NewMessageQueue(cfg rez.MessageQueueConfig, ts rez.TelemetryService, transport messages.Transport) (*MessageQueue, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{
		Name:  "watermill",
		Level: slog.LevelWarn,
	})
	ms := &MessageQueue{
		cfg:       cfg,
		transport: transport,
		logger:    watermill.NewSlogLogger(logger),
	}

	rcfg := message.RouterConfig{
		CloseTimeout: time.Second * 5,
	}
	router, routerErr := message.NewRouter(rcfg, ms.logger)
	if routerErr != nil {
		return nil, fmt.Errorf("create router: %w", routerErr)
	}
	router.AddMiddleware(
		ms.makeRetryMiddleware(),
		ms.makeThrottleMiddleware(),
		ms.attemptTimeout,
		ms.restoreMessageAccessScope,
		wotelfloss.ExtractRemoteParentSpanContext(),
		wotel.Trace(),
		middleware.Recoverer,
	)
	ms.router = router

	proc, procErr := ms.makeEventProcessor()
	if procErr != nil {
		return nil, fmt.Errorf("event processor: %w", procErr)
	}
	ms.processor = proc

	pubWrapper := newPublisherWrapper(ts)

	//pubCfg := forwarder.PublisherConfig{
	//	ForwarderTopic: "message_outbox",
	//}
	//forwardedPublisher := forwarder.NewPublisher(deps.OutboxPublisher, pubCfg)

	pub, pubErr := pubWrapper.wrap(transport.Publisher())
	if pubErr != nil {
		return nil, fmt.Errorf("decorate publisher: %w", pubErr)
	}
	bus, busErr := ms.makeEventBus(pub)
	if busErr != nil {
		return nil, fmt.Errorf("event bus: %w", busErr)
	}
	ms.bus = bus

	//livePub, livePubErr := ms.wrapPublisher(basePublisher)
	//if livePubErr != nil {
	//	return nil, fmt.Errorf("decorate live publisher: %w", livePubErr)
	//}
	//liveBus, liveBusErr := ms.makeEventBus(livePub)
	//if liveBusErr != nil {
	//	return nil, fmt.Errorf("live event bus: %w", liveBusErr)
	//}
	ms.liveBus = bus

	//fwdConfig := forwarder.Config{
	//	ForwarderTopic:      "message_outbox",
	//	AckWhenCannotUnwrap: false,
	//}
	//fwd, fwdErr := forwarder.NewForwarder(deps.OutboxSubscriber, deps.Transport.Publisher(), ms.logger, fwdConfig)
	//if fwdErr != nil {
	//	return nil, fmt.Errorf("message outbox forwarder: %w", fwdErr)
	//}
	//ms.forwarder = fwd

	return ms, nil
}

func (ms *MessageQueue) Register(def messages.Definition) error {
	for _, handler := range def.Handlers {
		if _, addErr := ms.processor.AddHandler(handler); addErr != nil {
			return fmt.Errorf("add message handler: %w", addErr)
		}
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
	return errors.Join(ms.router.Close(), ms.transport.Close())
}

func (ms *MessageQueue) eventTopic(eventName string) string {
	return ms.cfg.Namespace + ":events:" + eventName
}

func (ms *MessageQueue) verifyQueueReadyState() error {
	// TODO: check message queue state
	// rez.ErrMessageQueueNotRunning
	return nil
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
	if stateErr := ms.verifyQueueReadyState(); stateErr != nil {
		return fmt.Errorf("queue state: %w", stateErr)
	}
	return bus.Publish(ctx, ev)
}

func (ms *MessageQueue) Subscribe(ctx context.Context, opts *rez.MessageEventSubscriptionOpts, handlers ...rez.MessageEventHandler) error {
	if len(handlers) == 0 {
		return fmt.Errorf("subscribe: at least one event handler is required")
	}

	adhocSub, ok := ms.transport.(messages.TransportWithAdhocSubscriber)
	if !ok {
		return fmt.Errorf("transport: adhoc subscribe not supported")
	}

	var scopes rez.MessageEventScopes
	if opts != nil {
		scopes = opts.Scopes
	}

	group, subCtx := errgroup.WithContext(ctx)
	for _, handler := range handlers {
		group.Go(func() error {
			return ms.subscribe(subCtx, adhocSub, handler, scopes)
		})
	}
	return group.Wait()
}

func (ms *MessageQueue) subscribe(ctx context.Context, t messages.TransportWithAdhocSubscriber, handler rez.MessageEventHandler, scopes rez.MessageEventScopes) error {
	eventName := eventMarshaller.Name(handler.NewEvent())
	topic := ms.eventTopic(eventName)

	subscriberName := handler.HandlerName() + "-" + uuid.NewString()
	sub, subErr := t.AdhocSubscriber(subscriberName)
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
