<script lang="ts">
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import AlertIncidentsFilters from "./AlertIncidentsFilters.svelte";
	import { AlertIncidentsViewController } from "./alertIncidentsViewController.svelte";
	import type { Incident } from "$lib/api";

	const incState = new AlertIncidentsViewController();
</script>

{#snippet incidentListItem(inc: Incident)}
	<a href="/incidents/{inc.id}">
		<span>{inc.attributes.title}</span>
	</a>
{/snippet}

<div class="w-full h-full flex flex-col gap-2">
	<AlertIncidentsFilters bind:rosterId={incState.rosterId} />

	<div class="flex-1 flex flex-col gap-1 border">
		<!--LoadingQueryWrapper query={incState.query}>
			{#snippet view(incidents: Incident[])}
				{#each incidents as inc}
					{@render incidentListItem(inc)}
				{:else}
					<div class="p-2">
						<span>No results</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper-->
	</div>

	<!-- <Pagination {...incState.paginator.paginationProps} /> -->
</div>
