package postgres

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ChatService struct {
	db   *ent.Client
	prov rez.ChatProvider
}

func NewChatService(db *ent.Client, prov rez.ChatProvider) (*ChatService, error) {
	p := &ChatService{
		db:   db,
		prov: prov,
	}

	return p, nil
}

func (s *ChatService) Provider() rez.ChatProvider {
	return s.prov
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	return s.prov.SendOncallHandover(ctx, params)
}
