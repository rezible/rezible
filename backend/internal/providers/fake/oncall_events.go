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

func makeFakeDayEvent(date time.Time) *ent.OncallEvent {
	isAlert := rand.Float64() > 0.25
	eventKind := "incident"
	if isAlert {
		eventKind = "alert"
	}

	hour := rand.Intn(24)
	minute := rand.Intn(60)

	timestamp := time.Date(
		date.Year(), date.Month(), date.Day(),
		hour, minute, 0, 0, date.Location(),
	)

	id := uuid.New()

	return &ent.OncallEvent{
		ID:          id,
		ProviderID:  id.String(),
		Timestamp:   timestamp,
		Source:      "fake",
		Kind:        eventKind,
		Title:       "title",
		Description: "fake description",
	}
}

func makeFakeOncallEvents(start, end time.Time) []*ent.OncallEvent {
	numHours := end.Sub(start).Hours()
	if numHours <= 0 {
		return nil
	}
	numDays := int(numHours / 24)
	maxDailyEvents := 10
	events := make([]*ent.OncallEvent, 0, numDays*maxDailyEvents)

	for day := 0; day < numDays; day++ {
		for i := 0; i < rand.Intn(maxDailyEvents); i++ {
			events = append(events, makeFakeDayEvent(start.AddDate(0, 0, day)))
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
