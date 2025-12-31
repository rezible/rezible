package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

type ChatService struct {
	jobs         rez.JobsService
	messages     rez.MessageService
	integrations rez.IntegrationsService
	users        rez.UserService
	incidents    rez.IncidentService
	annos        rez.EventAnnotationsService
	components   rez.SystemComponentsService

	oauthConfig *oauth2.Config
}

func NewChatService(jobSvc rez.JobsService, messages rez.MessageService, integrations rez.IntegrationsService, users rez.UserService, incidents rez.IncidentService, annos rez.EventAnnotationsService, components rez.SystemComponentsService) (*ChatService, error) {
	s := &ChatService{
		jobs:         jobSvc,
		messages:     messages,
		integrations: integrations,
		users:        users,
		incidents:    incidents,
		annos:        annos,
		components:   components,
		oauthConfig:  LoadOAuthConfig(),
	}

	integrations.RegisterOAuth2Handler(integrationName, s)
	s.setupIncidentUpdateHandler()

	return s, nil
}

func (s *ChatService) EnableEventListener() bool {
	return UseSocketMode()
}

func (s *ChatService) withClient(ctx context.Context, fn func(*slack.Client) error) error {
	return withClient(ctx, s.integrations, fn)
}

func (s *ChatService) getChatTeamContext(ctx context.Context, teamId string) (context.Context, error) {
	return nil, errors.New("not implemented")
}

func (s *ChatService) getChatUserContext(ctx context.Context, userId string) (context.Context, error) {
	_, usrCtx, usrErr := s.lookupChatUser(ctx, userId)
	return usrCtx, usrErr
}

func (s *ChatService) lookupChatUser(baseCtx context.Context, chatId string) (*ent.User, context.Context, error) {
	usr, usrErr := s.users.GetByChatId(access.SystemContext(baseCtx), chatId)
	if usrErr != nil {
		log.Error().Err(usrErr).Str("chat_id", chatId).Msg("failed to lookup chat user")
		return nil, nil, usrErr
	}
	return usr, access.TenantContext(baseCtx, usr.TenantID), nil
}

func (s *ChatService) sendMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) error {
	return s.withClient(ctx, func(client *slack.Client) error {
		_, _, msgErr := client.PostMessageContext(ctx, channelId, msgOpts...)
		return msgErr
	})
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

func (s *ChatService) OAuth2Config() *oauth2.Config {
	return s.oauthConfig
}

func (s *ChatService) GetIntegrationFromToken(token *oauth2.Token) (*ent.Integration, error) {
	cfg, cfgErr := getIntegrationConfigFromOAuthToken(token)
	if cfgErr != nil {
		return nil, fmt.Errorf("get integration config: %w", cfgErr)
	}

	cfgJson, jsonErr := json.Marshal(cfg)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshalling provider config: %w", jsonErr)
	}

	return &ent.Integration{
		Name:   integrationName,
		Config: cfgJson,
	}, nil
}
