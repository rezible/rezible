package slack

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

type ChatService struct {
	oauthConfig *oauth2.Config

	jobs         rez.JobsService
	integrations rez.IntegrationsService
	users        rez.UserService
	incidents    rez.IncidentService
	annos        rez.EventAnnotationsService
	components   rez.SystemComponentsService
}

const integrationName = "slack"

func NewChatService(jobs rez.JobsService, integrations rez.IntegrationsService, users rez.UserService, incidents rez.IncidentService, annos rez.EventAnnotationsService, components rez.SystemComponentsService) (*ChatService, error) {
	s := &ChatService{
		oauthConfig:  LoadOAuthConfig(),
		integrations: integrations,
		jobs:         jobs,
		users:        users,
		incidents:    incidents,
		annos:        annos,
		components:   components,
	}
	integrations.RegisterOAuth2Handler(integrationName, s)
	return s, nil
}

func (s *ChatService) loadIntegrationConfig(ctx context.Context) (*IntegrationConfigData, error) {
	params := rez.ListIntegrationsParams{
		Name: integrationName,
	}
	results, listErr := s.integrations.ListIntegrations(ctx, params)
	if listErr != nil {
		return nil, listErr
	}
	// TODO: handle multiple??
	if len(results) != 1 {
		return nil, fmt.Errorf("expected 1 integration, got %d", len(results))
	}
	var cfg IntegrationConfigData
	if jsonErr := json.Unmarshal(results[0].Config, &cfg); jsonErr != nil {
		return nil, jsonErr
	}
	return &cfg, nil
}

func (s *ChatService) getClient(ctx context.Context) (*slack.Client, error) {
	if rez.Config.SingleTenantMode() {
		return LoadSingleTenantClient()
	}
	cfg, loadErr := s.loadIntegrationConfig(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("loading integration config: %w", loadErr)
	}
	return slack.New(cfg.AccessToken), nil
}

func (s *ChatService) withClient(ctx context.Context, fn func(*slack.Client) error) error {
	client, clientErr := s.getClient(ctx)
	if clientErr != nil {
		return fmt.Errorf("failed to get slack client: %w", clientErr)
	}
	return fn(client)
}

func (s *ChatService) EnableEventListener() bool {
	return UseSocketMode()
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
	return usr, access.TenantUserContext(baseCtx, usr.TenantID), nil
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

type teamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IntegrationConfigData struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	BotUserID   string    `json:"bot_user_id"`
	Team        *teamInfo `json:"team"`
	Enterprise  *teamInfo `json:"enterprise"`
}

func (s *ChatService) CompleteOAuth2Flow(ctx context.Context, code string) (*ent.Integration, error) {
	token, tokenErr := s.oauthConfig.Exchange(ctx, code)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}

	team := token.Extra("team")
	log.Debug().Interface("team", team).Msg("complete oauth2 flow")

	cfg := IntegrationConfigData{
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		Scope:       token.Extra("scope").(string),
		BotUserID:   token.Extra("bot_user_id").(string),
	}
	cfgJson, jsonErr := json.Marshal(cfg)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshalling provider config: %w", jsonErr)
	}

	return &ent.Integration{
		Name:    integrationName,
		Enabled: true,
		Config:  cfgJson,
	}, nil
}
