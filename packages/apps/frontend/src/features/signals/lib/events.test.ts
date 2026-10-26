import { expect, test } from "bun:test";
import type { Event } from "$lib/api";
import { eventRowPresentation } from "./events";

function event(kind: string, attributes: unknown): Event {
	return {
		id: "event-1",
		attributes: {
			kind,
			attributes: btoa(JSON.stringify(attributes)),
			occurredAt: "2026-05-14T12:00:00Z",
			receivedAt: "2026-05-14T12:00:01Z",
			providerEventRef: "ref",
			providerEventSource: "source",
			resourceRef: { provider: "demo", providerNamespace: "demo", resourceRef: "demo:alert:latency" },
		},
	};
}

test("an event's own name is the title and its identifier is secondary", () => {
	const row = eventRowPresentation(event("code_change", { display_name: "Raise search timeout" }));

	expect(row.title).toBe("Raise search timeout");
	expect(row.identifier).toBe("demo:alert:latency");
});

test("an event without a name falls back to its identifier as the title", () => {
	const row = eventRowPresentation(event("deployment", { status: "started" }));

	expect(row.title).toBe("demo:alert:latency");
	expect(row.identifier).toBeUndefined();
});
