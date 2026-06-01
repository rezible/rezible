package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type IntegrationInstallIds struct {
	TeamId       string `json:"teamId"`
	EnterpriseId string `json:"enterpriseId,omitempty"`
}

func (i IntegrationInstallIds) configValues() map[string]any {
	m := map[string]any{}
	if i.TeamId != "" {
		m["team.id"] = i.TeamId
	}
	if i.EnterpriseId != "" {
		m["enterprise.id"] = i.EnterpriseId
	}
	return m
}

func GetAllUsersInConversation(ctx context.Context, client *slack.Client, convId string) ([]string, error) {
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

func LogSlackViewErrorResponse(logger *slog.Logger, err error, resp *slack.ViewResponse) {
	args := []any{"error", err}
	if resp != nil {
		args = append(args, "response_messages", resp.ResponseMetadata.Messages)
	}
	logger.Error("slack view response error", args...)
}

func CommandErrorResponse(message string) *slack.Blocks {
	return &slack.Blocks{
		BlockSet: []slack.Block{
			&slack.SectionBlock{
				Type: slack.MBTSection,
				Text: PlainTextBlock(fmt.Sprintf("❌ %s", message)),
			},
		},
	}
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
