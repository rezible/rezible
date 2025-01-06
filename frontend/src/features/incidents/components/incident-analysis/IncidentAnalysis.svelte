<script lang="ts">
	import type { Incident } from "$lib/api";
    import IncidentTimeline from "./IncidentTimeline.svelte";
    import SystemDiagram from "./SystemDiagram.svelte";
    import { incidentData } from "./incident-data.svelte";

	type Props = {
		incident: Incident; 
	}
	const { incident }: Props = $props();
	
	incidentData.mount(() => incident.id);
</script>

<div class="h-full w-full">
	<div style:height="40%">
		<IncidentTimeline />
	</div>

	<div style:height="60%">
		<SystemDiagram />
	</div>
</div>
