<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam } from "$src/params/incidentView.js";
	import IncidentView from "$features/incidents/views/incident";

	const { data } = $props();
</script>

<IncidentView slug={data.slug} viewParam={convertIncidentViewParam(page.params.view)} />
