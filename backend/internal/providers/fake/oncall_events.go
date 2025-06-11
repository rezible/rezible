package fakeprovider

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"iter"
	"math/rand"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type OncallEventsDataProvider struct {
	providerUserMap map[string]*ent.User

	webhookCallback rez.DataProviderResourceUpdatedCallback
}

type OncallEventsDataProviderConfig struct {
}

func NewOncallEventsDataProvider(cfg OncallEventsDataProviderConfig) (*OncallEventsDataProvider, error) {
	p := &OncallEventsDataProvider{
		providerUserMap: make(map[string]*ent.User),
		webhookCallback: func(providerId string, updatedAt time.Time) {},
	}

	return p, nil
}

func (p *OncallEventsDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{}
}

type fakeAlertDesc struct {
	Title       string
	Description string
	Service     string
	AlertKey    string
}

var fakeAlerts = []fakeAlertDesc{
	{
		Title:       "Search API Response Time High",
		Description: "Average search API response time exceeded 2000ms threshold for the last 5 minutes. Current average: 3.2s",
		Service:     "search-api",
		AlertKey:    "search-api-response-time-high",
	},
	{
		Title:       "Elasticsearch Cluster CPU Usage Critical",
		Description: "Elasticsearch cluster CPU usage at 95% for production-search-cluster-1. Query performance severely degraded",
		Service:     "elasticsearch-production",
		AlertKey:    "es-cluster-cpu-critical",
	},
	{
		Title:       "Search Index Build Failed",
		Description: "Nightly search index rebuild job failed with exit code 1. Product catalog search may show stale results",
		Service:     "search-indexer",
		AlertKey:    "search-index-build-failure",
	},
	{
		Title:       "Search Service 4xx Error Rate Spike",
		Description: "Search service returning 4xx errors at 15% rate (normal: <2%). Potential issue with query validation or rate limiting",
		Service:     "search-api",
		AlertKey:    "search-4xx-error-spike",
	},
	{
		Title:       "Redis Search Cache Down",
		Description: "Redis instance search-cache-prod-01 is unreachable. Search queries bypassing cache, increased database load expected",
		Service:     "redis-search-cache",
		AlertKey:    "redis-search-cache-down",
	},
	{
		Title:       "Search Query Queue Backing Up",
		Description: "Search query processing queue depth at 5000+ messages (normal: <500). Processing lag of 45+ seconds",
		Service:     "search-query-processor",
		AlertKey:    "search-queue-backlog",
	},
	{
		Title:       "Search Cache Hit Rate Below Optimal",
		Description: "Redis search cache hit rate dropped to 78% (optimal: >85%). No immediate impact but may indicate cache tuning needed",
		Service:     "redis-search-cache",
		AlertKey:    "search-cache-hit-rate-suboptimal",
	},
	{
		Title:       "Search Analytics Data Pipeline Delay",
		Description: "Search analytics ETL job completed 15 minutes late. Daily search metrics dashboard will be slightly delayed",
		Service:     "search-analytics-pipeline",
		AlertKey:    "search-analytics-pipeline-delay",
	},
	{
		Title:       "Elasticsearch Shard Rebalancing Started",
		Description: "Elasticsearch cluster initiated automatic shard rebalancing on production-search-cluster-2. No performance impact expected",
		Service:     "elasticsearch-production",
		AlertKey:    "es-shard-rebalancing-info",
	},
	{
		Title:       "Search API Rate Limiting Activated",
		Description: "Rate limiting triggered for IP 192.168.1.45 due to 120 requests/minute (limit: 100/min). User experience unaffected",
		Service:     "search-api",
		AlertKey:    "search-rate-limit-activated",
	},
	{
		Title:       "Search Index Size Growth Warning",
		Description: "Product search index grew by 8% this week, now 2.1TB. Consider reviewing retention policies within next 30 days",
		Service:     "elasticsearch-production",
		AlertKey:    "search-index-size-growth",
	},
	{
		Title:       "Search Spell Check Dictionary Updated",
		Description: "Weekly spell check dictionary update completed successfully. 847 new terms added, 23 terms removed",
		Service:     "search-spellcheck",
		AlertKey:    "spellcheck-dictionary-updated",
	},
}

func makeFakeDayAlertEvent(date time.Time) *ent.OncallEvent {
	id := uuid.New()
	hour := rand.Intn(24)
	minute := rand.Intn(60)
	alert := fakeAlerts[rand.Intn(len(fakeAlerts))]
	return &ent.OncallEvent{
		ID:         id,
		ProviderID: id.String(),
		Timestamp: time.Date(
			date.Year(), date.Month(), date.Day(),
			hour, minute, 0, 0, date.Location(),
		),
		Source:      "fake",
		Kind:        "alert",
		Title:       alert.Title,
		Description: alert.Description,
	}
}

func makeFakeOncallEvents(start, end time.Time) []*ent.OncallEvent {
	numHours := end.Sub(start).Hours()
	if numHours <= 0 {
		return nil
	}
	numDays := int(numHours / 24)
	maxDailyEvents := 20
	events := make([]*ent.OncallEvent, 0, numDays*maxDailyEvents)

	for day := 0; day < numDays; day++ {
		for i := 0; i < rand.Intn(maxDailyEvents); i++ {
			events = append(events, makeFakeDayAlertEvent(start.AddDate(0, 0, day)))
		}
	}
	log.Debug().
		Int("days", numDays).
		Int("total", len(events)).
		Msg("created fake oncall events")

	return events
}

func (p *OncallEventsDataProvider) Source() string {
	return "fake"
}

func (p *OncallEventsDataProvider) PullEventsBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallEvent, error] {
	fakeEvents := makeFakeOncallEvents(time.Now().Add(-7*time.Hour*24), time.Now())
	return func(yield func(*ent.OncallEvent, error) bool) {
		for _, event := range fakeEvents {
			if !yield(event, nil) {
				break
			}
		}
	}
}
