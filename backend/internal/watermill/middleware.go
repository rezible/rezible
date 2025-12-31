package watermill

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
)

func addMiddleware(router *message.Router, logger watermill.LoggerAdapter, pub message.Publisher, sub message.Subscriber) error {
	retry := middleware.Retry{
		MaxRetries:      1,
		InitialInterval: time.Second,
		Logger:          logger,
	}

	poison, poisonErr := makePoisonQueueMiddleware(pub)
	if poisonErr != nil {
		return fmt.Errorf("failed initializing poison queue: %w", poisonErr)
	}

	throttle := middleware.NewThrottle(10, time.Second)

	router.AddConsumerHandler("PoisonQueueLogger", "poison", sub, func(m *message.Message) error {
		log.Error().Msg("message sent to poison queue")
		return nil
	})

	router.AddMiddleware(
		throttle.Middleware,
		middleware.CorrelationID,
		setMessageAccessContextMiddleware,
		// send errors to different queue
		poison,
		// catch errors & retry up to 1 time, then bubble up
		retry.Middleware,
		// catch panics and return as error
		middleware.Recoverer,
	)
	return nil

}

func setMessageMetadata(msg *message.Message) error {
	ac := access.GetContext(msg.Context())
	acs, jsonErr := json.Marshal(ac)
	if jsonErr != nil {
		return fmt.Errorf("marshalling access context: %w", jsonErr)
	}
	msg.Metadata.Set("ac", string(acs))
	return nil
}

func setMessageAccessContextMiddleware(fn message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		var ac access.Context
		if jsonErr := json.Unmarshal([]byte(msg.Metadata.Get("ac")), &ac); jsonErr != nil {
			return nil, fmt.Errorf("unmarshalling access context: %w", jsonErr)
		}
		msg.SetContext(access.SetContext(msg.Context(), ac))
		return fn(msg)
	}
}

func makePoisonQueueMiddleware(pub message.Publisher) (message.HandlerMiddleware, error) {
	poisonFilter := func(err error) bool {
		return true
	}
	poison, poisonErr := middleware.PoisonQueueWithFilter(pub, "poison", poisonFilter)
	if poisonErr != nil {
		return nil, fmt.Errorf("failed initializing poison queue: %w", poisonErr)
	}
	return poison, nil
}
