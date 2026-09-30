package slackincidents

import (
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/stretchr/testify/require"
)

func TestCheckInstallRequirements(t *testing.T) {
	enabledPrefs := &ent.OrganizationPreferences{
		EnableIncidentManagement: true,
	}
	disabledPrefs := &ent.OrganizationPreferences{
		EnableIncidentManagement: false,
	}
	chatIntegration := &fakeInstalledIntegration{
		capabilities: []string{"chat_context"},
	}
	incidentIntegration := &fakeInstalledIntegration{
		capabilities: []string{incidentManagementCapability},
	}

	testCases := []struct {
		name        string
		state       *integrations.IntegrationInstallState
		installable bool
	}{
		{
			name:        "missing preferences",
			state:       &integrations.IntegrationInstallState{},
			installable: false,
		},
		{
			name: "incident management disabled",
			state: &integrations.IntegrationInstallState{
				Preferences: disabledPrefs,
			},
			installable: false,
		},
		{
			name: "incident management enabled without a provider",
			state: &integrations.IntegrationInstallState{
				Preferences: enabledPrefs,
				Installed:   []rez.InstalledIntegration{chatIntegration},
			},
			installable: true,
		},
		{
			name: "incident management provider already installed",
			state: &integrations.IntegrationInstallState{
				Preferences: enabledPrefs,
				Installed:   []rez.InstalledIntegration{chatIntegration, incidentIntegration},
			},
			installable: false,
		},
	}

	intg := &Integration{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requirementsErr := intg.CheckInstallRequirements(tc.state)

			if tc.installable {
				require.NoError(t, requirementsErr)
			} else {
				require.Error(t, requirementsErr)
			}
		})
	}
}

type fakeInstalledIntegration struct {
	capabilities []string
}

func (f *fakeInstalledIntegration) Integration() *ent.Integration {
	return &ent.Integration{}
}

func (f *fakeInstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return nil
}

func (f *fakeInstalledIntegration) Capabilities() []string {
	return f.capabilities
}
