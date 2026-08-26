<script lang="ts">
	import { Button } from "$components/ui/button";
	import Icon from "$components/common/icon/Icon.svelte";
	import { mdiChevronDown, mdiChevronUp } from "@mdi/js";

	import { initIncidentAnalysisController } from "./controller.svelte";
	import { initSystemAnalysisController, SystemAnalysisDiagram } from "$components/system-analysis";
	import { useIncidentView } from "$features/incidents/views/incident/controller.svelte";

	import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";

	const incident = useIncidentView();
	initSystemAnalysisController(() => incident.systemAnalysisId || "");
	initIncidentAnalysisController();

	let hideTimeline = $state(false);
</script>

<div class="flex flex-col gap-2 h-full max-h-full overflow-hidden">
	<div class="relative grow">
		<SystemAnalysisDiagram />
		{@render toggleTimelineButton(false)}
	</div>

	<div class="relative h-[40%]" class:hidden={hideTimeline}>
		<IncidentTimeline />
		{@render toggleTimelineButton(true)}
	</div>
</div>

{#snippet toggleTimelineButton(hide: boolean)}
	<div
		class="absolute left-2 flex items-center h-10"
		class:top-2={hide}
		class:bottom-2={!hide}
		class:hidden={hideTimeline === hide}
	>
		<Button
			color={hide ? "default" : "accent"}
			onclick={() => {
				hideTimeline = hide;
			}}
		>
			{hide ? "Hide" : "Show"} Timeline
			<Icon data={hide ? mdiChevronDown : mdiChevronUp} />
		</Button>
	</div>
{/snippet}
