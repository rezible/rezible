<script lang="ts">
    import { analysis } from "./analysis.svelte";
	
	import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
    import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";
    import EditComponentDialog from "./edit-component-dialog/EditComponentDialog.svelte";
    import EditRelationshipDialog from "./edit-relationship-dialog/EditRelationshipDialog.svelte";

	analysis.setup();
</script>

<div class="grid grid-rows-2 gap-2 h-full">
	<SystemDiagram />
	<IncidentTimeline />
</div>

<EditComponentDialog />
<EditRelationshipDialog />