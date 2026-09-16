package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/rezible/rezible/ent"
)

const MessagesOutboxTopic = "message_outbox"

func generateMessagesTableName(topic string) string {
	return fmt.Sprintf("\"%s\".\"%s\"", SchemaName, MessagesOutboxTopic)
}

type MessageTransport struct {
	pool   *pgxpool.Pool
	logger watermill.LoggerAdapter
}

func NewMessageTransport(pool *ConnectionPool) (*MessageTransport, error) {
	t := &MessageTransport{
		pool:   pool.Pool,
		logger: watermill.NewSlogLogger(slog.Default()),
	}
	return t, nil
}

var (
	messageSchemaAdapter = sql.PostgreSQLQueueSchema{
		GenerateMessagesTableName: generateMessagesTableName,
		SubscribeBatchSize:        1,
	}
	messageSubscriberOffsetsAdapter = sql.PostgreSQLQueueOffsetsAdapter{
		DeleteOnAck:               true,
		GenerateMessagesTableName: generateMessagesTableName,
	}
)

func (t *MessageTransport) Subscriber() (*sql.Subscriber, error) {
	cfg := sql.SubscriberConfig{
		ConsumerGroup:    "",
		AckDeadline:      new(30 * time.Second),
		PollInterval:     100 * time.Millisecond,
		ResendInterval:   time.Second,
		RetryInterval:    time.Second,
		InitializeSchema: false,
		SchemaAdapter:    messageSchemaAdapter,
		OffsetsAdapter:   messageSubscriberOffsetsAdapter,
	}
	return sql.NewSubscriber(sql.BeginnerFromPgx(t.pool), cfg, t.logger)
}

func (t *MessageTransport) Publisher() (*sql.Publisher, error) {
	cfg := sql.PublisherConfig{
		SchemaAdapter:        messageSchemaAdapter,
		AutoInitializeSchema: false,
	}
	return sql.NewPublisher(newMessagePublishExecutor(t.pool), cfg, t.logger)
}

type messagePublishExecutor struct {
	executor sql.ContextExecutor
}

func newMessagePublishExecutor(pool *pgxpool.Pool) *messagePublishExecutor {
	return &messagePublishExecutor{executor: sql.BeginnerFromPgx(pool)}
}

func (e *messagePublishExecutor) getExecutor(ctx context.Context) (sql.ContextExecutor, error) {
	if tx := ent.TxFromContext(ctx); tx != nil {
		pgxTx, txErr := ent.ExtractPgxTx(tx)
		if txErr != nil {
			return nil, fmt.Errorf("extract transaction: %w", txErr)
		}
		return sql.TxFromPgx(pgxTx), nil
	}
	return e.executor, nil
}

func (e *messagePublishExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	exec, execErr := e.getExecutor(ctx)
	if execErr != nil {
		return nil, execErr
	}
	return exec.ExecContext(ctx, query, args...)
}

func (e *messagePublishExecutor) QueryContext(ctx context.Context, query string, args ...any) (sql.Rows, error) {
	exec, execErr := e.getExecutor(ctx)
	if execErr != nil {
		return nil, execErr
	}
	return exec.QueryContext(ctx, query, args...)
}
