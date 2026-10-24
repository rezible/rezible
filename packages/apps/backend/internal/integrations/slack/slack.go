package slackintegration

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

const ProviderName = "slack"

type InstallationIds struct {
	TeamId       string
	EnterpriseId string
}

func (i InstallationIds) InstallationTargetResourceRef() string {
	if i.TeamId != "" {
		return "team:" + i.TeamId
	}
	if i.EnterpriseId != "" {
		return "enterprise:" + i.EnterpriseId
	}
	return ""
}

func GetAllUsersInConversation(ctx context.Context, cw *ClientWrapper, convId string) ([]string, error) {
	params := &slack.GetUsersInConversationParameters{
		ChannelID: convId,
		Limit:     100,
	}
	var allIds []string
	client := cw.Client()
	for {
		ids, cursor, getErr := client.GetUsersInConversationContext(ctx, params)
		if getErr != nil {
			return nil, getErr
		}
		allIds = append(allIds, ids...)
		params.Cursor = cursor
		if cursor == "" || len(ids) == 0 {
			break
		}
	}
	return allIds, nil
}

func LogSlackViewErrorResponse(logger *slog.Logger, err error, resp *slack.ViewResponse) {
	args := []any{"error", err}
	if resp != nil {
		args = append(args, "response_messages", resp.ResponseMetadata.Messages)
	}
	logger.Error("slack view response error", args...)
}

// CommandReply is a command's reply: one section of mrkdwn, shown verbatim.
func CommandReply(text string) *slack.Blocks {
	return &slack.Blocks{BlockSet: []slack.Block{slack.NewSectionBlock(MarkdownBlock(text), nil, nil)}}
}

// slackTextEscaper escapes the characters Slack reads as markup, so text from people and sources cannot
// form mentions or links.
var slackTextEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// EscapeText escapes text Rezible did not write for use in mrkdwn.
func EscapeText(text string) string {
	return slackTextEscaper.Replace(text)
}

// SlackDate is a date token, which Slack shows in each reader's time zone.
func SlackDate(at time.Time) string {
	fallback := at.UTC().Format("2006-01-02 15:04 UTC")
	return fmt.Sprintf("<!date^%d^{date_short_pretty} at {time}|%s>", at.Unix(), fallback)
}

// TruncateRunes shortens text to at most limit runes, ending it with an ellipsis when cut.
func TruncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}

func ConvertSlackTs(ts string) time.Time {
	parts := strings.Split(ts, ".")
	if len(parts) < 2 {
		return time.Time{}
	}
	secs, parseErr := strconv.ParseInt(parts[0], 10, 32)
	if parseErr != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}

func TryConvertSlackTs(ts string, fallback time.Time) time.Time {
	if conv := ConvertSlackTs(ts); !conv.IsZero() {
		return conv
	}
	return fallback
}

type MessageId string

func (m MessageId) GetTimestamp() time.Time {
	_, ts, _ := strings.Cut(m.String(), "_")
	return ConvertSlackTs(ts)
}

func (m MessageId) String() string {
	return string(m)
}
