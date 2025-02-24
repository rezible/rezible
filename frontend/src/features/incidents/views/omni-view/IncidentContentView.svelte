<script lang="ts">
	import { incidentCtx, retrospectiveCtx } from "$features/incidents/lib/context.ts";
	import type { Incident, Retrospective } from "$lib/api";
	import type { IncidentViewRouteParam } from "$src/params/incidentView";

	import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
	import IncidentRetrospective from "$src/features/incidents/components/retrospective/Retrospective.svelte";

	type Props = {
		incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
	};
	const { incident, retrospective, viewParam }: Props = $props();

	incidentCtx.set(incident);
	retrospectiveCtx.set(retrospective);
</script>

{#if viewParam === undefined}
	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		<IncidentOverview />
	</div>
{:else}
	<IncidentRetrospective view={viewParam} />
{/if}