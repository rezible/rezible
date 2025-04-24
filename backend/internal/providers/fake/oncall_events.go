package fakeprovider

import (
	"context"
	"iter"
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

func (p *OncallEventsDataProvider) PullEventsBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallEvent, error] {
	return func(yield func(*ent.OncallEvent, error) bool) {

	}
}
