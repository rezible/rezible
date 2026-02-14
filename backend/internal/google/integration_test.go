package google

import (
	"testing"

	"github.com/rezible/rezible/ent"
	"github.com/stretchr/testify/require"
)

func TestConfiguredIntegrationIsVideoConferenceEnabled(t *testing.T) {
	baseCfg := []byte(`{"UserConfig":{"ServiceAccountCredentials":{"client_email":"x@y.z"}}}`)

	testCases := []struct {
		name     string
		cfg      []byte
		prefs    map[string]any
		expected bool
	}{
		{
			name:     "default enabled with service account",
			cfg:      baseCfg,
			expected: true,
		},
		{
			name:     "boolean false preference",
			cfg:      baseCfg,
			prefs:    map[string]any{PreferenceEnableIncidentVideoConferences: false},
			expected: false,
		},
		{
			name:     "string false preference",
			cfg:      baseCfg,
			prefs:    map[string]any{PreferenceEnableIncidentVideoConferences: "false"},
			expected: false,
		},
		{
			name:     "boolean true preference",
			cfg:      baseCfg,
			prefs:    map[string]any{PreferenceEnableIncidentVideoConferences: true},
			expected: true,
		},
		{
			name:     "missing service account disables",
			cfg:      []byte(`{"UserConfig":{}}`),
			prefs:    map[string]any{PreferenceEnableIncidentVideoConferences: true},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ci := &ConfiguredIntegration{
				intg: &ent.Integration{
					Config:          tc.cfg,
					UserPreferences: tc.prefs,
				},
			}
			require.Equal(t, tc.expected, ci.isVideoConferenceEnabled())
		})
	}
}
