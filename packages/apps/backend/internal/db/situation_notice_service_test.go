package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	sitnot "github.com/rezible/rezible/ent/situationnotice"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

var noticeTestOpenedAt = time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC)

type SituationNoticeServiceSuite struct {
	test.Suite
}

func TestSituationNoticeServiceSuite(t *testing.T) {
	suite.Run(t, &SituationNoticeServiceSuite{Suite: test.NewSuite()})
}

func (s *SituationNoticeServiceSuite) TestBuildFailure() {
	ctx, tdb := s.SetupTestDatabase()
	now := noticeTestOpenedAt.Add(time.Hour)
	situation := s.createRaisedSituation(ctx, tdb)

	// One notice has not been posted yet, the other was posted before its content could no longer be built.
	unposted := s.createNotice(ctx, tdb, situation.ID, "one", func(*ent.SituationNoticeCreate) {})
	posted := s.createNotice(ctx, tdb, situation.ID, "two", func(create *ent.SituationNoticeCreate) {
		create.
			SetStatus(sitnot.StatusPosted).
			SetMessageRef("message-1").
			SetPermalink("https://chat.test/message-1").
			SetFirstPostedAt(noticeTestOpenedAt)
	})

	briefs := mocks.NewMockSituationBriefService(s.T())
	briefs.EXPECT().
		ListBriefServices(mock.Anything, situation.ID).
		Return(nil, errors.New("brief services unavailable"))
	installations := mocks.NewMockIntegrationService(s.T())
	installations.EXPECT().
		ListInstallations(mock.Anything).
		Return(nil, nil)
	installations.EXPECT().
		AsInstalledIntegration(mock.Anything).
		RunAndReturn(func(installation *ent.Integration) (rez.InstalledIntegration, error) {
			return &noticeSender{installation: installation}, nil
		})
	notices := NewSituationNoticeService(tdb, s.Config(), test.NewClock(now), nil, nil, briefs, nil, installations, &OrganizationService{db: tdb})

	s.Run("an attempt that is not the last stores nothing", func() {
		syncErr := notices.sync(ctx, situation.ID, false)

		s.Require().Error(syncErr)
		s.Equal(sitnot.StatusPending, s.reloadNotice(ctx, tdb, unposted).Status)
		s.Equal(sitnot.StatusPosted, s.reloadNotice(ctx, tdb, posted).Status)
	})

	s.Run("the last attempt fails every notice, keeping a posted one's message", func() {
		syncErr := notices.sync(ctx, situation.ID, true)

		s.Require().Error(syncErr)
		for _, notice := range []*ent.SituationNotice{unposted, posted} {
			failed := s.reloadNotice(ctx, tdb, notice)
			s.Equal(sitnot.StatusFailed, failed.Status)
			s.Equal("could not build the notice", failed.Error)
			s.Equal(now, failed.LastAttemptAt.UTC())
			s.Equal(notice.MessageRef, failed.MessageRef)
			s.Equal(notice.Permalink, failed.Permalink)
		}
	})
}

func (s *SituationNoticeServiceSuite) TestFailingDeploymentLookupOmitsDeployment() {
	ctx, tdb := s.SetupTestDatabase()
	situation := s.createRaisedSituation(ctx, tdb)
	checkout := &ent.KnowledgeEntity{ID: uuid.New()}

	briefs := mocks.NewMockSituationBriefService(s.T())
	briefs.EXPECT().
		ListBriefServices(mock.Anything, situation.ID).
		Return([]*ent.KnowledgeEntity{checkout}, nil)
	briefs.EXPECT().
		LatestDeploymentBefore(mock.Anything, situation.ID).
		Return(nil, errors.New("changes lookup failed"))
	installations := mocks.NewMockIntegrationService(s.T())
	installations.EXPECT().
		ListAllInstalled(mock.Anything).
		Return([]rez.InstalledIntegration{&noticeSender{capabilities: []string{integrations.CapabilityDeployments}}}, nil)
	notices := NewSituationNoticeService(tdb, s.Config(), test.NewClock(noticeTestOpenedAt), nil, nil, briefs, nil, installations, &OrganizationService{db: tdb})

	content, readErr := notices.readContent(ctx, situation.ID)

	s.Require().NoError(readErr, "the notice is still sent")
	s.Equal([]*ent.KnowledgeEntity{checkout}, content.Services)
	s.Nil(content.Deployment)
}

func (s *SituationNoticeServiceSuite) createRaisedSituation(ctx context.Context, tdb rez.Database) *ent.Situation {
	client := tdb.Client(ctx)
	createSeed := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	seed, seedErr := createSeed.Save(ctx)
	s.Require().NoError(seedErr)

	createSituation := client.Situation.Create().
		SetTitle("Checkout errors").
		SetOpenedAt(noticeTestOpenedAt).
		SetRaisedAt(noticeTestOpenedAt).
		SetSeedEntityID(seed.ID)
	situation, situationErr := createSituation.Save(ctx)
	s.Require().NoError(situationErr)
	return situation
}

// createNotice creates a pending notice of the situation, by a new installation.
func (s *SituationNoticeServiceSuite) createNotice(ctx context.Context, tdb rez.Database, situationID uuid.UUID, ref string, setFn func(*ent.SituationNoticeCreate)) *ent.SituationNotice {
	client := tdb.Client(ctx)
	createInstallation := client.Integration.Create().
		SetProvider("chat").
		SetName("chat").
		SetDisplayName(ref).
		SetProviderInstallationRef(ref).
		SetInstallationConfig([]byte(`{}`))
	installation, installationErr := createInstallation.Save(ctx)
	s.Require().NoError(installationErr)

	createNotice := client.SituationNotice.Create().
		SetSituationID(situationID).
		SetIntegrationID(installation.ID).
		SetChannelID("channel-1")
	setFn(createNotice)
	notice, noticeErr := createNotice.Save(ctx)
	s.Require().NoError(noticeErr)
	return notice
}

func (s *SituationNoticeServiceSuite) reloadNotice(ctx context.Context, tdb rez.Database, notice *ent.SituationNotice) *ent.SituationNotice {
	reloaded, getErr := tdb.Client(ctx).SituationNotice.Get(ctx, notice.ID)
	s.Require().NoError(getErr)
	return reloaded
}

// noticeSender is an installation that sends notices to channel-1 and is never expected to send one.
type noticeSender struct {
	installation *ent.Integration
	capabilities []string
}

func (f *noticeSender) Integration() *ent.Integration {
	return f.installation
}

func (f *noticeSender) Config() rez.IntegrationInstallationConfig {
	return nil
}

func (f *noticeSender) Capabilities() []string {
	return f.capabilities
}

func (f *noticeSender) SituationNoticeChannel() string {
	return "channel-1"
}

func (f *noticeSender) PostSituationNotice(context.Context, string, *integrations.SituationNoticeContent) (string, error) {
	return "", errors.New("unexpected post")
}

func (f *noticeSender) UpdateSituationNotice(context.Context, string, string, *integrations.SituationNoticeContent) error {
	return errors.New("unexpected update")
}

func (f *noticeSender) SituationNoticeLink(context.Context, string, string) (string, error) {
	return "", errors.New("unexpected link")
}

func (f *noticeSender) ChatChannelName(context.Context, string) (string, error) {
	return "", errors.New("unexpected channel name")
}
