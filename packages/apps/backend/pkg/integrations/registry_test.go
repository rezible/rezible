package integrations

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/test"
)

type RegistrySuite struct {
	test.Suite
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, &RegistrySuite{Suite: test.NewSuite()})
}

func (s *RegistrySuite) TestNewRegistryKeepsAvailableDefinitionsInOrder() {
	first := &fakeDefinition{name: "first"}
	unavailable := &fakeDefinition{name: "unavailable", unavailable: true}
	second := &fakeDefinition{name: "second"}

	reg, registryErr := NewRegistry(first, unavailable, second)

	s.Require().NoError(registryErr)
	s.Require().Equal([]rez.IntegrationDefinition{first, second}, reg.GetAvailable())

	_, getErr := reg.Get("unavailable")
	s.Require().ErrorIs(getErr, ErrUnknownIntegration)
}

func (s *RegistrySuite) TestNewRegistryRejectsAvailabilityErrors() {
	broken := &fakeDefinition{
		name:            "broken",
		availabilityErr: errors.New("missing config"),
	}

	_, registryErr := NewRegistry(broken)

	s.Require().ErrorContains(registryErr, "missing config")
}

func (s *RegistrySuite) TestNewRegistryRejectsDuplicateNames() {
	first := &fakeDefinition{name: "same"}
	second := &fakeDefinition{name: "same"}

	_, registryErr := NewRegistry(first, second)

	s.Require().ErrorContains(registryErr, "already registered")
}

func (s *RegistrySuite) TestRegistryLookup() {
	plain := &fakeDefinition{name: "plain"}
	webhook := &fakeWebhookDefinition{fakeDefinition: fakeDefinition{name: "webhook"}}
	reg, registryErr := NewRegistry(plain, webhook)
	s.Require().NoError(registryErr)

	found, foundErr := reg.Lookup[IntegrationWithWebhookHandler]("webhook")
	s.Require().NoError(foundErr)
	s.Require().Same(webhook, found)

	// The root errors let the API map these to 422 and 404.
	_, unsupportedErr := reg.Lookup[IntegrationWithWebhookHandler]("plain")
	s.Require().ErrorIs(unsupportedErr, ErrCapabilityNotSupported)
	s.Require().ErrorIs(unsupportedErr, errs.ErrUnprocessableInput)

	_, unknownErr := reg.Lookup[IntegrationWithWebhookHandler]("missing")
	s.Require().ErrorIs(unknownErr, ErrUnknownIntegration)
	s.Require().ErrorIs(unknownErr, errs.ErrNotFound)
}

func (s *RegistrySuite) TestRegistryAllReturnsImplementersInOrder() {
	firstWebhook := &fakeWebhookDefinition{fakeDefinition: fakeDefinition{name: "first"}}
	plain := &fakeDefinition{name: "plain"}
	secondWebhook := &fakeWebhookDefinition{fakeDefinition: fakeDefinition{name: "second"}}
	reg, registryErr := NewRegistry(firstWebhook, plain, secondWebhook)
	s.Require().NoError(registryErr)

	webhooks := reg.All[IntegrationWithWebhookHandler]()

	s.Require().Equal([]IntegrationWithWebhookHandler{firstWebhook, secondWebhook}, webhooks)
}

func (s *RegistrySuite) TestGetOAuth2FlowIntegrationRejectsMissingConfig() {
	unconfigured := &fakeOAuthDefinition{fakeDefinition: fakeDefinition{name: "unconfigured"}}
	configured := &fakeOAuthDefinition{
		fakeDefinition: fakeDefinition{name: "configured"},
		config:         &oauth2.Config{ClientID: "client"},
	}
	reg, registryErr := NewRegistry(unconfigured, configured)
	s.Require().NoError(registryErr)

	_, unconfiguredErr := reg.GetOAuth2FlowIntegration("unconfigured")
	s.Require().ErrorIs(unconfiguredErr, ErrCapabilityNotSupported)

	found, configuredErr := reg.GetOAuth2FlowIntegration("configured")
	s.Require().NoError(configuredErr)
	s.Require().Same(configured, found)
}

type fakeDefinition struct {
	name            string
	unavailable     bool
	availabilityErr error
}

func (d *fakeDefinition) Name() string {
	return d.name
}

func (d *fakeDefinition) Provider() string {
	return "fake"
}

func (d *fakeDefinition) DisplayName() string {
	return d.name
}

func (d *fakeDefinition) Description() string {
	return ""
}

func (d *fakeDefinition) Capabilities() []string {
	return nil
}

func (d *fakeDefinition) IsAvailable() (bool, error) {
	return !d.unavailable, d.availabilityErr
}

func (d *fakeDefinition) MaxInstalls() *int {
	return nil
}

func (d *fakeDefinition) OAuthInstallRequired() bool {
	return false
}

func (d *fakeDefinition) InstallationLinks() []rez.IntegrationInstallationLink {
	return nil
}

func (d *fakeDefinition) ValidateInstallationConfig([]byte) (rez.IntegrationInstallationConfig, error) {
	return nil, nil
}

func (d *fakeDefinition) ValidateUserSettings(map[string]any) error {
	return nil
}

func (d *fakeDefinition) GetInstalledIntegration(*ent.Integration) (rez.InstalledIntegration, error) {
	return nil, nil
}

type fakeWebhookDefinition struct {
	fakeDefinition
}

func (d *fakeWebhookDefinition) WebhookHandler() http.Handler {
	return http.NotFoundHandler()
}

type fakeOAuthDefinition struct {
	fakeDefinition
	config *oauth2.Config
}

func (d *fakeOAuthDefinition) OAuth2Config() *oauth2.Config {
	return d.config
}

func (d *fakeOAuthDefinition) RetrieveInstallationTargetOptions(context.Context, *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	return nil, nil
}
