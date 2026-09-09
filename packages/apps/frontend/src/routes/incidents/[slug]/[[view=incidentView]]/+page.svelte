<script lang="ts">
	import type { PageProps } from "./$types";
	import { convertIncidentViewParam } from "$params/incidentView";
	import IncidentView from "$features/incidents/views/incident";

	const { data, params }: PageProps = $props();
</script>

<IncidentView slug={data.slug} param={convertIncidentViewParam(params.view)} />
