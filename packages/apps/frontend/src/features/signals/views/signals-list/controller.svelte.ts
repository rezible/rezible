import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { z } from "zod";
import { page } from "$app/state";
import { createQuery } from "@tanstack/svelte-query";
import { listAlertsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { attentionLevelOptions, type AttentionLevel } from "$features/signals/lib/attention";

type AttentionFilter = AttentionLevel | "any";

const signalFilterSchema = z.object({
	search: z.string().default("").catch(""),
	attention: z.enum(["any", "default", "watch_only", "join_only"]).default("any").catch("any"),
});

type SignalFilters = z.infer<typeof signalFilterSchema>;

const attentionFilterOptions: { value: AttentionFilter; label: string }[] = [
	{ value: "any", label: "All" },
	...attentionLevelOptions.map((option) => ({ value: option.value, label: option.label })),
];

function attentionQuery(attention: AttentionFilter) {
	if (attention === "any") {
		return undefined;
	}
	return attention;
}

class SignalsListViewController {
	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
	}

	private params = useSearchParams(signalFilterSchema, {
		debounce: 300,
		noScroll: true,
	});

	filters = $derived(signalFilterSchema.parse(this.params));
	private committed = $derived(signalFilterSchema.parse(Object.fromEntries(page.url.searchParams)));
	private search = $derived(this.committed.search.trim() || undefined);

	private filterKey = $derived(JSON.stringify([this.search, this.committed.attention]));

	paginatedAlertsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listAlertsOptions({
				query: {
					search: this.search,
					attention: attentionQuery(this.committed.attention),
					...pagination,
				},
			}),
		resetWhen: () => this.filterKey,
	});

	query = $derived(this.paginatedAlertsQuery.query);
	total = $derived(this.query.data?.pagination.total);
	activeFilterCount = $derived(
		[this.filters.search.trim() !== "", this.filters.attention !== "any"].filter(Boolean).length
	);

	private attentionTotals = attentionFilterOptions.map((option) =>
		createQuery(() =>
			listAlertsOptions({
				query: { search: this.search, attention: attentionQuery(option.value), page: 1, pageSize: 1 },
			})
		)
	);

	/** Attention options with how many signals each would show for the current search. */
	attentionOptions = $derived(
		attentionFilterOptions.map((option, index) => ({
			...option,
			count: this.attentionTotals[index].data?.pagination.total,
		}))
	);

	setFilters = (values: Partial<SignalFilters>) => this.params.update(values);
	setSearch = (search: string) => this.setFilters({ search });
	setAttention = (attention: AttentionFilter) => this.setFilters({ attention });
	resetFilters = () => this.setFilters(signalFilterSchema.parse({}));
}

const ctx = new Context<SignalsListViewController>("SignalsListViewController");
export const initSignalsListViewController = () => ctx.set(new SignalsListViewController());
