package river

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"go.opentelemetry.io/otel/trace"

	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

// metadataKeyExecutionContext holds the encoded identity a job's worker runs with. Other keys in a job's
// metadata, such as the trace context, belong to River and its middleware.
const metadataKeyExecutionContext = "ec"

type jobMetadata struct {
	EncodedExecutionContext []byte `json:"ec"`
}

type accessContextMiddleware struct {
	river.MiddlewareDefaults
}

func (m *accessContextMiddleware) InsertMany(ctx context.Context, params []*rivertype.JobInsertParams, doInner func(context.Context) ([]*rivertype.JobInsertResult, error)) ([]*rivertype.JobInsertResult, error) {
	for _, p := range params {
		encodedExec, encodeErr := m.encodeWorkerExecutionContext(ctx, p.Args)
		if encodeErr != nil {
			return nil, fmt.Errorf("job %q: %w", p.Kind, encodeErr)
		}

		meta := make(map[string]json.RawMessage)
		if len(p.Metadata) > 0 {
			if jsonErr := json.Unmarshal(p.Metadata, &meta); jsonErr != nil {
				return nil, fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
			}
		}
		encodedValue, encodeValueErr := json.Marshal(encodedExec)
		if encodeValueErr != nil {
			return nil, fmt.Errorf("failed to marshal execution context: %w", encodeValueErr)
		}
		meta[metadataKeyExecutionContext] = encodedValue

		var marshalErr error
		p.Metadata, marshalErr = json.Marshal(meta)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal job metadata: %w", marshalErr)
		}
	}

	return doInner(ctx)
}

// encodeWorkerExecutionContext returns the identity the job's worker runs with, taken from the
// inserting context. System jobs encode nothing and run under the job client's system root context.
func (m *accessContextMiddleware) encodeWorkerExecutionContext(ctx context.Context, args rivertype.JobArgs) ([]byte, error) {
	kind := execution.KindTenant
	if identityArgs, hasIdentity := args.(jobs.ArgsWithWorkerExecutionContext); hasIdentity {
		kind = identityArgs.WorkerExecutionContext()
	}
	caller := execution.GetContext(ctx)
	switch kind {
	case execution.KindTenant:
		tenantID, tenantOK := caller.TenantID()
		if !tenantOK {
			return nil, errs.ErrTenantContextMissing
		}
		return execution.GetContext(execution.NewTenantContext(ctx, tenantID)).Encode()
	case execution.KindUser, execution.KindAiAgent:
		if caller.ActorKind != kind {
			return nil, fmt.Errorf("%w: %s job inserted by %s actor", errs.ErrForbidden, kind, caller.ActorKind)
		}
		return caller.Encode()
	case execution.KindSystem:
		if !caller.IsSystem() {
			return nil, fmt.Errorf("%w: system job inserted by %s actor", errs.ErrForbidden, caller.ActorKind)
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("%w: invalid worker execution context %q", errs.ErrForbidden, kind)
	}
}

func (m *accessContextMiddleware) Work(ctx context.Context, job *rivertype.JobRow, doInner func(context.Context) error) error {
	ctx, restoreErr := restoreExecutionContext(ctx, job)
	if restoreErr != nil {
		return restoreErr
	}
	trace.SpanFromContext(ctx).SetAttributes(execution.Attrs(ctx)...)
	return doInner(ctx)
}

// restoreExecutionContext returns ctx with the identity the job was inserted to run with. System jobs have
// none and keep the job client's system root context.
func restoreExecutionContext(ctx context.Context, job *rivertype.JobRow) (context.Context, error) {
	var meta jobMetadata
	if jsonErr := json.Unmarshal(job.Metadata, &meta); jsonErr != nil {
		return ctx, fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
	}
	if len(meta.EncodedExecutionContext) == 0 {
		return ctx, nil
	}
	exec, decodeErr := execution.DecodeContext(meta.EncodedExecutionContext)
	if decodeErr != nil {
		return ctx, fmt.Errorf("invalid execution context for job: %w", decodeErr)
	}
	return execution.SetContext(ctx, exec), nil
}

// jobErrorHandler logs, once at Error, a job's final failed attempt and any panic. Earlier failed attempts
// are retried and stay in River's own logs below Warn. River calls it outside the middleware and the job's
// work span, so it restores the job's identity itself; the line carries no trace, and the job's ID joins it
// to the work span.
type jobErrorHandler struct{}

func (jobErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, jobErr error) *river.ErrorHandlerResult {
	if job.Attempt < job.MaxAttempts {
		return nil
	}
	slog.ErrorContext(jobLogContext(ctx, job), "job failed",
		"job_kind", job.Kind,
		"job_id", job.ID,
		"attempt", job.Attempt,
		"error", jobErr,
	)
	return nil
}

func (jobErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, stack string) *river.ErrorHandlerResult {
	slog.ErrorContext(jobLogContext(ctx, job), "job panicked",
		"job_kind", job.Kind,
		"job_id", job.ID,
		"attempt", job.Attempt,
		"panic", fmt.Sprint(panicVal),
		"stack", stack,
	)
	return nil
}

// jobLogContext is the job's execution context, when its metadata holds one.
func jobLogContext(ctx context.Context, job *rivertype.JobRow) context.Context {
	if restored, restoreErr := restoreExecutionContext(ctx, job); restoreErr == nil {
		return restored
	}
	return ctx
}

// riverLogHandler keeps River's own logs at Warn and above: it reports every failed attempt at Info.
// River also logs a recovered panic at Error, without the job's context; jobErrorHandler reports the panic
// with it, so River's record is lowered to Warn.
type riverLogHandler struct {
	slog.Handler
}

const riverPanicMessageSuffix = "panic recovery; possible bug with Worker"

func (h riverLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn && h.Handler.Enabled(ctx, level)
}

func (h riverLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level == slog.LevelError && strings.HasSuffix(record.Message, riverPanicMessageSuffix) {
		record.Level = slog.LevelWarn
	}
	return h.Handler.Handle(ctx, record)
}

func (h riverLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return riverLogHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h riverLogHandler) WithGroup(name string) slog.Handler {
	return riverLogHandler{Handler: h.Handler.WithGroup(name)}
}
