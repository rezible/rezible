import { expect, spyOn, test } from "bun:test";
import { QueryClient } from "@tanstack/svelte-query";
import { client } from "$lib/api";
import { filteredEventsOptions } from "../filters";

test("relative event windows advance on refresh without changing the cached working set", async () => {
	const originalConfig = client.getConfig();
	const requests: URL[] = [];
	const now = Date.parse("2026-09-09T08:00:00Z");
	const clock = spyOn(Date, "now").mockReturnValue(now);
	const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	const fetchMock = Object.assign(
		async (request: RequestInfo | URL) => {
			requests.push(new URL(request instanceof Request ? request.url : request));
			return new Response(
				JSON.stringify({ data: [], pagination: { page: 1, pageSize: 5, total: 0 } }),
				{
					headers: { "content-type": "application/json" },
				}
			);
		},
		{ preconnect: globalThis.fetch.preconnect }
	);
	client.setConfig({ baseUrl: "https://example.invalid/api/v1", fetch: fetchMock });
	try {
		const filters = { kind: "deployment", time: "24h" } as const;
		const options = filteredEventsOptions(filters, { page: 1, pageSize: 5 });
		await queryClient.fetchQuery(options);
		clock.mockReturnValue(now + 300_000);
		await queryClient.fetchQuery(filteredEventsOptions(filters, { page: 1, pageSize: 5 }));
		expect(requests).toHaveLength(2);
		expect(queryClient.getQueryCache().getAll()).toHaveLength(1);
		for (const [i, request] of requests.entries()) {
			expect(request.searchParams.get("kind")).toBe("deployment");
			expect(request.searchParams.get("to")).toBe(new Date(now + i * 300_000).toISOString());
			expect(request.searchParams.get("from")).toBe(
				new Date(now + i * 300_000 - 86_400_000).toISOString()
			);
			expect(request.searchParams.has("withProjection")).toBe(false);
		}
	} finally {
		clock.mockRestore();
		client.setConfig(originalConfig);
		queryClient.clear();
	}
});
