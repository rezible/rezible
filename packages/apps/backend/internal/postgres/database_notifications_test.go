package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DatabaseNotificationServiceSuite struct {
	test.Suite
}

func TestDatabaseNotificationServiceSuite(t *testing.T) {
	suite.Run(t, &DatabaseNotificationServiceSuite{Suite: test.NewSuite()})
}

func (s *DatabaseNotificationServiceSuite) TestListenDeliversPayload() {
	ctx := s.T().Context()

	svc := postgres.NewDatabaseNotificationService(s.PostgresPool())
	listenCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	connected := make(chan struct{})
	received := make(chan []byte, 1)
	errs := make(chan error, 1)

	go func() {
		errs <- svc.Listen(
			listenCtx,
			"rezible_test_notifications",
			func(context.Context) error {
				close(connected)
				return nil
			},
			func(_ context.Context, payload []byte) error {
				received <- payload
				cancel()
				return nil
			},
		)
	}()

	select {
	case <-connected:
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for listener to connect")
	}

	_, notifyErr := s.PostgresPool().Exec(ctx, "SELECT pg_notify($1, $2)", "rezible_test_notifications", "test-payload")
	s.Require().NoError(notifyErr)

	select {
	case payload := <-received:
		s.Equal([]byte("test-payload"), payload)
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for notification")
	}

	select {
	case listenErr := <-errs:
		s.Require().NoError(listenErr)
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for listener to stop")
	}
}

func (s *DatabaseNotificationServiceSuite) TestListenValidatesInputs() {
	svc := postgres.NewDatabaseNotificationService(nil)
	s.Require().ErrorContains(svc.Listen(s.T().Context(), "invalid-channel-name", nil, func(context.Context, []byte) error { return nil }), "invalid notification channel")
	s.Require().ErrorContains(svc.Listen(s.T().Context(), "valid_channel", nil, nil), "notification handler is required")
	s.Require().ErrorContains(svc.Listen(s.T().Context(), "valid_channel", nil, func(context.Context, []byte) error { return nil }), "pgx pool is nil")
}
