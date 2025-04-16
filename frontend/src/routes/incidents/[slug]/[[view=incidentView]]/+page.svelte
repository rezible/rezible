<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam } from "$src/params/incidentView.js";
	import IncidentView from "$features/incidents/views/incident/IncidentView.svelte";

	const { data } = $props();

	const view = $derived(convertIncidentViewParam(page.params.view));
</script>

<IncidentView incidentId={data.id} {view}/>
