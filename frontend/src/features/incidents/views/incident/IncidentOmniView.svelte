<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button } from "svelte-ux";
	import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { getIncidentOptions, getRetrospectiveForIncidentOptions } from "$lib/api";
	import { appShell, type PageBreadcrumb } from "$features/app/lib/appShellState.svelte";
	import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
	import ContextSidebar from "$features/incidents/components/context-sidebar/ContextSidebar.svelte";
	import PageActions from "./PageActions.svelte";
	import RetrospectiveView from "./RetrospectiveView.svelte";
	import CreateRetrospectiveDialog from "./CreateRetrospectiveDialog.svelte";

	type Props = {
		incidentId: string;
		view?: IncidentViewRouteParam;
	};
	const { incidentId, view }: Props = $props();

	appShell.setPageActions(PageActions, true);

	const isIncidentView = $derived(view === undefined);

	const incQuery = createQuery(() => getIncidentOptions({ path: { id: incidentId } }));
	const incident = $derived(incQuery.data?.data);

	const retroQueryOpts = $derived(getRetrospectiveForIncidentOptions({ path: { id: incidentId } }));
	const retroQuery = createQuery(() => retroQueryOpts);
	const retrospective = $derived(retroQuery.data?.data);

	let createRetroDialogOpen = $state(false);

	const retroNeedsCreating = $derived(retroQuery.isError && retroQuery.error);

	const analysisId = $derived(retrospective?.attributes.systemAnalysisId);

	const incidentBreadcrumb = $derived<PageBreadcrumb>({
		label: incident?.attributes.title,
		href: `/incidents/${incidentId}`,
	});
	const retroBreadcrumb = $derived<PageBreadcrumb>({
		label: view === "analysis" ? "System Analysis" : "Report",
		href: `/incidents/${incidentId}/${view}`,
	});
	const pageCrumbs = $derived<PageBreadcrumb[]>(
		isIncidentView ? [incidentBreadcrumb] : [incidentBreadcrumb, retroBreadcrumb]
	);
	appShell.setPageBreadcrumbs(() => [{ label: "Incidents", href: "/incidents" }, ...pageCrumbs]);
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<div class="w-40 h-fit border p-2 bg-surface-200 flex flex-col gap-2 overflow-y-auto">
		{#snippet navMenuItem(label: string, route?: IncidentViewRouteParam)}
			{@const active = route === view}
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
			{@render navMenuItem("Overview")}
		</div>

		<div class="flex flex-col gap-1">
			<span class="text-surface-content/75">Retrospective</span>
			{#if retroQuery.isSuccess}
				{#if analysisId}
					{@render navMenuItem("System Analysis", "analysis")}
				{/if}
				{@render navMenuItem("Report", "retrospective")}
			{:else if retroNeedsCreating}
				<Button color="accent" variant="fill-light" on:click={() => (createRetroDialogOpen = true)}>
					Create
				</Button>
			{:else}
				<span>loading...</span>
			{/if}
		</div>
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		{#if incident}
			{#key incident.id}
				{#if view === undefined}
					<IncidentOverview {incident} />
				{:else if retrospective}
					<RetrospectiveView {incident} {retrospective} {view} />
				{/if}
			{/key}
		{:else}
			<div class="flex flex-col items-center justify-center h-full">
				<div class="text-surface-content/75">Loading...</div>
			</div>
		{/if}
	</div>

	{#if incident}
		<ContextSidebar />
	{/if}
</div>

{#if incident && retroNeedsCreating}
	<CreateRetrospectiveDialog
		bind:open={createRetroDialogOpen}
		incidentId={incident.id}
		{isIncidentView}
		queryKey={retroQueryOpts.queryKey}
	/>
{/if}
