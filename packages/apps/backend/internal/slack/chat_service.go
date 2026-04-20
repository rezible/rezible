package slack

import (
	"context"
	"fmt"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/user"
	"github.com/slack-go/slack"
)

type ChatService struct {
	ci     *ConfiguredIntegration
	client *slack.Client
	logger *slog.Logger

	integrations rez.IntegrationsService
	users        rez.UserService
	incidents    rez.IncidentService
	annos        rez.EventAnnotationsService
}

func newChatService(ci *ConfiguredIntegration) *ChatService {
	return &ChatService{
		ci:           ci,
		client:       slack.New(ci.accessToken()),
		logger:       slog.Default().With("package", "slack"),
		users:        ci.svcs.Users,
		integrations: ci.svcs.Integrations,
		incidents:    ci.svcs.Incidents,
		annos:        ci.svcs.EventAnnotations,
	}
}

func (s *ChatService) createUserContext(ctx context.Context, userChatId string) (context.Context, error) {
	ctx = access.TenantContext(ctx, s.ci.intg.TenantID)
	usr, usrErr := s.users.Get(ctx, user.ChatID(userChatId))
	if usrErr != nil {
		s.logger.Error("failed to lookup chat user", "error", usrErr, "chat_id", userChatId)
		return nil, fmt.Errorf("lookup user: %w", usrErr)
	}
	return access.WithUser(ctx, usr), nil
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
