package alertmanager

import (
	"encoding/json"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
)

const settingServiceLabels = "service_labels"

// defaultServiceLabels may name a service when the installation has no service_labels setting. `job` is
// deliberately absent: in Prometheus it usually names a scrape job, not a service.
var defaultServiceLabels = []string{"service", "service_name", "app", "app_kubernetes_io_name"}

type installationSettings struct {
	// serviceLabels are tried in order; the first value that normalizes to a nonempty name wins. Empty
	// disables service attachment.
	serviceLabels []string
}

// storedSettings are the settings as an admin stores them; a nil ServiceLabels is an absent setting.
type storedSettings struct {
	ServiceLabels json.RawMessage `json:"service_labels"`
}

func parseInstallationSettings(settings []byte) (*installationSettings, error) {
	stored, decodeErr := integrations.DecodeUserSettings[storedSettings](settings)
	if decodeErr != nil {
		return nil, decodeErr
	}
	if stored.ServiceLabels == nil {
		return &installationSettings{serviceLabels: defaultServiceLabels}, nil
	}
	var labels []string
	if unmarshalErr := json.Unmarshal(stored.ServiceLabels, &labels); unmarshalErr != nil || labels == nil {
		return nil, fmt.Errorf("%w: %s must be a list of label names", errs.ErrInvalidInput, settingServiceLabels)
	}
	seen := mapset.NewThreadUnsafeSet[string]()
	for i, label := range labels {
		if strings.TrimSpace(label) == "" {
			return nil, fmt.Errorf("%w: %s[%d] must not be empty", errs.ErrInvalidInput, settingServiceLabels, i)
		}
		if !seen.Add(label) {
			return nil, fmt.Errorf("%w: %s contains %q more than once", errs.ErrInvalidInput, settingServiceLabels, label)
		}
	}
	return &installationSettings{serviceLabels: labels}, nil
}
