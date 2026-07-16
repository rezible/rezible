package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AiAgentSnapshotServiceSuite struct {
	test.Suite
}

func TestAiAgentSnapshotServiceSuite(t *testing.T) {
	suite.Run(t, &AiAgentSnapshotServiceSuite{Suite: test.NewSuite()})
}

func (s *AiAgentSnapshotServiceSuite) newSnapshotService() *AiAgentSnapshotService {
	msgs := mocks.NewMockMessageService(s.T())
	msgs.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	notifications := mocks.NewMockDatabaseNotificationService(s.T())

	svc, err := NewAiAgentSnapshotService(s.Database(), msgs, notifications)
	s.Require().NoError(err)
	return svc
}

func (s *AiAgentSnapshotServiceSuite) createAgentRun(ctx context.Context) *ent.AiAgentRun {
	run, err := s.Client(ctx).AiAgentRun.Create().
		SetAgentName("test_agent").
		SetInput([]byte(`{}`)).
		SetOwnerUserID(s.SeedUser.ID).
		Save(ctx)
	s.Require().NoError(err)
	return run
}

func (s *AiAgentSnapshotServiceSuite) createSnapshot(ctx context.Context, run *ent.AiAgentRun, status aars.Status) *ent.AiAgentRunSnapshot {
	now := time.Now().UTC()
	snap, err := s.Client(ctx).AiAgentRunSnapshot.Create().
		SetAiAgentRunID(run.ID).
		SetStatus(status).
		SetFinishReason("").
		SetState([]byte(`{}`)).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	s.Require().NoError(err)
	return snap
}

func (s *AiAgentSnapshotServiceSuite) TestSubscribeMissingSnapshotClosesChannel() {
	ctx := s.SeedTenantContext()
	svc := s.newSnapshotService()

	ch := svc.OnSnapshotStatusChange(ctx, uuid.New())

	_, ok := <-ch
	s.False(ok)
}

func (s *AiAgentSnapshotServiceSuite) TestNotificationPayloadNotifiesSubscriber() {
	ctx := s.SeedTenantContext()
	svc := s.newSnapshotService()
	run := s.createAgentRun(ctx)
	snap := s.createSnapshot(ctx, run, aars.StatusPending)

	ch := svc.OnSnapshotStatusChange(ctx, snap.ID)
	s.Require().Equal(aix.SnapshotStatusPending, <-ch)

	payload, marshalErr := json.Marshal(aiAgentSnapshotStatusNotification{
		TenantID:   s.SeedTenant.ID,
		SnapshotID: snap.ID,
		Status:     aars.StatusCompleted,
		UpdatedAt:  time.Now().UTC(),
	})
	s.Require().NoError(marshalErr)
	s.Require().NoError(svc.handleSnapshotStatusNotification(ctx, payload))

	select {
	case status, ok := <-ch:
		s.Require().True(ok)
		s.Equal(aix.SnapshotStatusCompleted, status)
	case <-time.After(time.Second):
		s.Fail("timed out waiting for status update")
	}
}

func (s *AiAgentSnapshotServiceSuite) TestStartListensForSnapshotStatusNotifications() {
	ctx := s.SeedTenantContext()

	var connected bool
	var notified bool
	notifications := mocks.NewMockDatabaseNotificationService(s.T())
	notifications.EXPECT().
		Listen(mock.Anything, aiAgentSnapshotStatusChannel, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, channel string, onConnect func(context.Context) error, onNotify func(context.Context, []byte) error) error {
			s.Equal(aiAgentSnapshotStatusChannel, channel)
			s.Require().NoError(onConnect(ctx))
			connected = true
			s.Require().NoError(onNotify(ctx, []byte(`{"snapshot_id":"00000000-0000-0000-0000-000000000001","status":"completed","updated_at":"2026-01-01T00:00:00Z"}`)))
			notified = true
			return nil
		})

	msgs := mocks.NewMockMessageService(s.T())
	svc, err := NewAiAgentSnapshotService(s.Database(), msgs, notifications)
	s.Require().NoError(err)
	s.Require().NoError(svc.Start(ctx))
	s.True(connected)
	s.True(notified)
}

func (s *AiAgentSnapshotServiceSuite) TestConnectRefreshDeliversChangeAfterSubscriptionBaseline() {
	ctx := s.SeedTenantContext()
	svc := s.newSnapshotService()
	run := s.createAgentRun(ctx)
	snap := s.createSnapshot(ctx, run, aars.StatusPending)

	ch := svc.OnSnapshotStatusChange(ctx, snap.ID)
	s.Require().Equal(aix.SnapshotStatusPending, <-ch)

	updatedAt := time.Now().UTC().Add(time.Second)
	_, err := s.Client(ctx).AiAgentRunSnapshot.UpdateOneID(snap.ID).
		SetStatus(aars.StatusAborted).
		SetUpdatedAt(updatedAt).
		Save(ctx)
	s.Require().NoError(err)

	s.Require().NoError(svc.refreshSubscribedStatuses(ctx))

	select {
	case status, ok := <-ch:
		s.Require().True(ok)
		s.Equal(aix.SnapshotStatusAborted, status)
	case <-time.After(time.Second):
		s.Fail("timed out waiting for polled status update")
	}
}

func (s *AiAgentSnapshotServiceSuite) TestSnapshotUpdatesAreSerializedBySnapshotID() {
	ctx := s.SeedTenantContext()
	svc := s.newSnapshotService()
	run := s.createAgentRun(ctx)
	snap := s.createSnapshot(ctx, run, aars.StatusPending)

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := svc.UpdateAgentRunSnapshot(ctx, snap.ID, func(curr *ent.AiAgentRunSnapshot, m *ent.AiAgentRunSnapshotMutation) (*rez.AiAgentSnapshotDelta, error) {
			close(firstStarted)
			<-releaseFirst
			m.SetCreatedAt(curr.CreatedAt)
			m.SetUpdatedAt(time.Now().UTC())
			m.SetStatus(aars.StatusCompleted)
			m.SetFinishReason("")
			return &rez.AiAgentSnapshotDelta{Status: new(aix.SnapshotStatusCompleted)}, nil
		})
		firstDone <- err
	}()
	<-firstStarted

	secondStarted := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		_, err := svc.UpdateAgentRunSnapshot(ctx, snap.ID, func(curr *ent.AiAgentRunSnapshot, m *ent.AiAgentRunSnapshotMutation) (*rez.AiAgentSnapshotDelta, error) {
			close(secondStarted)
			if curr.Status != aars.StatusPending {
				return nil, nil
			}
			m.SetCreatedAt(curr.CreatedAt)
			m.SetUpdatedAt(time.Now().UTC())
			m.SetStatus(aars.StatusAborted)
			m.SetFinishReason("")
			return &rez.AiAgentSnapshotDelta{Status: new(aix.SnapshotStatusAborted)}, nil
		})
		secondDone <- err
	}()

	select {
	case <-secondStarted:
		s.Fail("second update callback started before first transaction released snapshot lock")
	case <-time.After(100 * time.Millisecond):
	}

	close(releaseFirst)
	s.Require().NoError(<-firstDone)

	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		s.Fail("second update callback did not start after first transaction completed")
	}
	s.Require().NoError(<-secondDone)
}

func (s *AiAgentSnapshotServiceSuite) TestLatestSnapshotBreaksCreatedAtTiesByID() {
	ctx := s.SeedTenantContext()
	svc := s.newSnapshotService()
	run := s.createAgentRun(ctx)
	now := time.Now().UTC()

	lowID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	highID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	for _, id := range []uuid.UUID{lowID, highID} {
		_, err := s.Client(ctx).AiAgentRunSnapshot.Create().
			SetID(id).
			SetAiAgentRunID(run.ID).
			SetStatus(aars.StatusCompleted).
			SetFinishReason("").
			SetState([]byte(`{}`)).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		s.Require().NoError(err)
	}

	latest, err := svc.GetLatestSnapshotForRun(ctx, run.ID)
	s.Require().NoError(err)
	s.Require().NotNil(latest)
	s.Equal(highID, latest.ID)
}
