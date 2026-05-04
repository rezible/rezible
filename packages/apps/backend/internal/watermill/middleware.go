package watermill

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/rezible/rezible/execution"
)

const (
	messageMetadataKeyExecutionContext = "ec"
)

func (ms *MessageService) setupPoisonQueue(pub message.Publisher, sub message.Subscriber) (message.HandlerMiddleware, error) {
	poisonFilter := func(err error) bool {
		return true
	}

	poisonQueueTopic := "poison.queue"
	poison, poisonErr := middleware.PoisonQueueWithFilter(pub, poisonQueueTopic, poisonFilter)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed initializing poison queue: %w", poisonErr)
	}
	ms.router.AddConsumerHandler("PoisonQueueLogger", poisonQueueTopic, sub, ms.handlePoisonQueueMessageAdded)

	return poison, nil
}

func (ms *MessageService) handlePoisonQueueMessageAdded(msg *message.Message) error {
	ms.logger.Info("message sent to poison queue", watermill.LogFields{"uuid": msg.UUID})
	return nil
}

func (ms *MessageService) setMessageAccessScope(msg *message.Message) {
	encodedExec, encodeErr := execution.FromContext(msg.Context()).Encode()
	if encodeErr != nil {
		ms.logger.Error("failed to marshal execution context", encodeErr, nil)
		return
	}
	msg.Metadata.Set(messageMetadataKeyExecutionContext, string(encodedExec))
}

func (ms *MessageService) restoreMessageAccessScope(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		encodedExec := msg.Metadata.Get(messageMetadataKeyExecutionContext)
		if encodedExec == "" {
			return fn(msg)
		}

		exec, decodeErr := execution.Decode([]byte(encodedExec))
		if decodeErr != nil {
			return nil, fmt.Errorf("restoring execution context: %w", decodeErr)
		}
		exec.Provenance.ParentKind = "message"
		exec.Provenance.ParentID = msg.UUID
		msg.SetContext(execution.StoreInContext(msg.Context(), exec))
		return fn(msg)
	}
}
