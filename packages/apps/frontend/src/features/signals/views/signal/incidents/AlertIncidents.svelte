<script lang="ts">
	import { resolve } from "$app/paths";
	import type { AlertIncidentLink } from "$lib/api";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { AlertIncidentsViewController } from "./alertIncidentsViewController.svelte";

	const incidents = new AlertIncidentsViewController();
</script>

{#snippet incidentList(links: AlertIncidentLink[])}
	<ul class="divide-y rounded-lg border bg-card">
		{#each links as link (link.id)}
			<li>
				<a
					href={resolve(`/incidents/${link.attributes.incidentId}`)}
					class="block px-4 py-3 text-sm font-medium wrap-anywhere hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-ring"
				>
					{link.attributes.description || link.attributes.incidentId}
				</a>
			</li>
		{/each}
	</ul>
{/snippet}

{#snippet noIncidents()}
	<p class="text-sm text-muted-foreground">No incidents linked to this signal.</p>
{/snippet}

<PageCanvas width="reading">
	<LoadingQueryWrapper
		query={incidents.query}
		view={incidentList}
		empty={noIncidents}
		isEmpty={(links) => links.length === 0}
	/>
</PageCanvas>
