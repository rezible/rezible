package watermill

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
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

func setMessageAccessScope(msg *message.Message) {
	encodedScope, scopeErr := access.EncodeScope(msg.Context())
	if scopeErr != nil {
		log.Error().Err(scopeErr).Msg("failed to marshal access scope")
		return
	}
	msg.Metadata.Set(messageMetadataKeyAccessScope, string(encodedScope))
}

func restoreMessageAccessScope(fn message.HandlerFunc) message.HandlerFunc {
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
