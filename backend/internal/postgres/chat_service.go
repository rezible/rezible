package postgres

import (
	"context"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ChatService struct {
	db *ent.Client
	pl rez.ProviderLoader
}

func NewChatService(db *ent.Client, pl rez.ProviderLoader) (*ChatService, error) {
	s := &ChatService{
		db: db,
		pl: pl,
	}

	return s, nil
}

func (s *ChatService) HandleMentionEvent(chatId, threadId, userId, msgText string) {
	log.Debug().
		Str("userId", userId).
		Str("text", msgText).
		Msg("handling mention event")

	// TODO: lookup user by chat id, build user access ctx
	/*
		ctx := context.Background()
		prov, provErr := s.pl.GetChatProvider(ctx)

		err := s.prov.SendReply(context.Background(), chatId, threadId, "mention reply!")
		if err != nil {
			log.Error().Err(err).Msg("failed to send mention reply")
		}
	*/
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	prov, provErr := s.pl.GetChatProvider(ctx)
	if provErr != nil {
		return provErr
	}
	return prov.SendOncallHandover(ctx, params)
}

func (s *ChatService) SendOncallHandoverReminder(ctx context.Context, shift *ent.OncallUserShift) error {
	return nil
}
