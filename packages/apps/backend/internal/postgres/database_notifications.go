package postgres

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var postgresNotificationChannelPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

type DatabaseNotificationService struct {
	pool *PgxPool
}

type reconnectNotificationError struct {
	err error
}

func (e reconnectNotificationError) Error() string {
	return e.err.Error()
}

func (e reconnectNotificationError) Unwrap() error {
	return e.err
}

func NewDatabaseNotificationService(pool *PgxPool) *DatabaseNotificationService {
	return &DatabaseNotificationService{pool: pool}
}

func (s *DatabaseNotificationService) Listen(ctx context.Context, channel string, onConnect func(context.Context) error, onNotify func(context.Context, []byte) error) error {
	if !postgresNotificationChannelPattern.MatchString(channel) {
		return fmt.Errorf("invalid notification channel: %s", channel)
	}
	if onNotify == nil {
		return fmt.Errorf("notification handler is required")
	}
	if s.pool == nil {
		return fmt.Errorf("pgx pool is nil")
	}

	query := "LISTEN " + pgx.Identifier{channel}.Sanitize()
	for {
		if ctx.Err() != nil {
			return nil
		}

		listenErr := s.listen(ctx, query, onConnect, onNotify)
		if listenErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return nil
		}
		var reconnectErr reconnectNotificationError
		if !errors.As(listenErr, &reconnectErr) {
			return listenErr
		}
		if listenErr := waitBeforeNotificationReconnect(ctx); listenErr != nil {
			return nil
		}
	}
}

func (s *DatabaseNotificationService) listen(ctx context.Context, query string, onConnect func(context.Context) error, onNotify func(context.Context, []byte) error) error {
	conn, acquireErr := s.pool.Acquire(ctx)
	if acquireErr != nil {
		return reconnectNotificationError{err: fmt.Errorf("acquire notification connection: %w", acquireErr)}
	}
	defer conn.Release()

	if _, listenErr := conn.Exec(ctx, query); listenErr != nil {
		return reconnectNotificationError{err: fmt.Errorf("listen: %w", listenErr)}
	}
	if onConnect != nil {
		if connectErr := onConnect(ctx); connectErr != nil {
			return connectErr
		}
	}

	for {
		notification, waitErr := conn.Conn().WaitForNotification(ctx)
		if waitErr != nil {
			return reconnectNotificationError{err: fmt.Errorf("wait for notification: %w", waitErr)}
		}
		if notification == nil {
			continue
		}
		if notifyErr := onNotify(ctx, []byte(notification.Payload)); notifyErr != nil {
			return notifyErr
		}
	}
}

func waitBeforeNotificationReconnect(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
