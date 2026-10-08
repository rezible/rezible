package integrations

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
)

func TestAs(t *testing.T) {
	chat := &fakeChatInstallation{}
	querier, supportedErr := As[ChatChannelQuerier](chat)
	require.NoError(t, supportedErr)
	require.Same(t, chat, querier)

	plain := &fakeInstallation{}
	_, unsupportedErr := As[ChatChannelQuerier](plain)
	require.ErrorIs(t, unsupportedErr, ErrCapabilityNotSupported)
}

func TestLookupInstallationAs(t *testing.T) {
	chat := &fakeChatInstallation{}
	found := fakeInstallationGetter{installed: chat}
	querier, foundErr := LookupInstallationAs[ChatChannelQuerier](t.Context(), found, uuid.New())
	require.NoError(t, foundErr)
	require.Same(t, chat, querier)

	missing := fakeInstallationGetter{err: errs.ErrNotFound}
	_, missingErr := LookupInstallationAs[ChatChannelQuerier](t.Context(), missing, uuid.New())
	require.ErrorIs(t, missingErr, errs.ErrNotFound)
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
