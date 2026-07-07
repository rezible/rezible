package github

import (
	"encoding/json"
	"testing"

	"github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestIntegration() *Integration {
	cfg := rez.IntegrationsConfigGithub{}
	cfg.App.AppID = 123
	cfg.App.ClientID = "client-id"
	cfg.App.ClientSecret = "client-secret"
	intg := &Integration{cfg: cfg}
	return intg
}

func TestOAuth2Config(t *testing.T) {
	intg := makeTestIntegration()

	oauthCfg := intg.OAuth2Config()
	require.NotNil(t, oauthCfg)
	assert.Equal(t, "client-id", oauthCfg.ClientID)
	assert.Equal(t, "client-secret", oauthCfg.ClientSecret)
	assert.Equal(t, "https://github.com/login/oauth/authorize", oauthCfg.Endpoint.AuthURL)
	assert.Equal(t, "https://github.com/login/oauth/access_token", oauthCfg.Endpoint.TokenURL)

	authURL := oauthCfg.AuthCodeURL("state-value")
	assert.Contains(t, authURL, "client_id=client-id")
	assert.Contains(t, authURL, "state=state-value")
}

func TestEncodeDecodeInstallationConfig(t *testing.T) {
	org := "foobar"
	instId := int64(1)
	cfg := &InstallationConfig{Org: org, InstallationID: instId}
	expectedJson, jsonErr := json.Marshal(cfg)
	require.NoError(t, jsonErr)
	enc, encErr := cfg.Encode()
	require.NoError(t, encErr)
	assert.Equal(t, expectedJson, enc)
}

func TestExtractIntegrationOptionsFromToken(t *testing.T) {
	intg := makeTestIntegration()

	installations := []*github.Installation{
		{
			ID:      new(int64(456)),
			AppID:   new(int64(123)),
			Account: &github.User{Login: new("myorg")},
		},
	}

	options, err := intg.makeInstallationTargetOptions(installations)
	require.NoError(t, err)
	require.Len(t, options, 1)
	assert.Equal(t, "456", options[0].Config.ExternalRef())
	assert.Equal(t, "myorg", options[0].DisplayName)
}

func TestExtractIntegrationOptionsFromToken_NoInstallations(t *testing.T) {
	intg := &Integration{}

	_, err := intg.makeInstallationTargetOptions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid github app installations")
}

func TestExtractIntegrationOptionsFromToken_MultipleInstallations(t *testing.T) {
	intg := &Integration{}
	installations := []*github.Installation{
		{ID: new(int64(1)), Account: &github.User{Login: new("org-one")}},
		{ID: new(int64(2)), Account: &github.User{Login: new("org-two")}},
	}
	options, err := intg.makeInstallationTargetOptions(installations)
	require.NoError(t, err)
	require.Len(t, options, 2)
	assert.Equal(t, "1", options[0].Config.ExternalRef())
	assert.Equal(t, "2", options[1].Config.ExternalRef())
}
