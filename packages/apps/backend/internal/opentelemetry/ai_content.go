package opentelemetry

import (
	"context"
	"slices"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// aiContentKeys are the span attributes on which Genkit 1.13 records content: each action's input and
// output (prompts, tool results and the provider data tools read) and its init (session state passed with
// WithState: earlier messages, artifacts and custom state). Its other attributes are names, paths, states,
// flags and IDs (core/tracing/tracing.go, ai/exp/agent.go).
var aiContentKeys = []attribute.Key{"genkit:input", "genkit:output", "genkit:init"}

const redactedValue = "<redacted>"

// withAiContentPolicy returns exporter unchanged when AI content may be captured, and otherwise one that
// redacts Genkit's content attributes before export, as Genkit's own Google Cloud plugin does.
func withAiContentPolicy(exporter sdktrace.SpanExporter, captureAiContent bool) sdktrace.SpanExporter {
	if captureAiContent {
		return exporter
	}
	return aiContentRedactingExporter{SpanExporter: exporter}
}

type aiContentRedactingExporter struct {
	sdktrace.SpanExporter
}

func (e aiContentRedactingExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	redacted := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		redacted[i] = redactAiContent(span)
	}
	return e.SpanExporter.ExportSpans(ctx, redacted)
}

// genkitScope is the instrumentation scope of every span Genkit 1.13 creates: each is started by
// tracing.RunInNewSpan from tracing.Tracer().
const genkitScope = "genkit-tracer"

// redactAiContent redacts a Genkit span's content attributes and its error text, which can quote prompts,
// tool results or provider data: the status description and the message of its exception events. Genkit
// records errors without a stack trace. Other spans, and a Genkit span's other attributes and events, are
// unchanged.
func redactAiContent(span sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	if span.InstrumentationScope().Name != genkitScope {
		return span
	}
	attrs := slices.Clone(span.Attributes())
	for i, attr := range attrs {
		if slices.Contains(aiContentKeys, attr.Key) {
			attrs[i] = attr.Key.String(redactedValue)
		}
	}
	status := span.Status()
	if status.Description != "" {
		status.Description = redactedValue
	}
	events := slices.Clone(span.Events())
	for i, event := range events {
		if event.Name != semconv.ExceptionEventName {
			continue
		}
		event.Attributes = slices.Clone(event.Attributes)
		for j, attr := range event.Attributes {
			if attr.Key == semconv.ExceptionMessageKey {
				event.Attributes[j] = attr.Key.String(redactedValue)
			}
		}
		events[i] = event
	}
	return redactedSpan{ReadOnlySpan: span, attrs: attrs, status: status, events: events}
}

type redactedSpan struct {
	sdktrace.ReadOnlySpan
	attrs  []attribute.KeyValue
	status sdktrace.Status
	events []sdktrace.Event
}

func (s redactedSpan) Attributes() []attribute.KeyValue { return s.attrs }
func (s redactedSpan) Status() sdktrace.Status          { return s.status }
func (s redactedSpan) Events() []sdktrace.Event         { return s.events }
