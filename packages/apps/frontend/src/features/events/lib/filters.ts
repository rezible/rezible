import { z } from "zod";
import {
	listEvents,
	listEventsQueryKey,
	type ListEventsData,
	type ListEventsResponse,
	type ErrorModel,
} from "$lib/api";
import { queryOptions } from "@tanstack/svelte-query";

export const eventTimeOptions = [
	{ value: "any", label: "Any time", hours: 0 },
	{ value: "24h", label: "Last 24 hours", hours: 24 },
	{ value: "7d", label: "Last 7 days", hours: 168 },
	{ value: "30d", label: "Last 30 days", hours: 720 },
] as const;

export const eventFilterSchema = z.object({
	kind: z.string().default("").catch(""),
	time: z.enum(["any", "24h", "7d", "30d"]).default("any").catch("any"),
});

export type EventFilters = z.infer<typeof eventFilterSchema>;

export const eventQueryFilters = (filters: EventFilters, now: number) => {
	const hours = eventTimeOptions.find((option) => option.value === filters.time)!.hours;
	return {
		kind: filters.kind.trim() || undefined,
		from: hours ? new Date(now - hours * 3_600_000).toISOString() : undefined,
		to: hours ? new Date(now).toISOString() : undefined,
	};
};

// Relative windows keep the same query identity while their bounds advance on each fetch.
export const filteredEventsOptions = (
	filters: EventFilters,
	query: Omit<NonNullable<ListEventsData["query"]>, "kind" | "from" | "to">
) => {
	const queryKey = listEventsQueryKey({ query: { ...query, kind: filters.kind.trim() || undefined } });
	return queryOptions<ListEventsResponse, ErrorModel>({
		queryKey: [...queryKey, { time: filters.time }],
		queryFn: async ({ signal }: { signal: AbortSignal }) => {
			const { data } = await listEvents({
				query: { ...query, ...eventQueryFilters(filters, Date.now()) },
				signal,
				throwOnError: true,
			});
			return data;
		},
	});
};
