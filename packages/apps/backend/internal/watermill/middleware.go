package watermill

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/rezible/rezible/access"
)

const (
	messageMetadataKeyAccessScope = "ac"
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
	encodedScope, scopeErr := access.EncodeScope(msg.Context())
	if scopeErr != nil {
		ms.logger.Error("failed to marshal access scope", scopeErr, nil)
		return
	}
	msg.Metadata.Set(messageMetadataKeyAccessScope, string(encodedScope))
}

func (ms *MessageService) restoreMessageAccessScope(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		mdAc := msg.Metadata.Get(messageMetadataKeyAccessScope)
		restoredCtx, restoreErr := access.RestoreScope(msg.Context(), []byte(mdAc))
		if restoreErr != nil {
			return nil, fmt.Errorf("restoring access scope: %w", restoreErr)
		}
		msg.SetContext(restoredCtx)
		return fn(msg)
	}
}
