package alertmanager

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	rez "github.com/rezible/rezible"
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

func parseInstallationSettings(settings map[string]any) (*installationSettings, error) {
	raw, present := settings[settingServiceLabels]
	if !present {
		return &installationSettings{serviceLabels: defaultServiceLabels}, nil
	}
	var values []any
	switch typed := raw.(type) {
	case []any:
		values = typed
	case []string:
		for _, label := range typed {
			values = append(values, label)
		}
	default:
		return nil, fmt.Errorf("%w: %s must be a list of label names", rez.ErrInvalidInput, settingServiceLabels)
	}
	labels := make([]string, 0, len(values))
	seen := mapset.NewThreadUnsafeSet[string]()
	for i, value := range values {
		label, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("%w: %s[%d] must be a string", rez.ErrInvalidInput, settingServiceLabels, i)
		}
		if strings.TrimSpace(label) == "" {
			return nil, fmt.Errorf("%w: %s[%d] must not be empty", rez.ErrInvalidInput, settingServiceLabels, i)
		}
		if !seen.Add(label) {
			return nil, fmt.Errorf("%w: %s contains %q more than once", rez.ErrInvalidInput, settingServiceLabels, label)
		}
		labels = append(labels, label)
	}
	return &installationSettings{serviceLabels: labels}, nil
}
