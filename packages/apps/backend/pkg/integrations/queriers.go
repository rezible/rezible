package integrations

import "context"

// ChatChannelQuerier lists channels from a chat provider on demand. Results are
// returned to the caller and are never ingested as provider events.
type ChatChannelQuerier interface {
	ListChatChannels(ctx context.Context, params ListChatChannelsParams) (*ChatChannelPage, error)
}

const (
	DefaultChatChannelPageSize = 100
	MaxChatChannelPageSize     = 200
)

type ListChatChannelsParams struct {
	Cursor         string // opaque provider cursor from a previous page; empty for the first page
	Limit          int    // 0 uses DefaultChatChannelPageSize
	IncludePrivate bool   // backend-only; the API always sends false
}

type ChatChannel struct {
	ID         string
	Name       string
	IsPrivate  bool
	IsArchived bool
	IsMember   bool // the app's bot is a member, not the current user
}

type ChatChannelPage struct {
	Channels   []ChatChannel
	NextCursor string // empty when the provider reports no more pages
}
