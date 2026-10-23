package grafana

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/pkg/errs"
)

func TestValidateUserSettings(t *testing.T) {
	integration, makeErr := MakeIntegration(nil)
	require.NoError(t, makeErr)

	for _, tc := range []struct {
		name     string
		settings string
		contains string
	}{
		{
			name:     "a non-string value",
			settings: `{"logs_data_source_uid": 12}`,
			contains: "logs_data_source_uid",
		},
		{
			name:     "a label that would widen a selector",
			settings: `{"log_service_label": "service_name=~\".+\",x"}`,
			contains: "log_service_label",
		},
		{
			name:     "a UID with a slash",
			settings: `{"metrics_data_source_uid": "prometheus/../admin"}`,
			contains: "metrics_data_source_uid",
		},
		{
			name:     "a UID longer than 40 characters",
			settings: `{"logs_data_source_uid": "` + strings.Repeat("a", 41) + `"}`,
			contains: "logs_data_source_uid",
		},
		{
			name:     "an unknown setting",
			settings: `{"logs_datasource_uid": "loki"}`,
			contains: "logs_datasource_uid",
		},
	} {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			validateErr := integration.ValidateUserSettings([]byte(tc.settings))

			require.ErrorIs(t, validateErr, errs.ErrInvalidInput)
			require.ErrorContains(t, validateErr, tc.contains)
		})
	}

	t.Run("accepts empty values", func(t *testing.T) {
		require.NoError(t, integration.ValidateUserSettings(nil))

		empty := `{"logs_data_source_uid": "", "metrics_data_source_uid": "", "log_service_label": "", "metric_service_label": ""}`
		require.NoError(t, integration.ValidateUserSettings([]byte(empty)))
	})

	t.Run("accepts valid values", func(t *testing.T) {
		valid := `{"logs_data_source_uid": "loki", "metrics_data_source_uid": "P1809F7CD0C75ACF3", "log_service_label": "service_name", "metric_service_label": "_app"}`
		require.NoError(t, integration.ValidateUserSettings([]byte(valid)))
	})
}

func TestInstallationConfig(t *testing.T) {
	integration, makeErr := MakeIntegration(nil)
	require.NoError(t, makeErr)

	t.Run("the URL is normalised and identifies the installation", func(t *testing.T) {
		raw := []byte(`{"url":" https://grafana.example.com/grafana/ ","token":"glsa_secret"}`)

		config, configErr := integration.ValidateInstallationConfig(raw)

		require.NoError(t, configErr)
		require.Equal(t, "https://grafana.example.com/grafana", config.InstallationTargetRef().ResourceRef)
	})

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"a missing token", `{"url":"https://grafana.example.com"}`},
		{"a missing URL", `{"token":"glsa_secret"}`},
		{"a URL that is not http", `{"url":"ftp://grafana.example.com","token":"glsa_secret"}`},
		{"a URL with credentials", `{"url":"https://admin:pw@grafana.example.com","token":"glsa_secret"}`},
	} {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			_, configErr := integration.ValidateInstallationConfig([]byte(tc.raw))

			require.ErrorIs(t, configErr, errs.ErrInvalidInput)
			require.NotContains(t, configErr.Error(), "glsa_secret")
		})
	}
}
