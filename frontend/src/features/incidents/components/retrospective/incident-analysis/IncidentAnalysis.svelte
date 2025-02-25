<script lang="ts">
	import { Button, Icon } from "svelte-ux";
	import { mdiChevronDown, mdiChevronUp } from "@mdi/js";

	import { analysis } from "./analysisState.svelte";

	import SystemDiagram from "./system-diagram/SystemDiagram.svelte";
	import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";
	import ComponentDialog from "./component-dialog/ComponentDialog.svelte";
	import RelationshipDialog from "./relationship-dialog/RelationshipDialog.svelte";

	analysis.setup();

	let hideTimeline = $state(false);
</script>

<div class="flex flex-col gap-2 h-full max-h-full overflow-hidden">
	<div class="relative grow">
		<SystemDiagram />
		{@render floatingTimelineToggleButton(false)}
	</div>

	<div class="relative grow" class:hidden={hideTimeline}>
		<IncidentTimeline />
		{@render floatingTimelineToggleButton(true)}
	</div>
</div>

{#snippet floatingTimelineToggleButton(hide: boolean)}
	<div class="absolute left-2 flex items-center h-10" class:top-2={hide} class:bottom-2={!hide} class:hidden={hideTimeline === hide}>
		<Button color={hide ? "default" : "accent"} variant="fill-light" on:click={() => {hideTimeline = hide}}>
			{hide ? "Hide" : "Show"} Timeline
			<Icon data={hide ? mdiChevronDown : mdiChevronUp} />
		</Button>
	</div>
{/snippet}

<ComponentDialog />

<RelationshipDialog />
