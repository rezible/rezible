package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func (i *integration) MakeProviderEventProcessor() rez.ProviderEventProcessor {
	return &eventProcessor{}
}

type eventProcessor struct {
}

const (
	sourceAlerts    = "alerts"
	sourceIncidents = "incidents"
)

func (p *eventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourceAlerts:
		return p.processAlert(prov)
	case sourceIncidents:
		return p.processIncident(prov)
	default:
		return nil, fmt.Errorf("unknown provider source: %s", prov.ProviderSource)
	}
}

func (p *eventProcessor) processAlert(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload alertFiredPayload
	if jsonErr := json.Unmarshal(prov.Payload, &payload); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal alertFiredPayload: %w", jsonErr)
	}

	//attrs := eventprojections.UserObservedAttributes{}

	result := &ent.NormalizedEvent{
		Provider:       integrationName,
		ProviderSource: sourceAlerts,
		//Kind:             ne.KindAlert,
		SubjectKind:      "alert",
		SubjectRef:       prov.SubjectRef,
		ProviderEventRef: prov.ProviderEventRef,
		OccurredAt:       time.Now(),
		ReceivedAt:       prov.ReceivedAt,
		//Attributes:       attrs.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processIncident(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {

	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceIncidents,
		ProviderEventRef: prov.ProviderEventRef,
		//Kind:             ne.KindIncident,
		SubjectKind: "incident",
		SubjectRef:  prov.SubjectRef,
		OccurredAt:  time.Now(),
		ReceivedAt:  prov.ReceivedAt,
		//Attributes:       attrs.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}
