package webhook

import (
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

// preset is a payload format Rezible defines. An installation accepts one preset, named in its config. A
// preset's provider events use its name as their event source.
type preset interface {
	// MapDelivery validates one delivery and returns its provider event, or a fieldError for a 400.
	MapDelivery(integrationID uuid.UUID, body []byte, receivedAt time.Time) (*rez.ProviderEvent, error)
	// ProcessEvent normalizes one of the preset's provider events.
	ProcessEvent(ev rez.ProviderEvent) (ent.NormalizedEvents, error)
}

const presetDeployment = "deployment"

var presets = map[string]preset{
	presetDeployment: deploymentPreset{},
}

// fieldError rejects a delivery, naming the field so a pipeline log shows what to fix.
type fieldError struct {
	field  string
	reason string
}

func (e *fieldError) Error() string {
	return e.field + ": " + e.reason
}
