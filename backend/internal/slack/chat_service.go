package slack

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type ChatService struct {
	jobs  rez.JobsService
	users rez.UserService
	annos rez.EventAnnotationsService

	client *slack.Client
}

func NewChatService(jobs rez.JobsService, users rez.UserService, annos rez.EventAnnotationsService) (*ChatService, error) {
	client, clientErr := LoadClient()
	if clientErr != nil {
		return nil, clientErr
	}
	s := &ChatService{
		jobs:   jobs,
		users:  users,
		annos:  annos,
		client: client,
	}
	return s, nil
}

func (s *ChatService) EnableEventListener() bool {
	return UseSocketMode()
}

func (s *ChatService) lookupChatUser(baseCtx context.Context, chatId string) (*ent.User, context.Context, error) {
	usr, usrErr := s.users.GetByChatId(access.SystemContext(baseCtx), chatId)
	if usrErr != nil {
		log.Error().Err(usrErr).Str("chat_id", chatId).Msg("failed to lookup chat user")
		return nil, nil, usrErr
	}
	return usr, access.TenantUserContext(baseCtx, usr.TenantID), nil
}

func (s *ChatService) sendMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) error {
	_, _, msgErr := s.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgErr
}

func (s *ChatService) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (s *ChatService) SendTextMessage(ctx context.Context, channelId string, text string) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (s *ChatService) SendReply(ctx context.Context, channelId string, threadId string, text string) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	channel, msg, err := buildHandoverMessage(params)
	if err != nil {
		return fmt.Errorf("creating handover message: %w", err)
	}
	return s.sendMessage(ctx, channel, msg)
}

func (s *ChatService) SendOncallHandoverReminder(ctx context.Context, shift *ent.OncallShift) error {
	return nil
}
