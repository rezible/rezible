import { expect, test } from "bun:test";
import { consumeSessionEvents, type SessionStreamEvent } from "./stream";

test("consumes the generated SSE stream and dispatches individual named events in order", async () => {
	const startedAt = "2026-10-07T00:00:00.123456Z";
	const events: SessionStreamEvent[] = [
		{
			event: "turn-updated",
			data: { sessionId: "session", turnId: "turn", startedAt, status: "running" },
		},
		{
			event: "turn-chunk",
			data: {
				sessionId: "session",
				turnId: "turn",
				startedAt,
				model: {
					index: 0,
					role: "model",
					aggregated: false,
					parts: [{ kind: "text", text: "prefix" }],
				},
			},
		},
	];
	let requests = 0;
	const fetchStream = Object.assign(
		async () => {
			requests++;
			return new Response(
				events
					.map(({ event, data }) => `event: ${event}\ndata: ${JSON.stringify(data)}\n\n`)
					.join(""),
				{ headers: { "Content-Type": "text/event-stream" } }
			);
		},
		{ preconnect: () => {} }
	);
	const received: SessionStreamEvent[] = [];
	let dispatching = false;
	await consumeSessionEvents(
		{ path: { id: "session" }, baseUrl: "https://example.test", fetch: fetchStream },
		async (event) => {
			expect(dispatching).toBe(false);
			dispatching = true;
			await new Promise((resolve) => setTimeout(resolve, 1));
			received.push(event);
			dispatching = false;
		}
	);
	expect(requests).toBe(1);
	expect(received).toEqual(events);
});

test("an already aborted stream does not open a request", async () => {
	const abort = new AbortController();
	abort.abort();
	let requests = 0;
	await consumeSessionEvents(
		{
			path: { id: "session" },
			baseUrl: "https://example.test",
			signal: abort.signal,
			fetch: Object.assign(
				async () => {
					requests++;
					return new Response();
				},
				{ preconnect: () => {} }
			),
		},
		async () => {
			throw new Error("unexpected dispatch");
		}
	);
	expect(requests).toBe(0);
});
