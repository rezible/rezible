package watermill

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/rezible/rezible/test"
)

type MessageQueueSuite struct {
	test.Suite
}

func TestMessageQueueSuite(t *testing.T) {
	suite.Run(t, &MessageQueueSuite{Suite: test.NewSuite()})
}

func (s *MessageQueueSuite) TestMessageExecutionContextRoundTripsTenantActor() {
	ctx := execution.NewTenantContext(s.T().Context(), 42)
	expected := execution.GetContext(ctx)
	msg := message.NewMessage(uuid.NewString(), []byte("event"))
	msg.SetContext(ctx)
	s.Require().NoError(setMessageExecutionContext(msg))

	msg.SetContext(s.T().Context())
	s.Require().NoError((&MessageQueue{}).restoreMessageContext(msg))

	s.Equal(expected, execution.GetContext(msg.Context()))
}

type failingTestEvent struct{}

func (failingTestEvent) MessageName() string { return "test.failing.v1" }

func (s *MessageQueueSuite) TestFailingHandlerStopsAfterItsRetries() {
	var calls atomic.Int64
	handler := messages.NewEventHandler("test.failing", func(context.Context, *failingTestEvent) error {
		calls.Add(1)
		return errors.New("always fails")
	})
	queue := s.runQueue(s.SetupTestPool(), handler)

	publishCtx := execution.NewTenantContext(s.T().Context(), 1)
	s.Require().NoError(queue.Publish(publishCtx, failingTestEvent{}))

	// The first attempt and two retries, a second apart; then it goes to the poison queue and is not redelivered.
	s.Eventually(func() bool { return calls.Load() >= 3 }, 10*time.Second, 50*time.Millisecond)
	time.Sleep(2 * time.Second)
	s.Equal(int64(3), calls.Load())
}

type outboxTestEvent struct {
	Name string `json:"name"`
}

func (outboxTestEvent) MessageName() string { return "test.outbox.v1" }

func (s *MessageQueueSuite) TestPublishDeliveredOnlyOnCommit() {
	pool := s.SetupTestPool()
	db, dbErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(dbErr)

	type delivery struct {
		name     string
		tenantID int
	}
	received := make(chan delivery, 2)
	handler := messages.NewEventHandler("test.outbox", func(ctx context.Context, ev *outboxTestEvent) error {
		tenantID, _ := execution.GetContext(ctx).TenantID()
		received <- delivery{name: ev.Name, tenantID: tenantID}
		return nil
	})
	queue := s.runQueue(pool, handler)

	ctx := execution.NewTenantContext(s.T().Context(), 7)
	errRollback := errors.New("roll back")
	rollbackErr := db.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		s.Require().NoError(queue.Publish(txCtx, outboxTestEvent{Name: "rolled back"}))
		return errRollback
	})
	s.Require().ErrorIs(rollbackErr, errRollback)
	commitErr := db.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		if publishErr := queue.Publish(txCtx, outboxTestEvent{Name: "committed"}); publishErr != nil {
			return publishErr
		}
		// Several outbox polls pass with the transaction open; neither event may be delivered yet.
		s.Never(func() bool { return len(received) > 0 }, time.Second, 50*time.Millisecond)
		return nil
	})
	s.Require().NoError(commitErr)

	select {
	case got := <-received:
		s.Equal(delivery{name: "committed", tenantID: 7}, got)
	case <-time.After(10 * time.Second):
		s.FailNow("committed event was not delivered")
	}
}

// runQueue runs an in-process message queue, with its outbox on the pool, until the test ends.
func (s *MessageQueueSuite) runQueue(pool *postgres.ConnectionPool, handlers ...messages.EventHandler) *MessageQueue {
	queue, queueErr := NewMessageQueue(s.Config().MessageQueue, NewGoChannelTransport(), pool.Pool)
	s.Require().NoError(queueErr)
	s.Require().NoError(queue.Register(messages.Definition{Handlers: handlers}))

	ctx, cancel := context.WithCancel(s.T().Context())
	ready := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- queue.Run(ctx, ready)
	}()
	s.T().Cleanup(func() {
		cancel()
		s.NoError(<-done, "stop message queue")
		s.NoError(queue.Shutdown(context.WithoutCancel(s.T().Context())), "shut down message queue")
	})
	<-ready
	return queue
}
