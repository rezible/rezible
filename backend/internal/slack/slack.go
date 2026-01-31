package slack

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type loader struct {
	svcs *rez.Services
}

func newLoader(svcs *rez.Services) *loader {
	return &loader{svcs: svcs}
}

func (l *loader) loadWithIntegration(intg *ent.Integration) (*ChatService, error) {
	cfg, cfgErr := decodeConfig(intg.Config)
	if cfgErr != nil {
		return nil, fmt.Errorf("unable to decode config: %w", cfgErr)
	}
	return newChatService(cfg.makeClient(), l.svcs), nil
}

func (l *loader) loadByTenantLookup(ctx context.Context, teamId string, enterpriseId string) (*ChatService, context.Context, error) {
	intg, lookupErr := lookupIntegration(ctx, l.svcs.Integrations, teamId, enterpriseId)
	if lookupErr != nil {
		return nil, nil, lookupErr
	}
	chat, chatErr := l.loadWithIntegration(intg)
	if chatErr != nil {
		return nil, nil, fmt.Errorf("load chat service failed: %w", chatErr)
	}
	return chat, access.TenantContext(ctx, intg.TenantID), nil
}

func (l *loader) loadFromContext(ctx context.Context) (*ChatService, error) {
	intg, lookupErr := l.svcs.Integrations.Get(ctx, integrationName)
	if lookupErr != nil {
		return nil, lookupErr
	}
	return l.loadWithIntegration(intg)
}

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
