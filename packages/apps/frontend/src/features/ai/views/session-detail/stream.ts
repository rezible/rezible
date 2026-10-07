import { streamAgentSessionEvents, type StreamAgentSessionEventsResponse } from "$lib/api";

export type SessionStreamEvent = StreamAgentSessionEventsResponse[number];

export const consumeSessionEvents = async (
	options: Omit<Parameters<typeof streamAgentSessionEvents>[0], "onSseEvent">,
	onEvent: (event: SessionStreamEvent) => Promise<void>
) => {
	let received: SessionStreamEvent | undefined;
	const { stream } = await streamAgentSessionEvents({
		...options,
		onSseEvent: ({ event, data }) => {
			if (event !== "turn-updated" && event !== "turn-chunk") return;
			// OpenAPI describes SSE as an array of envelopes, but each wire event contains
			// one payload. The callback preserves its name; the generator yields only data.
			received = { event, data } as unknown as SessionStreamEvent;
		},
	});
	// The generated client is lazy: iteration opens the request and keeps it alive.
	// Await dispatch before advancing so async status refetches cannot race each other.
	for await (const _payload of stream) {
		const event = received;
		received = undefined;
		if (event && !options.signal?.aborted) {
			await onEvent(event);
		}
	}
};
