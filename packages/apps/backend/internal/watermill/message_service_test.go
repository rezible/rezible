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
	queue := s.runQueue(handler)

	publishCtx := execution.NewTenantContext(s.T().Context(), 1)
	s.Require().NoError(queue.Publish(publishCtx, failingTestEvent{}))

	// The first attempt and two retries, a second apart; then it goes to the poison queue and is not redelivered.
	s.Eventually(func() bool { return calls.Load() >= 3 }, 10*time.Second, 50*time.Millisecond)
	time.Sleep(2 * time.Second)
	s.Equal(int64(3), calls.Load())
}

// runQueue runs an in-process message queue with the handlers until the test ends.
func (s *MessageQueueSuite) runQueue(handlers ...messages.EventHandler) *MessageQueue {
	queue, queueErr := NewMessageQueue(s.Config().MessageQueue, NewGoChannelTransport())
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
	})
	<-ready
	return queue
}
