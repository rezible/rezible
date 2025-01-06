<script lang="ts">
    import IncidentTimeline from "./timeline/IncidentTimeline.svelte";
    import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
</script>

<div class="h-full w-full">
	<div style:height="40%">
		<IncidentTimeline />
	</div>

	<div style:height="60%">
		<SystemDiagram />
	</div>
</div>
