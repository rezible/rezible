package postgres

import (
	"context"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ChatService struct {
	db   *ent.Client
	jobs rez.JobsService
	prov rez.ChatProvider
}

func NewChatService(db *ent.Client, jobs rez.JobsService, prov rez.ChatProvider) (*ChatService, error) {
	s := &ChatService{
		db:   db,
		jobs: jobs,
		prov: prov,
	}

	return s, nil
}

func (s *ChatService) HandleMentionEvent(chatId, threadId, userId, msgText string) {
	log.Debug().Str("userId", userId).Str("text", msgText).Msg("handling mention event")

	err := s.prov.SendReply(context.Background(), chatId, threadId, "mention reply!")
	if err != nil {
		log.Error().Err(err).Msg("failed to send mention reply")
	}
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	return s.prov.SendOncallHandover(ctx, params)
}
