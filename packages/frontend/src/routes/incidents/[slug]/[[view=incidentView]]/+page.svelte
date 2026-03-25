<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam } from "$params/incidentView";
	import IncidentView from "$features/incidents/views/incident";

	const { data } = $props();
</script>

<IncidentView slug={data.slug} routeParam={convertIncidentViewParam(page.params.view)} />
