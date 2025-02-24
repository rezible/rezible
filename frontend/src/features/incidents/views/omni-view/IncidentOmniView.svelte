<script lang="ts">
	import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { createQuery } from "@tanstack/svelte-query";
	import { getIncidentOptions, getRetrospectiveForIncidentOptions } from "$lib/api";
	import IncidentContentView from "./IncidentContentView.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	type Props = {
		incidentId: string;
		viewParam: IncidentViewRouteParam;
	};
	const { incidentId, viewParam }: Props = $props();

	const incQuery = createQuery(() => getIncidentOptions({ path: { id: incidentId } }));
	const incident = $derived(incQuery.data?.data);

	const retroQuery = createQuery(() => getRetrospectiveForIncidentOptions({ path: { id: incidentId } }));
	const retrospective = $derived(retroQuery.data?.data);

	const currRoute = $derived(viewParam || "");

	const title = $derived(incident?.attributes.title);

	setPageBreadcrumbs(() => [
		{ label: "Incidents", href: "/incidents" },
		{ label: title, href: `/incidents/${incidentId}` },
	]);
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<div class="w-40 h-fit border p-2 bg-surface-200 flex flex-col gap-2 overflow-y-auto">
		{#snippet navMenuItem(label: string, route: string)}
			{@const active = route === currRoute}
			<a href="/incidents/{incidentId}/{route}">
				<div
					class="p-2 rounded border"
					class:border-r-4={active}
					class:bg-primary-600={active}
					class:text-primary-content={active}
				>
					<span>{label}</span>
				</div>
			</a>
		{/snippet}

		<div class="flex flex-col gap-1">
			<span class="text-surface-content/75">Details</span>
			{@render navMenuItem("Overview", "")}
		</div>
		<div class="flex flex-col gap-1">
			<span class="text-surface-content/75">Retrospective</span>
			<!-- todo: check if analysis enabled -->
			{@render navMenuItem("System Analysis", "analysis")}
			{@render navMenuItem("Report", "report")}
		</div>
	</div>

	{#if incident && retrospective}
		<IncidentContentView {incident} {retrospective} {viewParam} />
	{:else}
		<div class="flex-1 min-h-0 overflow-y-auto border p-2">
			<div class="flex flex-col items-center justify-center h-full">
				<div class="text-surface-content/75">Loading...</div>
			</div>
		</div>
	{/if}
</div>
