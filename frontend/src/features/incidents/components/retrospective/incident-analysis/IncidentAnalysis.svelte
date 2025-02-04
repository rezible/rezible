<script lang="ts">
	import { analysis } from "./analysis.svelte";

	import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
	import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";
	import ComponentDialog from "./component-dialog/ComponentDialog.svelte";
	import RelationshipDialog from "./relationship-dialog/RelationshipDialog.svelte";

	analysis.setup();
</script>

<div class="grid grid-rows-2 gap-2 h-full">
	<SystemDiagram />
	<IncidentTimeline />
</div>

<ComponentDialog />

<RelationshipDialog />
