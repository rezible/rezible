package fakeprovider

import (
	"context"
	"iter"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/event"
	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/ent"
)

type AlertDataProvider struct {
	providerUserMap map[string]*ent.User
}

type AlertDataProviderConfig struct {
}

func NewAlertDataProvider(intg *ent.Integration) (*AlertDataProvider, error) {
	p := &AlertDataProvider{
		providerUserMap: make(map[string]*ent.User),
	}

	return p, nil
}

var fakeAlerts = []*ent.Alert{
	{
		Title: "Search API Response Time High",
		//Description: "Average search API response time exceeded 2000ms threshold for the last 5 minutes. Current average: 3.2s",
		//Service:     "search-api",
		ExternalID: "search-api-response-time-high",
	},
	{
		Title: "Elasticsearch Cluster CPU Usage Critical",
		//Description: "Elasticsearch cluster CPU usage at 95% for production-search-cluster-1. Query performance severely degraded",
		//Service:     "elasticsearch-production",
		ExternalID: "es-cluster-cpu-critical",
	},
	{
		Title: "Search Index Build Failed",
		//Description: "Nightly search index rebuild job failed with exit code 1. Product catalog search may show stale results",
		//Service:     "search-indexer",
		ExternalID: "search-index-build-failure",
	},
	{
		Title: "Search Service 4xx Error Rate Spike",
		//Description: "Search service returning 4xx errors at 15% rate (normal: <2%). Potential issue with query validation or rate limiting",
		//Service:     "search-api",
		ExternalID: "search-4xx-error-spike",
	},
	{
		Title: "Redis Search Cache Down",
		//Description: "Redis instance search-cache-prod-01 is unreachable. Search queries bypassing cache, increased database load expected",
		//Service:     "redis-search-cache",
		ExternalID: "redis-search-cache-down",
	},
	{
		Title: "Search Query Queue Backing Up",
		//Description: "Search query processing queue depth at 5000+ messages (normal: <500). Processing lag of 45+ seconds",
		//Service:     "search-query-processor",
		ExternalID: "search-queue-backlog",
	},
	{
		Title: "Search Cache Hit Rate Below Optimal",
		//Description: "Redis search cache hit rate dropped to 78% (optimal: >85%). No immediate impact but may indicate cache tuning needed",
		//Service:     "redis-search-cache",
		ExternalID: "search-cache-hit-rate-suboptimal",
	},
	{
		Title: "Search Analytics Data Pipeline Delay",
		//Description: "Search analytics ETL job completed 15 minutes late. Daily search metrics dashboard will be slightly delayed",
		//Service:     "search-analytics-pipeline",
		ExternalID: "search-analytics-pipeline-delay",
	},
	{
		Title: "Elasticsearch Shard Rebalancing Started",
		//Description: "Elasticsearch cluster initiated automatic shard rebalancing on production-search-cluster-2. No performance impact expected",
		//Service:     "elasticsearch-production",
		ExternalID: "es-shard-rebalancing-info",
	},
	{
		Title: "Search API Rate Limiting Activated",
		//Description: "Rate limiting triggered for IP 192.168.1.45 due to 120 requests/minute (limit: 100/min). User experience unaffected",
		//Service:     "search-api",
		ExternalID: "search-rate-limit-activated",
	},
	{
		Title: "Search Index Size Growth Warning",
		//Description: "Product search index grew by 8% this week, now 2.1TB. Consider reviewing retention policies within next 30 days",
		//Service:     "elasticsearch-production",
		ExternalID: "search-index-size-growth",
	},
	{
		Title: "Search Spell Check Dictionary Updated",
		//Description: "Weekly spell check dictionary update completed successfully. 847 new terms added, 23 terms removed",
		//Service:     "search-spellcheck",
		ExternalID: "spellcheck-dictionary-updated",
	},
}

func (p *AlertDataProvider) Source() string {
	return "fake"
}

func (p *AlertDataProvider) PullAlerts(ctx context.Context) iter.Seq2[*ent.Alert, error] {
	return func(yield func(*ent.Alert, error) bool) {
		for _, fa := range fakeAlerts {
			if !yield(fa, nil) {
				break
			}
		}
	}
}

func (p *AlertDataProvider) makeFakeDayAlertInstance(date time.Time) *ent.AlertInstance {
	id := uuid.New()
	hour := rand.Intn(24)
	minute := rand.Intn(60)
	alert := fakeAlerts[rand.Intn(len(fakeAlerts))]

	e := &ent.Event{
		Timestamp: time.Date(
			date.Year(), date.Month(), date.Day(),
			hour, minute, 0, 0, date.Location(),
		),
		Source:      "fake",
		Kind:        event.KindAlert,
		Title:       alert.Title,
		Description: "", // alert.Description,
	}
	return &ent.AlertInstance{
		ID:         id,
		ExternalID: id.String(),
		Edges: ent.AlertInstanceEdges{
			Alert: alert,
			Event: e,
		},
	}
}

func (p *AlertDataProvider) makeFakeAlertInstances(start, end time.Time) []*ent.AlertInstance {
	numHours := end.Sub(start).Hours()
	if numHours <= 0 {
		return nil
	}
	numDays := int(numHours / 24)
	maxDaily := 20
	instances := make([]*ent.AlertInstance, 0, numDays*maxDaily)

	for day := 0; day < numDays; day++ {
		for i := 0; i < rand.Intn(maxDaily); i++ {
			instances = append(instances, p.makeFakeDayAlertInstance(start.AddDate(0, 0, day)))
		}
	}
	log.Debug().
		Int("days", numDays).
		Int("total", len(instances)).
		Msg("created fake oncall events")

	return instances
}

func (p *AlertDataProvider) PullAlertInstancesBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.AlertInstance, error] {
	oneWeek := 7 * 24 * time.Hour
	instances := p.makeFakeAlertInstances(time.Now().Add(-oneWeek), time.Now().Add(oneWeek))
	return func(yield func(*ent.AlertInstance, error) bool) {
		for _, inst := range instances {
			if !yield(inst, nil) {
				break
			}
		}
	}
}
