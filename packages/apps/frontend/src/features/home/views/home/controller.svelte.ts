import { Context } from "runed";
import {
	listActivityOptions,
	listInboxItemsOptions,
	listIncidentsOptions,
	listSituationsOptions,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { createInvestigationConclusions } from "$features/situations/lib/conclusions.svelte";
import { activityRow, inboxRow, incidentRow, situationRow, type HomeActivityRow } from "./model";

const PREVIEW_PAGE_SIZE = 5;
const ACTIVITY_PAGE_SIZE = 8;

class HomeController {
	private inboxPage = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listInboxItemsOptions({
				query: { page: 1, pageSize: PREVIEW_PAGE_SIZE, state: "open", scope: "mine" },
			}),
	});
	inboxQuery = $derived(this.inboxPage.query);

	private incidentsPage = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listIncidentsOptions({
				query: { page: 1, pageSize: PREVIEW_PAGE_SIZE, responseStates: ["started", "mitigated"] },
			}),
	});
	incidentsQuery = $derived(this.incidentsPage.query);

	private situationsPage = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listSituationsOptions({
				query: { page: 1, pageSize: PREVIEW_PAGE_SIZE, stage: ["candidate", "raised"] },
			}),
	});
	situationsQuery = $derived(this.situationsPage.query);

	private activityPage = createPaginatedQuery({
		source: "local",
		queryOptions: () => listActivityOptions({ query: { page: 1, pageSize: ACTIVITY_PAGE_SIZE } }),
	});
	activityQuery = $derived(this.activityPage.query);

	private situations = $derived(this.situationsQuery.data?.data ?? []);
	private conclusions = createInvestigationConclusions(() =>
		this.situations.map((situation) => situation.attributes.investigation?.investigation.id)
	);

	inboxRows = $derived((this.inboxQuery.data?.data ?? []).map((item) => inboxRow(item)));
	incidentRows = $derived((this.incidentsQuery.data?.data ?? []).map(incidentRow));
	situationRows = $derived(this.buildSituationRows());
	activityGroups = $derived(this.buildActivityGroups());

	inboxTotal = $derived(this.inboxQuery.data?.pagination.total);
	incidentsTotal = $derived(this.incidentsQuery.data?.pagination.total);
	situationsTotal = $derived(this.situationsQuery.data?.pagination.total);

	private buildSituationRows() {
		return this.situations.map((situation) => {
			const investigationId = situation.attributes.investigation?.investigation.id;
			return situationRow(situation, this.conclusions.conclusionFor(investigationId));
		});
	}

	private buildActivityGroups() {
		const rows = (this.activityQuery.data?.data ?? []).map((item) => activityRow(item));
		const groups: { label: HomeActivityRow["day"]; rows: HomeActivityRow[] }[] = [
			{ label: "Today", rows: rows.filter((row) => row.day === "Today") },
			{ label: "Earlier", rows: rows.filter((row) => row.day === "Earlier") },
		];
		return groups.filter((group) => group.rows.length > 0);
	}
}

const ctx = new Context<HomeController>("HomeController");
export const initHomeController = () => ctx.set(new HomeController());
export const useHomeController = () => ctx.get();
