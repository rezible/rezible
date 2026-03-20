package slack

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type ChatService struct {
	ci     *ConfiguredIntegration
	client *slack.Client

	integrations rez.IntegrationsService
	users        rez.UserService
	incidents    rez.IncidentService
	annos        rez.EventAnnotationsService
}

func newChatService(ci *ConfiguredIntegration) *ChatService {
	return &ChatService{
		ci:           ci,
		client:       slack.New(ci.accessToken()),
		users:        ci.svcs.Users,
		integrations: ci.svcs.Integrations,
		incidents:    ci.svcs.Incidents,
		annos:        ci.svcs.EventAnnotations,
	}
}

func (s *ChatService) getUserContext(ctx context.Context, userChatId string) (context.Context, error) {
	_, userCtx, lookupErr := s.lookupUser(ctx, userChatId)
	return userCtx, lookupErr
}

func (s *ChatService) lookupUser(ctx context.Context, userChatId string) (*ent.User, context.Context, error) {
	ctx = s.ci.tenantContext(ctx)
	usr, usrErr := s.users.GetByChatId(ctx, userChatId)
	if usrErr != nil {
		log.Error().Err(usrErr).Str("chat_id", userChatId).Msg("failed to lookup chat user")
		return nil, nil, fmt.Errorf("lookup user: %w", usrErr)
	}
	return usr, s.users.CreateUserAccessContext(ctx, usr), nil
}

func (s *ChatService) postMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) (string, error) {
	_, msgTs, msgErr := s.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgTs, msgErr
}

func (s *ChatService) postEphemeralMessage(ctx context.Context, channelId, userId string, msgOpts ...slack.MsgOption) (string, error) {
	return s.client.PostEphemeralContext(ctx, channelId, userId, msgOpts...)
}

func (s *ChatService) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) (string, error) {
	blocks := convertContentToBlocks("", content)
	return s.postMessage(ctx, channelId, slack.MsgOptionBlocks(blocks...))
}

func (s *ChatService) SendTextMessage(ctx context.Context, channelId string, text string) (string, error) {
	return s.postMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (s *ChatService) SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error) {
	return s.postMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (s *ChatService) openModalView(ctx context.Context, triggerId string, viewReq slack.ModalViewRequest) error {
	resp, respErr := s.client.OpenViewContext(ctx, triggerId, viewReq)
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return respErr
	}
	return nil
}

func (s *ChatService) openOrUpdateModal(ctx context.Context, ic *slack.InteractionCallback, view *slack.ModalViewRequest) error {
	var viewResp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		viewResp, respErr = s.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		viewResp, respErr = s.client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, viewResp)
		return respErr
	}

	return nil
}
