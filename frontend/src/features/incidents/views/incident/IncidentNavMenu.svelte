<script lang="ts">
	import type { IncidentViewRouteParam } from '$src/params/incidentView';
	import { useIncidentViewState } from './viewState.svelte';
	
	import { Button } from 'svelte-ux';

	type Props = {
		view?: IncidentViewRouteParam;
	};
	const { view }: Props = $props();

	const viewState = useIncidentViewState();
</script>

<div class="w-40 h-fit p-2 bg-surface-200 flex flex-col gap-2 overflow-y-auto">
	{#snippet navMenuItem(label: string, route?: IncidentViewRouteParam)}
		{@const active = route === view}
		<a href="/incidents/{viewState.incidentId}/{route}">
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
		{#if viewState.retrospective}
			{#if !!viewState.systemAnalysisId}
				{@render navMenuItem("System Analysis", "analysis")}
			{/if}
			{@render navMenuItem("Report", "retrospective")}
		{:else if viewState.retrospectiveNeedsCreating}
			<Button color="accent" variant="fill-light" on:click={() => (viewState.createRetrospectiveDialogOpen = true)}>
				Create
			</Button>
		{:else}
			<span>loading...</span>
		{/if}
	</div>
</div>