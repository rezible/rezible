<script lang="ts">
	import { page } from "$app/state";
	import { convertIncidentViewParam } from "$src/params/incidentView.js";
	import { IncidentView, setIncidentViewState } from "$features/incident";

	const { data } = $props();

	const viewParams = $derived({
		slug: data.slug,
		viewRouteParam: convertIncidentViewParam(page.params.view),
	});

	setIncidentViewState(() => viewParams);
</script>

<IncidentView />
