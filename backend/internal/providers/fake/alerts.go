package fakeprovider

import (
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AlertsDataProvider struct {
	providerUserMap map[string]*ent.User

	webhookCallback rez.DataProviderResourceUpdatedCallback
}

type AlertsDataProviderConfig struct {
}

func NewAlertsDataProvider(cfg AlertsDataProviderConfig) (*AlertsDataProvider, error) {
	p := &AlertsDataProvider{
		providerUserMap: make(map[string]*ent.User),
		webhookCallback: func(providerId string, updatedAt time.Time) {},
	}

	return p, nil
}

func (p *AlertsDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{}
}

func (p *AlertsDataProvider) SetOnAlertInstanceUpdatedCallback(cb rez.DataProviderResourceUpdatedCallback) {
	p.webhookCallback = cb
}

/*
func (p *AlertsDataProvider) PullAlertInstancesBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallAlertInstance, error] {
	return func(yield func(*ent.OncallAlertInstance, error) bool) {

	}
}
*/
