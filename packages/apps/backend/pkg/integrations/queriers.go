package integrations

import (
	"context"
	"time"
)

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

// HealthChecker reports whether an installation can currently reach its provider.
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}

// LogQuerier reads logs for one service on demand. Results are returned to the
// caller and are never ingested as provider events.
type LogQuerier interface {
	SearchLogs(ctx context.Context, params LogSearchParams) (*LogSearchResult, error)
	CountLogs(ctx context.Context, params LogCountParams) (*LogCountResult, error)
}

// MetricQuerier reads metrics on demand. Results are returned to the caller and
// are never ingested as provider events.
type MetricQuerier interface {
	ListMetricNames(ctx context.Context, params MetricNamesParams) (*MetricNamesResult, error)
	QueryMetricRange(ctx context.Context, params MetricRangeParams) (*MetricRangeResult, error)
}

type LogSearchParams struct {
	Service  string // the value of the installation's log service label
	Start    time.Time
	End      time.Time
	Contains string // optional literal text each line must contain
	Limit    int    // 0 uses DefaultLogSearchLimit
}

type LogCountParams struct {
	Service  string // the value of the installation's log service label
	Start    time.Time
	End      time.Time
	Contains string        // optional literal text each counted line must contain
	Step     time.Duration // 0 uses the minimum step for the range
}

type MetricNamesParams struct {
	Service string // the value of the installation's metric service label
	Start   time.Time
	End     time.Time
}

type MetricRangeParams struct {
	Query string // PromQL, not bound to a service
	Start time.Time
	End   time.Time
	Step  time.Duration // 0 uses the minimum step for the range
}

// ReadRecord describes the read that produced a result.
type ReadRecord struct {
	DataSourceUID string
	Query         string // the exact query sent
	Start         time.Time
	End           time.Time
	Step          time.Duration // zero when the read has no step
	RanAt         time.Time
	Limits        []string // each limit that cut or adjusted the result, in plain words
	ExploreURL    string   // opens the same query and range in the provider; empty when there is none
}

type LogLine struct {
	Time   time.Time
	Line   string
	Labels map[string]string
}

type LogSearchResult struct {
	Lines []LogLine // newest first
	Read  ReadRecord
}

type MetricPoint struct {
	Time  time.Time
	Value float64
}

type LogCountResult struct {
	Points []MetricPoint
	Read   ReadRecord
}

type MetricSeries struct {
	Labels map[string]string
	Points []MetricPoint
}

type MetricRangeResult struct {
	Series []MetricSeries
	Read   ReadRecord
}

type MetricNamesResult struct {
	Names []string
	Read  ReadRecord
}
