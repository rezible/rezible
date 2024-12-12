package documents

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ChatService struct {
	loader   rez.ProviderLoader
	provider rez.ChatProvider
	users    rez.UserService
}

func NewChatService(ctx context.Context, pl rez.ProviderLoader, users rez.UserService) (*ChatService, error) {
	s := &ChatService{
		loader: pl,
		users:  users,
	}

	if provErr := s.LoadDataProvider(ctx); provErr != nil {
		return nil, provErr
	}

	return s, nil
}

func (s *ChatService) LoadDataProvider(ctx context.Context) error {
	prov, provErr := s.loader.LoadChatProvider(ctx)
	if provErr != nil {
		return provErr
	}
	s.provider = prov
	s.provider.SetUserLookupFunc(s.users.GetByChatId)
	return nil
}

func (s *ChatService) SendUserMessage(ctx context.Context, user *ent.User, msgText string) error {
	return s.provider.SendUserMessage(ctx, user, msgText)
}

func (s *ChatService) SendMessage(ctx context.Context, user *ent.User, msg *rez.DocumentNode) error {
	//TODO implement me
	panic("implement me")
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	return s.provider.SendOncallHandover(ctx, params)
}

func (s *ChatService) SendUserLinkMessage(ctx context.Context, user *ent.User, msgText string, linkUrl string, linkText string) error {
	return s.provider.SendUserLinkMessage(ctx, user, msgText, linkUrl, linkText)
}
