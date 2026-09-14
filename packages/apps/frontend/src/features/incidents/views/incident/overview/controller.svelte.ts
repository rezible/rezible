import { page } from "$app/state";
import { tick } from "svelte";
import { Context, watch } from "runed";
import { listIncidentUpdatesOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useIncidentView } from "../controller.svelte";

class IncidentOverviewController {
	incident = useIncidentView();

	paginatedUpdatesQuery = createPaginatedQuery({
		source: "url",
		queryOptions: (pagination) => ({
			...listIncidentUpdatesOptions({ path: { id: this.incident.incidentId }, query: pagination }),
			enabled: !!this.incident.incidentId,
		}),
	});
	updatesQuery = $derived(this.paginatedUpdatesQuery.query);
	updates = $derived(this.updatesQuery.data?.data);

	constructor() {
		watch(
			() => [page.url.hash, this.updatesQuery.data],
			() => {
				if (!page.url.hash.startsWith("#update-") || !this.updatesQuery.data) return;
				const id = page.url.hash.slice(1);
				void tick().then(() => document.getElementById(id)?.scrollIntoView({ block: "start" }));
			}
		);
	}
}

const context = new Context<IncidentOverviewController>("IncidentOverviewController");
export const initIncidentOverviewController = () => context.set(new IncidentOverviewController());
