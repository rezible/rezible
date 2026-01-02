package watermill

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
)

const (
	messageMetadataKeyAccessContext = "ac"
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

func (ms *MessageService) setMessageMetadataPublisherTransform(msg *message.Message) {
	ac := access.GetContext(msg.Context())
	acs, jsonErr := json.Marshal(ac)
	if jsonErr != nil {
		log.Error().Err(jsonErr).Msg("failed to marshal access context")
		return
	}
	msg.Metadata.Set(messageMetadataKeyAccessContext, string(acs))
}

func setMessageAccessContextMiddleware(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		mdAc := msg.Metadata.Get(messageMetadataKeyAccessContext)
		if len(mdAc) > 0 {
			var ac access.Context
			if jsonErr := json.Unmarshal([]byte(mdAc), &ac); jsonErr != nil {
				return nil, fmt.Errorf("unmarshalling access context: %w", jsonErr)
			}
			msg.SetContext(access.SetContext(msg.Context(), ac))
		}
		return fn(msg)
	}
}
