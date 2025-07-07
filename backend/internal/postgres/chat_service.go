package postgres

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
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

	prov.SetMentionHandler(s.handleMention)

	return s, nil
}

func (s *ChatService) handleMention(ev *rez.ChatMentionEvent) {
	log.Debug().Interface("event", ev).Msg("mention")

	err := s.prov.SendReply(context.Background(), ev.ChatId, ev.ThreadId, "mention reply!")
	if err != nil {
		log.Error().Err(err).Msg("failed to send mention reply")
	}
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	return s.prov.SendOncallHandover(ctx, params)
}
