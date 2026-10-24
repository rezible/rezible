package integrations

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
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

// SituationNoticeSender posts an installation's notices of raised situations to the chat channel its admin
// chose, and edits them as the situations change. Errors it returns are shown to people when they are
// NoticeDeliveryErrors.
type SituationNoticeSender interface {
	// SituationNoticeChannel is the channel notices are posted to, or empty to post none.
	SituationNoticeChannel() string
	// PostSituationNotice posts the notice and returns the reference its message is edited by.
	PostSituationNotice(ctx context.Context, channelID string, content *SituationNoticeContent) (string, error)
	UpdateSituationNotice(ctx context.Context, channelID, messageRef string, content *SituationNoticeContent) error
	// SituationNoticeLink is the link to the posted notice's message.
	SituationNoticeLink(ctx context.Context, channelID, messageRef string) (string, error)
	ChatChannelName(ctx context.Context, channelID string) (string, error)
}

// SituationNoticeContent is what a notice shows, read when it is synced.
type SituationNoticeContent struct {
	// Situation has its links, with their linked situations, loaded.
	Situation *ent.Situation
	Services  []*ent.KnowledgeEntity
	// Deployment is the latest deployment before the situation started, or nil to omit it.
	Deployment *rez.SituationLatestDeployment
	// Report is the investigation's latest completed report, or nil while there is none.
	Report *ent.InvestigationReport
	// SituationURL links to the situation, and MergedIntoURL to the one it was merged into, if any.
	SituationURL  string
	MergedIntoURL string
	// IncidentURL links to the newest unresolved incident the situation is part of, if any. Otherwise
	// DeclareIncidentURL opens the declare dialog for the situation, when declaring is possible.
	IncidentURL        string
	DeclareIncidentURL string
	Now                time.Time
}

// MergedInto is the situation's link to the one it was merged into, or nil.
func (c *SituationNoticeContent) MergedInto() *ent.SituationLink {
	for _, link := range c.Situation.Edges.Links {
		if link.Kind == sitlink.KindMergedInto {
			return link
		}
	}
	return nil
}

// NoticeDeliveryError is a notice the provider did not accept. Reason is shown to people, so it never
// holds credentials.
type NoticeDeliveryError struct {
	Reason string
	Err    error
}

func (e *NoticeDeliveryError) Error() string {
	return e.Reason
}

func (e *NoticeDeliveryError) Unwrap() error {
	return e.Err
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
	// CheckLogsAvailable returns why the installation cannot read logs, or nil, without calling the provider.
	CheckLogsAvailable() error
}

// MetricQuerier reads metrics on demand. Results are returned to the caller and
// are never ingested as provider events.
type MetricQuerier interface {
	ListMetricNames(ctx context.Context, params MetricNamesParams) (*MetricNamesResult, error)
	QueryMetricRange(ctx context.Context, params MetricRangeParams) (*MetricRangeResult, error)
}

// LogService identifies the service whose logs are read. The integration chooses the value to query by.
type LogService struct {
	DisplayName string
	// Aliases are the reading installation's own provider resource refs for the service.
	Aliases []string
}

// LogServiceName is the value an integration read a service's logs by.
type LogServiceName struct {
	Value string
	// FromAlias is whether the value came from the installation's aliases rather than the display name.
	FromAlias bool
}

type LogSearchParams struct {
	Service  LogService
	Start    time.Time
	End      time.Time
	Contains string // optional literal text each line must contain
	Limit    int    // 0 uses DefaultLogSearchLimit
}

type LogCountParams struct {
	Service  LogService
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
	Service LogServiceName
	Lines   []LogLine // newest first
	Read    ReadRecord
}

type MetricPoint struct {
	Time  time.Time
	Value float64
}

type LogCountResult struct {
	Service LogServiceName
	Points  []MetricPoint
	Read    ReadRecord
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

// Limits on log and metric reads. A request outside them is errs.ErrInvalidInput naming the limit; the
// others adjust the result, and the result's ReadRecord.Limits says so.
const (
	MaxTelemetryReadRange = 6 * time.Hour
	DefaultLogSearchLimit = 50
	MaxLogSearchLimit     = 200
	MaxLogLineLength      = 2000 // runes
	MaxTelemetryPoints    = 300  // per series; sets the minimum step
	MaxMetricSeries       = 20
	MaxMetricNames        = 200
	TelemetryReadTimeout  = 10 * time.Second
)

func (p LogSearchParams) Validate() error {
	if serviceErr := validateReadService(p.Service.DisplayName); serviceErr != nil {
		return serviceErr
	}
	if p.Limit < 0 {
		return fmt.Errorf("%w: limit must not be negative", errs.ErrInvalidInput)
	}
	if p.Limit > MaxLogSearchLimit {
		return fmt.Errorf("%w: limit of %d is above the maximum of %d lines", errs.ErrInvalidInput, p.Limit, MaxLogSearchLimit)
	}
	return validateReadRange(p.Start, p.End)
}

// EffectiveLimit is the number of lines the search may return.
func (p LogSearchParams) EffectiveLimit() int {
	if p.Limit == 0 {
		return DefaultLogSearchLimit
	}
	return p.Limit
}

func (p LogCountParams) Validate() error {
	if serviceErr := validateReadService(p.Service.DisplayName); serviceErr != nil {
		return serviceErr
	}
	if p.Step < 0 {
		return fmt.Errorf("%w: step must not be negative", errs.ErrInvalidInput)
	}
	return validateReadRange(p.Start, p.End)
}

func (p MetricNamesParams) Validate() error {
	if serviceErr := validateReadService(p.Service); serviceErr != nil {
		return serviceErr
	}
	return validateReadRange(p.Start, p.End)
}

func (p MetricRangeParams) Validate() error {
	if strings.TrimSpace(p.Query) == "" {
		return fmt.Errorf("%w: query is required", errs.ErrInvalidInput)
	}
	if p.Step < 0 {
		return fmt.Errorf("%w: step must not be negative", errs.ErrInvalidInput)
	}
	return validateReadRange(p.Start, p.End)
}

func validateReadService(service string) error {
	if strings.TrimSpace(service) == "" {
		return fmt.Errorf("%w: service is required", errs.ErrInvalidInput)
	}
	return nil
}

func validateReadRange(start, end time.Time) error {
	if !end.After(start) {
		return fmt.Errorf("%w: end must be after start", errs.ErrInvalidInput)
	}
	if end.Sub(start) > MaxTelemetryReadRange {
		return fmt.Errorf("%w: range of %s is longer than the %s limit", errs.ErrInvalidInput, end.Sub(start), MaxTelemetryReadRange)
	}
	return nil
}

// ReadStep returns the step, in whole seconds, for a range read of at most MaxTelemetryPoints points. A
// zero requested step uses the minimum. When a smaller step is raised, it also returns the limit to record.
func ReadStep(start, end time.Time, requested time.Duration) (time.Duration, string) {
	minSeconds := max(1, math.Ceil(end.Sub(start).Seconds()/MaxTelemetryPoints))
	minStep := time.Duration(minSeconds) * time.Second
	if requested == 0 {
		return minStep, ""
	}
	if requested < minStep {
		limit := fmt.Sprintf("step raised from %s to %s to stay within %d points", requested, minStep, MaxTelemetryPoints)
		return minStep, limit
	}
	return time.Duration(math.Ceil(requested.Seconds())) * time.Second, ""
}

type readMetrics struct {
	duration   metric.Float64Histogram
	resultSize metric.Int64Histogram
}

var loadReadMetrics = sync.OnceValues(func() (*readMetrics, error) {
	meter := otel.Meter("github.com/rezible/rezible")
	duration, durationErr := meter.Float64Histogram("rezible.backend.provider.read.duration",
		metric.WithDescription("Provider read duration"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	resultSize, resultSizeErr := meter.Int64Histogram("rezible.backend.provider.read.result_size",
		metric.WithDescription("Items returned by successful provider reads"),
		metric.WithUnit("{item}"))
	if instrumentsErr := errors.Join(durationErr, resultSizeErr); instrumentsErr != nil {
		return nil, instrumentsErr
	}
	return &readMetrics{duration: duration, resultSize: resultSize}, nil
})

// MeasureRead runs one provider read as the unit of work "provider.read" and records its measurements.
// read returns how many items it returned.
func MeasureRead(ctx context.Context, provider, operation string, read func(context.Context) (int, error)) error {
	metrics, metricsErr := loadReadMetrics()
	if metricsErr != nil {
		return fmt.Errorf("provider read metrics: %w", metricsErr)
	}

	providerAttr := attribute.String("provider", provider)
	operationAttr := attribute.String("operation", operation)
	var items int
	started := time.Now()
	readErr := execution.Do(ctx, "provider.read", func(ctx context.Context) error {
		var itemsErr error
		items, itemsErr = read(ctx)
		return itemsErr
	}, providerAttr, operationAttr)
	duration := time.Since(started)

	// The provider may have been called whatever the outcome.
	rezai.AddRead(ctx)
	outcome := "succeeded"
	if readErr != nil {
		outcome = "failed"
	}
	metrics.duration.Record(ctx, duration.Seconds(), metric.WithAttributes(providerAttr, operationAttr, attribute.String("outcome", outcome)))
	if readErr == nil {
		metrics.resultSize.Record(ctx, int64(items), metric.WithAttributes(providerAttr, operationAttr))
	}
	return readErr
}

// DirectMessageSender renders and sends plain text to a member's chat account.
type DirectMessageSender interface {
	SendDirectMessage(context.Context, *ent.User, *DirectMessage) error
}

type DirectMessage struct {
	Text      string
	LinkURL   string
	LinkLabel string
}

var ErrNoRecipient = errors.New("no chat recipient")
