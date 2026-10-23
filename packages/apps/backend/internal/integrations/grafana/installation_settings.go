package grafana

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
)

const defaultServiceLabel = "service_name"

// Labels and UIDs are inserted into queries and proxy paths, so anything else is rejected.
var (
	dataSourceUIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,40}$`)
	labelNamePattern     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

var settingRuleDescriptions = map[string]string{
	"grafanauid": "a Grafana data source UID: 1 to 40 letters, digits, '-' or '_'",
	"labelname":  "a label name: letters, digits and '_', not starting with a digit",
}

var settingsValidator = newSettingsValidator()

func newSettingsValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("json")
	})
	rules := map[string]*regexp.Regexp{
		"grafanauid": dataSourceUIDPattern,
		"labelname":  labelNamePattern,
	}
	for tag, pattern := range rules {
		matches := func(fl validator.FieldLevel) bool {
			return pattern.MatchString(fl.Field().String())
		}
		if registerErr := v.RegisterValidation(tag, matches); registerErr != nil {
			panic(registerErr)
		}
	}
	return v
}

type installationSettings struct {
	LogsDataSourceUID    string `json:"logs_data_source_uid" validate:"omitempty,grafanauid"`
	MetricsDataSourceUID string `json:"metrics_data_source_uid" validate:"omitempty,grafanauid"`
	LogServiceLabel      string `json:"log_service_label" validate:"omitempty,labelname"`
	MetricServiceLabel   string `json:"metric_service_label" validate:"omitempty,labelname"`
}

func parseInstallationSettings(raw []byte) (*installationSettings, error) {
	settings, decodeErr := integrations.DecodeUserSettings[installationSettings](raw)
	if decodeErr != nil {
		return nil, decodeErr
	}
	if validateErr := settingsValidator.Struct(settings); validateErr != nil {
		var fieldErrs validator.ValidationErrors
		if errors.As(validateErr, &fieldErrs) && len(fieldErrs) > 0 {
			fieldErr := fieldErrs[0]
			return nil, fmt.Errorf("%w: %s %q is not %s", errs.ErrInvalidInput,
				fieldErr.Field(), fieldErr.Value(), settingRuleDescriptions[fieldErr.Tag()])
		}
		return nil, fmt.Errorf("validate settings: %w", validateErr)
	}
	if settings.LogServiceLabel == "" {
		settings.LogServiceLabel = defaultServiceLabel
	}
	if settings.MetricServiceLabel == "" {
		settings.MetricServiceLabel = defaultServiceLabel
	}
	return settings, nil
}

// logServiceName chooses the value a service's logs carry under the current log service label: the value in
// the service's logs alias for that label (see serviceResourceRef), else its display name. An alias under a
// label the installation no longer uses is ignored. Several values under the current label pick the lowest.
func (s *installationSettings) logServiceName(service integrations.LogService) integrations.LogServiceName {
	prefix := serviceResourceRef(signalLogs, s.LogServiceLabel, "")
	var values []string
	for _, alias := range service.Aliases {
		if value, found := strings.CutPrefix(alias, prefix); found && value != "" {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return integrations.LogServiceName{Value: service.DisplayName}
	}
	return integrations.LogServiceName{Value: slices.Min(values), FromAlias: true}
}

// logSelector selects the lines of one service, optionally only those containing some text. Values are
// quoted, so a caller cannot widen the selector.
func (s *installationSettings) logSelector(service, contains string) string {
	selector := "{" + s.LogServiceLabel + "=" + strconv.Quote(service) + "}"
	if contains != "" {
		selector += " |= " + strconv.Quote(contains)
	}
	return selector
}

func (s *installationSettings) metricSelector(service string) string {
	return "{" + s.MetricServiceLabel + "=" + strconv.Quote(service) + "}"
}
