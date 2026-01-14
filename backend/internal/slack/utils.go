package slack

import (
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func logSlackViewErrorResponse(err error, resp *slack.ViewResponse) {
	line := log.Error().Err(err)
	if resp != nil {
		line.Strs("response_messages", resp.ResponseMetadata.Messages)
	}
	line.Msg("slack view response error")
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

type messageId string

func (m messageId) getTimestamp() time.Time {
	_, ts, _ := strings.Cut(m.String(), "_")
	return convertSlackTs(ts)
}

func (m messageId) String() string {
	return string(m)
}
