package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func (i *integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	return newEventQuerier(&ConfiguredIntegration{intg: intg}), nil
}

type eventQuerier struct {
	ci *ConfiguredIntegration
}

func newEventQuerier(ci *ConfiguredIntegration) *eventQuerier {
	return &eventQuerier{ci: ci}
}

func (q *eventQuerier) Provider() string {
	return integrationName
}

func (q *eventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if alertsCursor, ok := req.SourceCursors[sourceAlerts]; ok || len(req.SourceCursors) == 0 {
			for ev, evErr := range q.pullAlertEvents(ctx, alertsCursor) {
				yield(ev, evErr)
			}
		}
	}
}

type alertFiredPayload struct {
	OccurredAt time.Time `json:"occurred_at"`
}

func (q *eventQuerier) pullAlertEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		payload := alertFiredPayload{
			OccurredAt: time.Now(),
		}
		payloadBytes, jsonErr := json.Marshal(payload)
		if jsonErr != nil {
			yield(nil, fmt.Errorf("json marshal err: %w", jsonErr))
			return
		}
		res := &rez.ProviderEventQueryResult{
			Event: rez.ProviderEvent{
				Provider:         integrationName,
				ProviderSource:   sourceAlerts,
				ProviderEventRef: "",
				SubjectRef:       "",
				ReceivedAt:       time.Now(),
				Payload:          payloadBytes,
				ContentType:      "application/json",
			},
		}

		if !yield(res, nil) {
			return
		}
	}
}
