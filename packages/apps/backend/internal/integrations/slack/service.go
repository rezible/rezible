package slack

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	SlashCommandHandler        func(context.Context, *slack.SlashCommand) (*slack.Blocks, error)
	EventsApiHandler           func(context.Context, *slackevents.EventsAPIEvent) error
	InteractionCallbackHandler func(context.Context, *slack.InteractionCallback) error
)

type Service struct {
	integrationName string

	eventHandler *EventHandler

	intgs rez.IntegrationService
	users rez.UserService

	webhookHandler     http.Handler
	socketModeListener *SocketModeListener

	slashCommandHandlers        map[string]SlashCommandHandler
	eventsApiHandler            EventsApiHandler
	interactionCallbackHandlers map[slack.InteractionType]InteractionCallbackHandler
}

func NewService(name string, msgs rez.MessageService) (*Service, error) {
	s := &Service{
		integrationName:             name,
		webhookHandler:              http.NotFoundHandler(),
		slashCommandHandlers:        make(map[string]SlashCommandHandler),
		interactionCallbackHandlers: make(map[slack.InteractionType]InteractionCallbackHandler),
		eventsApiHandler: func(ctx context.Context, event *slackevents.EventsAPIEvent) error {
			return nil
		},
	}

	evth, evthErr := MakeEventHandler(s, msgs)
	if evthErr != nil {
		return nil, fmt.Errorf("failed to make event handler: %w", evthErr)
	}
	s.eventHandler = evth

	return s, nil
}

func (s *Service) AddSlashCommandHandler(cmd string, handler SlashCommandHandler) {
	s.slashCommandHandlers[cmd] = handler
}

func (s *Service) AddInteractionCallbackHandler(t slack.InteractionType, handler InteractionCallbackHandler) {
	s.interactionCallbackHandlers[t] = handler
}

func (s *Service) SetEventsApiHandler(handler EventsApiHandler) {
	s.eventsApiHandler = handler
}

func (s *Service) SetupWebhooks(secret string) error {
	wh, whErr := MakeWebhookHandler(secret, s.eventHandler)
	if whErr != nil {
		return fmt.Errorf("webhook listener: %w", whErr)
	}
	s.webhookHandler = wh
	return nil
}

func (s *Service) SetupSocketMode(client *slack.Client) error {
	sml, smlErr := MakeSocketModeListener(client, s.eventHandler)
	if smlErr != nil {
		return fmt.Errorf("socketmode listener: %w", smlErr)
	}
	s.socketModeListener = sml
	return nil
}

func (s *Service) WebhookHandler() http.Handler {
	return s.webhookHandler
}

func (s *Service) Start(ctx context.Context) error {
	if s.socketModeListener != nil {
		return s.socketModeListener.Start(ctx)
	}
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.socketModeListener != nil {
		return s.socketModeListener.Shutdown(ctx)
	}
	return nil
}

// TODO: look up by some stable install id?

func (s *Service) LookupIntegration(ctx context.Context, ids IntegrationInstallIds) (rez.ConfiguredIntegration, error) {
	params := rez.ListIntegrationsParams{
		Providers:    []string{s.integrationName},
		ConfigValues: ids.configValues(),
	}
	if ids.TeamId != "" {
		params.ExternalRefs = []string{ids.TeamId}
	}
	res, listErr := s.intgs.ListConfigured(execution.NewSystemContext(ctx), params)
	if listErr != nil {
		return nil, fmt.Errorf("listing configured integrations: %w", listErr)
	}
	return res[0], nil
}

func (s *Service) createUserContext(ctx context.Context, ids IntegrationInstallIds, userChatId string) (context.Context, error) {
	ci, intgsErr := s.LookupIntegration(ctx, ids)
	if intgsErr != nil {
		return nil, fmt.Errorf("lookup integration: %w", intgsErr)
	}
	ctx = execution.NewTenantContext(ctx, ci.Integration().TenantID)
	usr, usrErr := s.users.Get(ctx, user.ChatID(userChatId))
	if usrErr != nil {
		slog.Error("failed to lookup chat user", "error", usrErr, "chat_id", userChatId)
		return nil, fmt.Errorf("lookup user: %w", usrErr)
	}
	return execution.NewUserAuthContext(ctx, *usr, time.Time{}), nil
}

func (s *Service) handleEventsApiEvent(ctx context.Context, ev *slackevents.EventsAPIEvent) error {
	ids := IntegrationInstallIds{
		TeamId:       ev.TeamID,
		EnterpriseId: ev.EnterpriseID,
	}
	intg, intgsErr := s.LookupIntegration(ctx, ids)
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}
	ctx = execution.NewTenantContext(ctx, intg.Integration().TenantID)

	if err := s.eventsApiHandler(ctx, ev); err != nil {
		return fmt.Errorf("handle events api event: %w", err)
	}
	return nil
}

func (s *Service) handleInteractionCallback(ctx context.Context, ic *slack.InteractionCallback) error {
	ids := IntegrationInstallIds{
		TeamId:       ic.Team.ID,
		EnterpriseId: ic.Enterprise.ID,
	}
	ctx, usrErr := s.createUserContext(ctx, ids, ic.User.ID)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	if handler, ok := s.interactionCallbackHandlers[ic.Type]; ok {
		return handler(ctx, ic)
	}
	// log unhandled
	slog.Warn("unknown interaction type")
	return nil
}

func (s *Service) handleSlashCommand(baseCtx context.Context, cmd *slack.SlashCommand) error {
	ids := IntegrationInstallIds{
		TeamId:       cmd.TeamID,
		EnterpriseId: cmd.EnterpriseID,
	}
	ctx, usrErr := s.createUserContext(baseCtx, ids, cmd.UserID)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	var response *slack.Blocks
	var handlerErr error
	handler, hasHandler := s.slashCommandHandlers[cmd.Command]
	if hasHandler {
		response, handlerErr = handler(ctx, cmd)
	} else {
		slog.Debug("unknown slack command, ignoring", "command", cmd.Command)
	}
	if handlerErr != nil {
		return fmt.Errorf("handling command: %w", handlerErr)
	}
	if response != nil {
		//_, msgErr := s.postEphemeralMessage(ctx, cmd.ChannelID, cmd.UserID, slack.MsgOptionBlocks(response.BlockSet...))
		//if msgErr != nil {
		//	return fmt.Errorf("failed to post ephemeral message: %w", msgErr)
		//}
	}
	return nil
}

func (s *Service) GetClient(ctx context.Context) *slack.Client {
	return nil
}

func (s *Service) postMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) (string, error) {
	_, msgTs, msgErr := s.GetClient(ctx).PostMessageContext(ctx, channelId, msgOpts...)
	return msgTs, msgErr
}

func (s *Service) postEphemeralMessage(ctx context.Context, channelId, userId string, msgOpts ...slack.MsgOption) (string, error) {
	return s.GetClient(ctx).PostEphemeralContext(ctx, channelId, userId, msgOpts...)
}

func (s *Service) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) (string, error) {
	blocks := ConvertContentToBlocks("", content)
	return s.postMessage(ctx, channelId, slack.MsgOptionBlocks(blocks...))
}

func (s *Service) SendTextMessage(ctx context.Context, channelId string, text string) (string, error) {
	return s.postMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (s *Service) SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error) {
	return s.postMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (s *Service) OpenModalView(ctx context.Context, triggerId string, viewReq slack.ModalViewRequest) error {
	resp, respErr := s.GetClient(ctx).OpenViewContext(ctx, triggerId, viewReq)
	if respErr != nil {
		LogSlackViewErrorResponse(slog.Default(), respErr, resp)
		return respErr
	}
	return nil
}

func (s *Service) OpenOrUpdateModal(ctx context.Context, ic *slack.InteractionCallback, view *slack.ModalViewRequest) error {
	var viewResp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		viewResp, respErr = s.GetClient(ctx).OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		viewResp, respErr = s.GetClient(ctx).UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		LogSlackViewErrorResponse(slog.Default(), respErr, viewResp)
		return respErr
	}

	return nil
}
