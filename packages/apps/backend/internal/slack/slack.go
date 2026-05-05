package slack

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func getAllUsersInConversation(ctx context.Context, client *slack.Client, convId string) ([]string, error) {
	params := &slack.GetUsersInConversationParameters{
		ChannelID: convId,
		Limit:     100,
	}
	var allIds []string
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

func logSlackViewErrorResponse(err error, resp *slack.ViewResponse) {
	args := []any{"error", err}
	if resp != nil {
		args = append(args, "response_messages", resp.ResponseMetadata.Messages)
	}
	slog.Error("slack view response error", args...)
}

func convertSlackTs(ts string) time.Time {
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

func tryConvertTs(ts string, fallback time.Time) time.Time {
	if conv := convertSlackTs(ts); !conv.IsZero() {
		return conv
	}
	return fallback
}

type messageId string

func (m messageId) getTimestamp() time.Time {
	_, ts, _ := strings.Cut(m.String(), "_")
	return convertSlackTs(ts)
}

func (m messageId) String() string {
	return string(m)
}
