package grafana

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
)

type AlertsDataProvider struct {
	apiEndpoint   string
	apiToken      string
	webhookSecret string

	providerUserMap map[string]*ent.User

	webhookCallback rez.DataProviderResourceUpdatedCallback
}

type AlertsDataProviderConfig struct {
	ApiEndpoint   string `json:"api_endpoint"`
	ApiToken      string `json:"api_token"`
	WebhookSecret string `json:"webhook_secret"`
}

func NewAlertsDataProvider(cfg AlertsDataProviderConfig) (*AlertsDataProvider, error) {
	if cfg.WebhookSecret == "" {
		return nil, errors.New("no secret provided")
	}

	p := &AlertsDataProvider{
		apiEndpoint:     strings.TrimSuffix(cfg.ApiEndpoint, "/"),
		apiToken:        cfg.ApiToken,
		providerUserMap: make(map[string]*ent.User),
		webhookSecret:   cfg.WebhookSecret,
		webhookCallback: func(providerId string, updatedAt time.Time) {},
	}

	return p, nil
}

func (p *AlertsDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"alerts": p.makeAlertsWebhookHandler(p.webhookSecret),
	}
}

func (p *AlertsDataProvider) makeAlertsWebhookHandler(secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	})
}

func (p *AlertsDataProvider) SetOnAlertInstanceUpdatedCallback(cb rez.DataProviderResourceUpdatedCallback) {
	p.webhookCallback = cb
}

func (p *AlertsDataProvider) PullAlertInstancesBetweenDates(ctx context.Context, start, end time.Time) iter.Seq2[*ent.OncallAlertInstance, error] {
	params := &url.Values{}
	params.Set("started_at", fmt.Sprintf("%s_%s", formatOncallTime(start), formatOncallTime(end)))
	initialUrl := oncallApiUrl(p.apiEndpoint, "alert_groups", params)

	return func(yield func(*ent.OncallAlertInstance, error) bool) {
		reqUrl := &initialUrl
		for reqUrl != nil {
			var resp oncallPaginatedResponse[alertGroup]
			if getErr := oncallGet(ctx, *reqUrl, p.apiToken, &resp); getErr != nil {
				yield(nil, getErr)
				return
			}
			for _, group := range resp.Results {
				fmt.Printf("alert group: %+v\n", group)

				instance := &ent.OncallAlertInstance{
					AlertID:        uuid.UUID{},
					CreatedAt:      time.Time{},
					AckedAt:        time.Time{},
					ReceiverUserID: uuid.UUID{},
					Edges:          ent.OncallAlertInstanceEdges{},
				}

				if !yield(instance, nil) {
					return
				}
			}
			reqUrl = resp.Next
		}
	}
}
