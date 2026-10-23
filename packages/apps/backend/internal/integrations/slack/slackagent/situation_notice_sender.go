package slackagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/slack-go/slack"

	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/integrations"
)

// maxNoticeErrorRunes bounds the reason a failed notice shows.
const maxNoticeErrorRunes = 200

var _ integrations.SituationNoticeSender = (*InstalledIntegration)(nil)

func (ii *InstalledIntegration) SituationNoticeChannel() string {
	settings, settingsErr := decodeUserSettings(ii.intg.UserSettings)
	if settingsErr != nil {
		slog.Error("decode slack agent settings", "error", settingsErr, "integration_id", ii.intg.ID)
		return ""
	}
	return settings.SituationNoticeChannelID
}

func (ii *InstalledIntegration) PostSituationNotice(ctx context.Context, channelID string, content *integrations.SituationNoticeContent) (string, error) {
	cw, clientErr := ii.clientWrapper()
	if clientErr != nil {
		return "", clientErr
	}
	messageRef, postErr := cw.PostMessage(ctx, channelID, (*noticeContent)(content).messageOptions()...)
	if postErr != nil {
		return "", noticeDeliveryError(postErr)
	}
	return messageRef, nil
}

func (ii *InstalledIntegration) UpdateSituationNotice(ctx context.Context, channelID, messageRef string, content *integrations.SituationNoticeContent) error {
	cw, clientErr := ii.clientWrapper()
	if clientErr != nil {
		return clientErr
	}
	if updateErr := cw.UpdateMessage(ctx, channelID, messageRef, (*noticeContent)(content).messageOptions()...); updateErr != nil {
		return noticeDeliveryError(updateErr)
	}
	return nil
}

func (ii *InstalledIntegration) SituationNoticeLink(ctx context.Context, channelID, messageRef string) (string, error) {
	cw, clientErr := ii.clientWrapper()
	if clientErr != nil {
		return "", clientErr
	}
	return cw.GetPermalink(ctx, channelID, messageRef)
}

func (ii *InstalledIntegration) ChatChannelName(ctx context.Context, channelID string) (string, error) {
	cw, clientErr := ii.clientWrapper()
	if clientErr != nil {
		return "", clientErr
	}
	return cw.GetChannelName(ctx, channelID)
}

func (ii *InstalledIntegration) clientWrapper() (*slackintegration.ClientWrapper, error) {
	cw, clientErr := slackintegration.NewClientWrapper(ii.intg, ii.clientOptions...)
	if clientErr != nil {
		return nil, fmt.Errorf("create slack client: %w", clientErr)
	}
	return cw, nil
}

var slackTokenPattern = regexp.MustCompile(`xox[a-z]-[A-Za-z0-9-]+`)

// noticeDeliveryError gives Slack's error code as the reason when it gave one, otherwise the error's message
// without any token.
func noticeDeliveryError(deliveryErr error) error {
	reason := truncateRunes(slackTokenPattern.ReplaceAllString(deliveryErr.Error(), "[token]"), maxNoticeErrorRunes)
	if slackErr, isSlackErr := errors.AsType[slack.SlackErrorResponse](deliveryErr); isSlackErr {
		reason = truncateRunes(slackErr.Err, maxNoticeErrorRunes)
	}
	if _, isRateLimited := errors.AsType[*slack.RateLimitedError](deliveryErr); isRateLimited {
		reason = "rate_limited"
	}
	return &integrations.NoticeDeliveryError{Reason: reason, Err: deliveryErr}
}
