<script lang="ts">
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import AlertIncidentsFilters from "./AlertIncidentsFilters.svelte";
	import { AlertIncidentsViewController } from "./alertIncidentsViewController.svelte";
	import type { AlertIncidentLink } from "$lib/api";
	import { resolve } from "$app/paths";

	const incState = new AlertIncidentsViewController();
</script>

{#snippet incidentListItem(link: AlertIncidentLink)}
	<a href={resolve(`/incidents/${link.attributes.incidentId}`)}>
		<span>{link.attributes.description || link.attributes.incidentId}</span>
	</a>
{/snippet}

<div class="w-full h-full flex flex-col gap-2">
	<AlertIncidentsFilters bind:rosterId={incState.rosterId} />

	<div class="flex-1 flex flex-col gap-1 border">
		<LoadingQueryWrapper query={incState.query}>
			{#snippet view(links: AlertIncidentLink[])}
				{#each links as link (link.id)}
					{@render incidentListItem(link)}
				{:else}
					<div class="p-2">
						<span>No results</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
