package integrations

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

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

func (s *IntegrationsSuite) TestLookupLogSourceChoosesOneInstallation() {
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	oldest := newFakeLogReader(createdAt, "00000000-0000-0000-0000-000000000002", nil)
	tied := newFakeLogReader(createdAt, "00000000-0000-0000-0000-000000000003", nil)
	newer := newFakeLogReader(createdAt.Add(time.Hour), "00000000-0000-0000-0000-000000000001", nil)
	unavailable := newFakeLogReader(createdAt.Add(-time.Hour), "00000000-0000-0000-0000-000000000000", errors.New("no data source"))
	getter := fakeInstallationGetter{all: []rez.InstalledIntegration{newer, tied, unavailable, oldest, &fakeInstallation{}}}
	statedBy := func(stating ...*fakeLogReader) func(*ent.Integration) (bool, error) {
		return func(installation *ent.Integration) (bool, error) {
			return slices.ContainsFunc(stating, func(reader *fakeLogReader) bool { return reader.intg == installation }), nil
		}
	}

	s.Run("the oldest installation that states the service", func() {
		source, lookupErr := LookupLogSource(s.T().Context(), getter, statedBy(newer, tied))
		s.Require().NoError(lookupErr)
		s.Same(tied.intg, source.Installation)
	})

	s.Run("else the oldest that can read logs, ties ordered by ID", func() {
		source, lookupErr := LookupLogSource(s.T().Context(), getter, statedBy())
		s.Require().NoError(lookupErr)
		s.Same(oldest.intg, source.Installation)
	})
}

type fakeInstallationGetter struct {
	installed rez.InstalledIntegration
	all       []rez.InstalledIntegration
	err       error
}

func (g fakeInstallationGetter) GetInstalledIntegration(context.Context, uuid.UUID) (rez.InstalledIntegration, error) {
	return g.installed, g.err
}

func (g fakeInstallationGetter) ListAllInstalled(context.Context, ...predicate.Integration) ([]rez.InstalledIntegration, error) {
	return g.all, nil
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

// fakeLogReader is an installation that reads logs, unless unavailable says why it cannot.
type fakeLogReader struct {
	fakeInstallation
	intg        *ent.Integration
	unavailable error
}

func newFakeLogReader(createdAt time.Time, id string, unavailable error) *fakeLogReader {
	intg := &ent.Integration{ID: uuid.MustParse(id), Name: "fake", CreatedAt: createdAt}
	return &fakeLogReader{intg: intg, unavailable: unavailable}
}

func (r *fakeLogReader) Integration() *ent.Integration {
	return r.intg
}

func (r *fakeLogReader) SearchLogs(context.Context, LogSearchParams) (*LogSearchResult, error) {
	return &LogSearchResult{}, nil
}

func (r *fakeLogReader) CountLogs(context.Context, LogCountParams) (*LogCountResult, error) {
	return &LogCountResult{}, nil
}

func (r *fakeLogReader) CheckLogsAvailable() error {
	return r.unavailable
}

type fakeDirectMessageSender struct {
	fakeInstallation
}

func (ii *fakeDirectMessageSender) SendDirectMessage(context.Context, *ent.User, *DirectMessage) error {
	return nil
}

func (ii *fakeDirectMessageSender) DirectMessageDescription() string {
	return "a chat direct message"
}

func (s *IntegrationsSuite) TestDirectMessageSendersDescription() {
	unresolved := []error{errors.New("invalid installation config")}
	cases := map[string]struct {
		senders     DirectMessageSenders
		description string
		known       bool
	}{
		"a sender":                        {senders: DirectMessageSenders{Senders: []DirectMessageSender{&fakeDirectMessageSender{}}}, description: "a chat direct message", known: true},
		"a sender and an unresolved one":  {senders: DirectMessageSenders{Senders: []DirectMessageSender{&fakeDirectMessageSender{}}, Unresolved: unresolved}, description: "a chat direct message", known: true},
		"no sender":                       {senders: DirectMessageSenders{}, known: true},
		"only an unresolved installation": {senders: DirectMessageSenders{Unresolved: unresolved}, known: false},
	}
	for name, tc := range cases {
		description, known := tc.senders.Description()
		s.Equal(tc.description, description, name)
		s.Equal(tc.known, known, name)
	}
}
