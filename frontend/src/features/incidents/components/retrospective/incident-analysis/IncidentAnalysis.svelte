<script lang="ts">
	import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
    import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";
    import EditComponentDialog from "./EditComponentDialog.svelte";

	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';
    import { analysis } from "./analysis.svelte";

	analysis.setup();
</script>

<div class="grid grid-rows-2 gap-2 h-full">
	<SystemDiagram />
	<IncidentTimeline />
</div>

<EditComponentDialog />