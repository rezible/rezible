package execution

import (
	"context"
	"slices"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "github.com/rezible/rezible/pkg/execution"

// attrsKey holds attributes added for the current unit of work. They are never
// serialized: jobs and messages carry only the execution Context.
type attrsKey struct{}

// Do runs fn as one named unit of work: it starts a span called name, adds attrs to the
// context so every log line and child span inside fn carries them, records a returned
// error on the span, and returns fn's error unchanged. It never logs.
func Do(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	ctx = WithAttrs(ctx, attrs...)
	ctx, span := otel.Tracer(tracerName).Start(ctx, name)
	defer span.End()
	span.SetAttributes(Attrs(ctx)...)

	fnErr := fn(ctx)
	if fnErr != nil {
		span.RecordError(fnErr)
		span.SetStatus(codes.Error, fnErr.Error())
	}
	return fnErr
}

// WithAttrs adds attrs to the context without starting a span, for values learned partway through.
func WithAttrs(ctx context.Context, attrs ...attribute.KeyValue) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	added, _ := ctx.Value(attrsKey{}).([]attribute.KeyValue)
	return context.WithValue(ctx, attrsKey{}, mergeAttrs(added, attrs))
}

// Attrs returns the identity attributes from the execution context, followed by those
// added with Do or WithAttrs.
func Attrs(ctx context.Context) []attribute.KeyValue {
	added, _ := ctx.Value(attrsKey{}).([]attribute.KeyValue)
	return mergeAttrs(identityAttrs(ctx), added)
}

// identityAttrs omits what is unset, including the actor of an anonymous context.
func identityAttrs(ctx context.Context) []attribute.KeyValue {
	exec, ok := getContext(ctx)
	if !ok {
		return nil
	}
	var attrs []attribute.KeyValue
	if exec.Auth.TenantID != nil {
		attrs = append(attrs, attribute.Int("tenant_id", *exec.Auth.TenantID))
	}
	if exec.ActorKind != "" && !exec.IsAnonymous() {
		attrs = append(attrs, attribute.String("actor", string(exec.ActorKind)))
	}
	if exec.Auth.UserID != nil {
		attrs = append(attrs, attribute.String("user_id", exec.Auth.UserID.String()))
	}
	if exec.Auth.AgentSessionID != nil {
		attrs = append(attrs, attribute.String("agent_session_id", exec.Auth.AgentSessionID.String()))
	}
	if exec.Provenance.ID != "" {
		attrs = append(attrs, attribute.String("request_id", exec.Provenance.ID))
	}
	return attrs
}

// mergeAttrs returns base followed by added, where a later value for a key replaces an earlier one in place.
func mergeAttrs(base, added []attribute.KeyValue) []attribute.KeyValue {
	merged := slices.Clone(base)
	for _, kv := range added {
		if i := slices.IndexFunc(merged, func(m attribute.KeyValue) bool { return m.Key == kv.Key }); i >= 0 {
			merged[i] = kv
		} else {
			merged = append(merged, kv)
		}
	}
	return merged
}
