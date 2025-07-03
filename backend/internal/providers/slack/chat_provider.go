package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"net/http"

	rez "github.com/rezible/rezible"
)

type ChatProvider struct {
	client         *slack.Client
	webhookHandler *webhookHandler

	annotateMessageFn    rez.AnnotateMessageFn
	lookupUserFn         rez.LookupChatUserFn
	lookupMessageEventFn rez.LookupChatMessageEventFn
}

type ChatProviderConfig struct {
	BotApiKey     string `json:"bot_api_key"`
	SigningSecret string `json:"signing_secret"`
}

func NewChatProvider(cfg ChatProviderConfig) (*ChatProvider, error) {
	p := &ChatProvider{}
	p.client = slack.New(cfg.BotApiKey)
	p.webhookHandler = newWebhookHandler(cfg.SigningSecret, p)
	return p, nil
}

func (p *ChatProvider) SetAnnotateMessageFn(fn func(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error)) {
	p.annotateMessageFn = fn
}
func (p *ChatProvider) SetUserLookupFn(fn func(context.Context, string) (*ent.User, error)) {
	p.lookupUserFn = fn
}
func (p *ChatProvider) SetMessageEventLookupFn(fn func(context.Context, string) (*ent.OncallEvent, error)) {
	p.lookupMessageEventFn = fn
}

func (p *ChatProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"slack/options":     http.HandlerFunc(p.webhookHandler.handleOptions),
		"slack/events":      http.HandlerFunc(p.webhookHandler.handleEvents),
		"slack/interaction": http.HandlerFunc(p.webhookHandler.handleInteractions),
	}
}

func (p *ChatProvider) sendMessage(ctx context.Context, channel string, msg slack.MsgOption) error {
	_, _, msgErr := p.client.PostMessageContext(ctx, channel, msg)
	return msgErr
}

func (p *ChatProvider) sendUserMessage(ctx context.Context, id string, msg slack.MsgOption) error {
	params := &slack.OpenConversationParameters{Users: []string{id}}
	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, params)
	if convoErr != nil {
		return fmt.Errorf("failed to open conversation with user %s: %w", id, convoErr)
	}

	if sendErr := p.sendMessage(ctx, convo.ID, msg); sendErr != nil {
		return fmt.Errorf("send user %s message: %w", id, sendErr)
	}

	return nil
}

func (p *ChatProvider) SendMessage(ctx context.Context, id string, content *rez.ContentNode) error {
	return p.sendUserMessage(ctx, id, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (p *ChatProvider) SendTextMessage(ctx context.Context, id string, text string) error {
	return p.sendUserMessage(ctx, id, slack.MsgOptionText(text, false))
}

func (p *ChatProvider) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	mb, builderErr := newHandoverMessageBuilder(params.EndingShift, params.StartingShift, params.PinnedAnnotations)
	if builderErr != nil {
		return fmt.Errorf("failed to create handover message builder: %w", builderErr)
	}

	if buildErr := mb.build(params.Content); buildErr != nil {
		return fmt.Errorf("building handover message: %w", buildErr)
	}

	return p.sendMessage(ctx, mb.getChannel(), mb.getMessage())
}

func (p *ChatProvider) onMentionEvent(data *slackevents.AppMentionEvent) {
	_, _, msgErr := p.client.PostMessage(data.Channel, slack.MsgOptionText("hello", false))
	if msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to message")
	}
}

func logSlackViewErrorResponse(err error, resp *slack.ViewResponse) {
	if resp != nil {
		log.Debug().
			Strs("messages", resp.ResponseMetadata.Messages).
			Msg("publish response")
	}
	log.Error().Err(err).Msg("slack view response error")
}

func (p *ChatProvider) onUserHomeOpenedEvent(data *slackevents.AppHomeOpenedEvent) {
	ctx := context.Background()
	homeView, hash, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		log.Error().Err(viewErr).Msg("failed to create user home view")
		return
	}
	resp, publishErr := p.client.PublishViewContext(ctx, data.User, *homeView, hash)
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
	}
}

func (p *ChatProvider) onMessageEvent(data *slackevents.MessageEvent) {
	log.Debug().Interface("data", data).Msg("slack message event")
}

func (p *ChatProvider) queryUserRosters(ctx context.Context, userId string) ([]*ent.OncallRoster, error) {
	if p.lookupUserFn == nil || p.lookupMessageEventFn == nil {
		return nil, fmt.Errorf("lookup funcs nil")
	}

	user, userErr := p.lookupUserFn(ctx, userId)
	if userErr != nil {
		return nil, userErr
	}

	rosters, rostersErr := user.QueryOncallSchedules().QuerySchedule().QueryRoster().All(ctx)
	if rostersErr != nil && !ent.IsNotFound(rostersErr) {
		return nil, fmt.Errorf("failed to query oncall rosters for user: %w", rostersErr)
	}

	return rosters, nil
}

func getMessageId(ic *slack.InteractionCallback) string {
	return fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)
}

func (p *ChatProvider) makeCreateAnnotationViewDetails(ctx context.Context, ic *slack.InteractionCallback) (*createAnnotationModalDetails, error) {
	msgId := getMessageId(ic)
	userId := ic.User.ID

	d := &createAnnotationModalDetails{}
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &d.metadata); jsonErr != nil {
			return nil, jsonErr
		}
	} else {
		d.metadata = createAnnotationMessageMetadata{
			MsgId:        msgId,
			UserId:       ic.Message.User,
			MsgText:      ic.Message.Text,
			MsgTimestamp: convertSlackTs(ic.MessageTs),
		}
	}
	var rostersErr error
	d.rosters, rostersErr = p.queryUserRosters(ctx, userId)
	if rostersErr != nil {
		return nil, fmt.Errorf("failed to get annotation information: %w", rostersErr)
	}

	_, selectedId := getSelectedRoster(ic.View.State)
	if selectedId != uuid.Nil && len(d.rosters) > 0 {
		for _, roster := range d.rosters {
			if roster.ID == selectedId {
				d.selectedRoster = roster
				break
			}
		}
		event, eventErr := p.lookupMessageEventFn(ctx, msgId)
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to lookup message event: %w", eventErr)
		}
		if event != nil {
			anno, annoErr := event.QueryAnnotations().Where(oncallannotation.RosterID(selectedId)).First(ctx)
			if annoErr != nil && !ent.IsNotFound(annoErr) {
				return nil, fmt.Errorf("failed to lookup existing event annotation: %w", annoErr)
			}
			d.currAnnotation = anno
		}
	}

	return d, nil
}

func (p *ChatProvider) handleCreateAnnotationInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	details, detailsErr := p.makeCreateAnnotationViewDetails(ctx, ic)
	if detailsErr != nil {
		return fmt.Errorf("failed to get message annotation context: %w", detailsErr)
	}

	view, viewErr := makeCreateAnnotationModalView(details)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	resp, respErr := p.client.OpenViewContext(ctx, ic.TriggerID, *view)
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return fmt.Errorf("open view: %w", respErr)
	}
	return nil
}

func (p *ChatProvider) handleBlockActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		details, detailsErr := p.makeCreateAnnotationViewDetails(ctx, ic)
		if detailsErr != nil {
			return fmt.Errorf("failed to get message annotation context: %w", detailsErr)
		}
		view, viewErr := makeCreateAnnotationModalView(details)
		if viewErr != nil || view == nil {
			return fmt.Errorf("failed to create annotation view: %w", viewErr)
		}

		resp, respErr := p.client.UpdateViewContext(ctx, *view, "", "", ic.View.ID)
		if respErr != nil {
			logSlackViewErrorResponse(respErr, resp)
			return fmt.Errorf("open view: %w", respErr)
		}
	}
	return nil
}

func (p *ChatProvider) handleCreateAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	user, userErr := p.lookupUserFn(ctx, ic.User.ID)
	if userErr != nil || user == nil {
		return fmt.Errorf("failed to lookup user: %w", userErr)
	}
	anno, annoErr := getCreateAnnotationModalViewAnnotation(ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	anno.CreatorID = user.ID
	if p.annotateMessageFn == nil {
		return fmt.Errorf("no chat message annotator")
	}
	_, createErr := p.annotateMessageFn(ctx, anno)
	return createErr
}
