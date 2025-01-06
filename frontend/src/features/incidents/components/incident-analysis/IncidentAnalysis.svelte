<script lang="ts">
    import { onMount } from "svelte";
    import { watch } from "runed";
	import type { Incident } from "$lib/api";
    import IncidentTimeline from "./timeline/IncidentTimeline.svelte";
    import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
    import { data } from "./data.svelte";

	type Props = {
		incident: Incident; 
	}
	const { incident }: Props = $props();
	
	const incidentId = $derived(incident.id);

	watch(() => incidentId, (id: string) => {data.setIncidentId(id)});
	// 	onMount(() => {return () => {data.cleanup()}});
</script>

<div class="h-full w-full">
	<div style:height="40%">
		<IncidentTimeline />
	</div>

	<div style:height="60%">
		<SystemDiagram />
	</div>
</div>
