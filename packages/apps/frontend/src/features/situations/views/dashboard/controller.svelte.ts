import { listSituationsOptions, type Situation } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useSearchParams } from "runed/kit";
import { z } from "zod";
import { Context } from "runed";
import { byAttentionPriority } from "$features/situations/lib/presentation";

export const situationViews = ["needs-attention", "active", "history"] as const;
export type SituationView = (typeof situationViews)[number];

export const openedRanges = {
	"": { label: "Any time", hours: 0 },
	"24h": { label: "Last 24 hours", hours: 24 },
	"7d": { label: "Last 7 days", hours: 24 * 7 },
	"30d": { label: "Last 30 days", hours: 24 * 30 },
} as const;
export type OpenedRange = keyof typeof openedRanges;

const paramsSchema = z.object({
	view: z.enum(situationViews).default("needs-attention").catch("needs-attention"),
	q: z.string().default("").catch(""),
	opened: z.enum(Object.keys(openedRanges) as [OpenedRange]).default("" as OpenedRange).catch("" as OpenedRange),
});

export class SituationsDashboardController {
	private params = useSearchParams(paramsSchema, { debounce: 300 });

	private paginatedQuery = createPaginatedQuery({
		queryOptions: (pagination) => {
			const range = openedRanges[this.params.opened].hours;
			return listSituationsOptions({
				query: {
					...pagination,
					search: this.params.q || undefined,
					status: this.params.view === "history" ? "closed" : "open",
					openedAfter: range > 0 ? new Date(Date.now() - range * 3_600_000).toISOString() : undefined,
				},
			});
		},
		resetWhen: () => [this.params.view, this.params.q, this.params.opened],
	});

	query = $derived(this.paginatedQuery.query);
	private situationsData = $derived(this.query.data?.data ?? []);
	private sortedSituationsData = $derived([...this.situationsData].sort(byAttentionPriority));

	view = $derived<SituationView>(this.params.view);
	search = $derived(this.params.q);
	openedRange = $derived<OpenedRange>(this.params.opened);

	situations = $derived(this.view !== "needs-attention" ? this.situationsData : this.sortedSituationsData);

	setView(view: SituationView) {
		this.params.view = view;
	}

	setSearch(search: string) {
		this.params.q = search;
	}

	setOpenedRange(range: OpenedRange) {
		this.params.opened = range;
	}

	clearFilters() {
		this.params.q = "";
		this.params.opened = "" as OpenedRange;
	}

	hasFilters = $derived(this.params.q !== "" || this.params.opened !== ("" as OpenedRange));

	refresh = () => {
		this.query.refetch();
	};
}

const ctx = new Context<SituationsDashboardController>("SituationsDashboardController");
export const initSituationsDashboardController = () => ctx.set(new SituationsDashboardController());
export const useSituationsDashboardController = () => ctx.get();
