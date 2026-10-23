package integrations

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/test"
)

type IntegrationsSuite struct {
	test.Suite
}

func TestIntegrationsSuite(t *testing.T) {
	suite.Run(t, &IntegrationsSuite{Suite: test.NewSuite()})
}

func (s *IntegrationsSuite) TestAs() {
	chat := &fakeChatInstallation{}
	querier, supportedErr := As[ChatChannelQuerier](chat)
	s.Require().NoError(supportedErr)
	s.Same(chat, querier)

	plain := &fakeInstallation{}
	_, unsupportedErr := As[ChatChannelQuerier](plain)
	s.ErrorIs(unsupportedErr, ErrCapabilityNotSupported)
}

func (s *IntegrationsSuite) TestLookupInstallationAs() {
	chat := &fakeChatInstallation{}
	found := fakeInstallationGetter{installed: chat}
	querier, foundErr := LookupInstallationAs[ChatChannelQuerier](s.T().Context(), found, uuid.New())
	s.Require().NoError(foundErr)
	s.Same(chat, querier)

	missing := fakeInstallationGetter{err: errs.ErrNotFound}
	_, missingErr := LookupInstallationAs[ChatChannelQuerier](s.T().Context(), missing, uuid.New())
	s.ErrorIs(missingErr, errs.ErrNotFound)
}

type fakeInstallationGetter struct {
	installed rez.InstalledIntegration
	err       error
}

func (g fakeInstallationGetter) GetInstalledIntegration(context.Context, uuid.UUID) (rez.InstalledIntegration, error) {
	return g.installed, g.err
}

func (g fakeInstallationGetter) ListAllInstalled(context.Context, ...predicate.Integration) ([]rez.InstalledIntegration, error) {
	return nil, nil
}

type fakeInstallation struct{}

func (ii *fakeInstallation) Integration() *ent.Integration {
	return &ent.Integration{Name: "fake"}
}

func (ii *fakeInstallation) Config() rez.IntegrationInstallationConfig {
	return nil
}

func (ii *fakeInstallation) Capabilities() []string {
	return nil
}

type fakeChatInstallation struct {
	fakeInstallation
}

func (ii *fakeChatInstallation) ListChatChannels(context.Context, ListChatChannelsParams) (*ChatChannelPage, error) {
	return &ChatChannelPage{}, nil
}
