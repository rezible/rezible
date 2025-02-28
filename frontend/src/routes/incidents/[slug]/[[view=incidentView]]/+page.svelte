<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam, type IncidentViewRouteParam } from "$src/params/incidentView.js";
	import IncidentOmniView from "$features/incidents/views/incident/IncidentOmniView.svelte";

	const { data } = $props();

	const view = $derived(convertIncidentViewParam(page.params.view));
</script>

<IncidentOmniView incidentId={data.id} {view}/>
