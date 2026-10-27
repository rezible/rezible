package watermill

import (
	"context"
	"fmt"
	"time"

	wsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres/migrations"
)

// The table is unqualified, so the connection's search path resolves its schema, as for the ent tables.
func generateOutboxTableName(topic string) string {
	return fmt.Sprintf("\"%s\"", migrations.OutboxTableName)
}

var (
	outboxSchemaAdapter = wsql.PostgreSQLQueueSchema{
		GenerateMessagesTableName: generateOutboxTableName,
		SubscribeBatchSize:        1,
	}
	outboxOffsetsAdapter = wsql.PostgreSQLQueueOffsetsAdapter{
		DeleteOnAck:               true,
		GenerateMessagesTableName: generateOutboxTableName,
	}
)

func (ms *MessageQueue) makeOutboxSubscriber() (message.Subscriber, error) {
	cfg := wsql.SubscriberConfig{
		ConsumerGroup:    "",
		AckDeadline:      new(30 * time.Second),
		PollInterval:     100 * time.Millisecond,
		ResendInterval:   time.Second,
		RetryInterval:    time.Second,
		InitializeSchema: false,
		SchemaAdapter:    outboxSchemaAdapter,
		OffsetsAdapter:   outboxOffsetsAdapter,
	}
	return wsql.NewSubscriber(ms.outboxDB, cfg, ms.logger)
}

func (ms *MessageQueue) makeOutboxPublisher() (message.Publisher, error) {
	cfg := wsql.PublisherConfig{
		SchemaAdapter:        outboxSchemaAdapter,
		AutoInitializeSchema: false,
	}
	return wsql.NewPublisher(&outboxPublishExecutor{executor: ms.outboxDB}, cfg, ms.logger)
}

func (ms *MessageQueue) registerOutboxBacklogGauge() (metric.Registration, error) {
	meter := otel.Meter("github.com/rezible/rezible")
	backlog, backlogErr := meter.Int64ObservableGauge("backend.messages.outbox_backlog",
		metric.WithDescription("Events written to the message outbox and not yet forwarded"))
	if backlogErr != nil {
		return nil, fmt.Errorf("create gauge: %w", backlogErr)
	}
	return meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		// Raw SQL: the outbox table is not an ent entity.
		rows, queryErr := ms.outboxDB.QueryContext(ctx, "SELECT count(*) FROM "+generateOutboxTableName(""))
		if queryErr != nil {
			return fmt.Errorf("count message outbox: %w", queryErr)
		}
		defer rows.Close()
		var count int64
		if rows.Next() {
			if scanErr := rows.Scan(&count); scanErr != nil {
				return fmt.Errorf("scan message outbox count: %w", scanErr)
			}
		}
		observer.ObserveInt64(backlog, count)
		return nil
	}, backlog)
}

// outboxPublishExecutor writes through the ent transaction in the context when there is one, so the event is
// committed or rolled back with it.
type outboxPublishExecutor struct {
	executor wsql.ContextExecutor
}

func (e *outboxPublishExecutor) getExecutor(ctx context.Context) (wsql.ContextExecutor, error) {
	if tx := ent.TxFromContext(ctx); tx != nil {
		pgxTx, txErr := ent.ExtractPgxTx(tx)
		if txErr != nil {
			return nil, fmt.Errorf("extract transaction: %w", txErr)
		}
		return wsql.TxFromPgx(pgxTx), nil
	}
	return e.executor, nil
}

func (e *outboxPublishExecutor) ExecContext(ctx context.Context, query string, args ...any) (wsql.Result, error) {
	exec, execErr := e.getExecutor(ctx)
	if execErr != nil {
		return nil, execErr
	}
	return exec.ExecContext(ctx, query, args...)
}

func (e *outboxPublishExecutor) QueryContext(ctx context.Context, query string, args ...any) (wsql.Rows, error) {
	exec, execErr := e.getExecutor(ctx)
	if execErr != nil {
		return nil, execErr
	}
	return exec.QueryContext(ctx, query, args...)
}
