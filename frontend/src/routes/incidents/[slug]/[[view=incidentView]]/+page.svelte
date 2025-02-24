<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam } from "$src/params/incidentView";
	import IncidentOmniView from "$features/incidents/views/omni-view/IncidentOmniView.svelte";

	const { data } = $props();

	const viewParam = $derived(convertIncidentViewParam(page.params.view));
</script>

{#key data.slug}
	<IncidentOmniView incidentId={data.slug} {viewParam} />
{/key}
