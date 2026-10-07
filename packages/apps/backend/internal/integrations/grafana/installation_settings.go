package grafana

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"

	rez "github.com/rezible/rezible"
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
		return field.Tag.Get("mapstructure")
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
	LogsDataSourceUID    string `mapstructure:"logs_data_source_uid" validate:"omitempty,grafanauid"`
	MetricsDataSourceUID string `mapstructure:"metrics_data_source_uid" validate:"omitempty,grafanauid"`
	LogServiceLabel      string `mapstructure:"log_service_label" validate:"omitempty,labelname"`
	MetricServiceLabel   string `mapstructure:"metric_service_label" validate:"omitempty,labelname"`
}

func parseInstallationSettings(raw map[string]any) (*installationSettings, error) {
	var settings installationSettings
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &settings,
	}
	decoder, decoderErr := mapstructure.NewDecoder(decoderConfig)
	if decoderErr != nil {
		return nil, fmt.Errorf("settings decoder: %w", decoderErr)
	}
	if decodeErr := decoder.Decode(raw); decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid settings: %w", rez.ErrInvalidInput, decodeErr)
	}
	if validateErr := settingsValidator.Struct(settings); validateErr != nil {
		var fieldErrs validator.ValidationErrors
		if errors.As(validateErr, &fieldErrs) && len(fieldErrs) > 0 {
			fieldErr := fieldErrs[0]
			return nil, fmt.Errorf("%w: %s %q is not %s", rez.ErrInvalidInput,
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
	return &settings, nil
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
