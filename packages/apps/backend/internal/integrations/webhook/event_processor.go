package webhook

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

// EventProcessor normalizes webhook provider events with the preset named by their event source. It has no
// dependencies.
type EventProcessor struct{}

func (p EventProcessor) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	preset, known := presets[prov.ProviderEventSource]
	if !known {
		return nil, fmt.Errorf("unknown provider event source: %s", prov.ProviderEventSource)
	}
	return preset.ProcessEvent(prov)
}
