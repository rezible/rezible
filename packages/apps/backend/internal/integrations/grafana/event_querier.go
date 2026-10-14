package grafana

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"slices"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
)

const (
	// SourceServices lists the services Grafana observes, by the values of the installation's service labels.
	SourceServices = "services"

	// ServiceLabelWindow is how far back a listing reads service label values.
	ServiceLabelWindow = 24 * time.Hour
	// MaxServiceLabelValues is how many values, in sorted order, a listing keeps.
	MaxServiceLabelValues = 500

	signalLogs = "logs"
)

var (
	_ integrations.IntegrationWithProviderEventQuerier = (*Integration)(nil)
	_ integrations.IntegrationWithScheduledSync        = (*Integration)(nil)
)

// ScheduledSyncSources lists services hourly, so services that start sending logs are observed.
func (i *Integration) ScheduledSyncSources() []string {
	return []string{SourceServices}
}

// serviceListing is a services event's attributes: the values of one signal's service label.
type serviceListing struct {
	Signal string   `json:"signal"`
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

func (l serviceListing) eventRef() string {
	encoded, encodeErr := json.Marshal(l.Values)
	if encodeErr != nil {
		// A string slice always encodes.
		panic(fmt.Sprintf("encode service label values: %v", encodeErr))
	}
	sum := sha256.Sum256(encoded)
	return l.Signal + ":" + l.Label + ":" + hex.EncodeToString(sum[:])
}

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	installed, installedErr := i.GetInstalledIntegration(intg)
	if installedErr != nil {
		return nil, installedErr
	}
	return &eventQuerier{ii: installed.(*InstalledIntegration)}, nil
}

type eventQuerier struct {
	ii *InstalledIntegration
}

// QueryProviderEvents lists services whatever cursors it is given, so every sync reads them.
func (q *eventQuerier) QueryProviderEvents(ctx context.Context, _ rez.ProviderEventSourceCursors) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		ds, dsErr := q.ii.logsDataSource()
		if dsErr != nil {
			// Without a logs data source there is nothing to list.
			return
		}
		listing, listErr := q.ii.listLogServices(ctx, ds)
		if listErr != nil {
			yield(nil, listErr)
			return
		}
		attributes, encodeErr := json.Marshal(listing)
		if encodeErr != nil {
			yield(nil, fmt.Errorf("encode service listing: %w", encodeErr))
			return
		}
		event := rez.ProviderEvent{
			Provider:            ProviderName,
			ProviderNamespace:   q.ii.intg.ID.String(),
			ProviderEventSource: SourceServices,
			ProviderEventRef:    listing.eventRef(),
			Attributes:          attributes,
			ReceivedAt:          q.ii.clock.Now(),
		}
		// The sync stops at the first result without a cursor.
		yield(&rez.ProviderEventQueryResult{Event: event, ProviderEventSourceCursorAfter: new(listing.Signal)}, nil)
	}
}

// listLogServices lists the values of the log service label over the last ServiceLabelWindow, sorted, keeping
// at most MaxServiceLabelValues.
func (ii *InstalledIntegration) listLogServices(ctx context.Context, ds dataSource) (*serviceListing, error) {
	label := ii.settings.LogServiceLabel
	end := ii.clock.Now()
	params := url.Values{
		"start": {formatTime(end.Add(-ServiceLabelWindow))},
		"end":   {formatTime(end)},
	}
	var values []string
	readErr := integrations.MeasureRead(ctx, ProviderName, "list_service_label_values", func(ctx context.Context) (int, error) {
		if getErr := ii.client.get(ctx, ds, "/loki/api/v1/label/"+label+"/values", params, &values); getErr != nil {
			return 0, fmt.Errorf("list service label values: %w", getErr)
		}
		return len(values), nil
	})
	if readErr != nil {
		return nil, readErr
	}
	if values == nil {
		values = []string{}
	}
	slices.Sort(values)
	if len(values) > MaxServiceLabelValues {
		slog.WarnContext(ctx, "dropped grafana service label values past the limit",
			"integration_id", ii.intg.ID,
			"dropped", len(values)-MaxServiceLabelValues,
			"first_dropped", values[MaxServiceLabelValues],
		)
		values = values[:MaxServiceLabelValues]
	}
	listing := &serviceListing{
		Signal: signalLogs,
		Label:  label,
		Values: values,
	}
	return listing, nil
}
