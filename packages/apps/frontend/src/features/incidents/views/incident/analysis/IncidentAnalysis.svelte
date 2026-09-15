<script lang="ts">
	import { Button } from "$components/ui/button";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { initIncidentAnalysisController } from "./controller.svelte";
	import { SystemAnalysis } from "$components/system-analysis";
	import { useIncidentView } from "$features/incidents/views/incident/controller.svelte";
	import IncidentTimeline from "./incident-timeline/IncidentTimeline.svelte";
	import EventDialog from "./incident-timeline/event-dialog/EventDialog.svelte";
	import AnalysisInspector from "./inspector/AnalysisInspector.svelte";

	const viewController = useIncidentView();
	const incAttrs = $derived(viewController.incident?.attributes);
	const controller = initIncidentAnalysisController();
	const retroQuery = $derived(viewController.retrospectiveQuery);
</script>

<div class="flex h-full min-h-0 min-w-0 flex-col gap-3 overflow-hidden">
	{#if incAttrs?.retrospectiveId && viewController.retrospectiveQuery.isPending}
		<p role="status" class="p-6 text-sm text-muted-foreground">Loading analysis...</p>
	{:else if retroQuery.error && !viewController.retrospective}
		<div role="alert" class="p-4">
			<ErrorAlert error={retroQuery.error} />
			<Button variant="outline" onclick={() => retroQuery.refetch()}>Retry analysis</Button>
		</div>
	{:else if !controller.systemAnalysis.analysisId}
		<p class="p-6 text-sm text-muted-foreground">No analysis is associated with this incident.</p>
	{:else}
		<div class="flex shrink-0 flex-wrap items-center gap-2">
			<Button variant="outline" size="sm" onclick={controller.systemAnalysis.refreshAll}>
				Refresh analysis
			</Button>
			<Button
				variant="outline"
				size="sm"
				bind:ref={controller.fallbackFocus}
				disabled={!controller.hasSelection}
				onclick={controller.openInspector}
			>
				Inspect selection
			</Button>
		</div>

		<div class="min-h-0 flex-1">
			<SystemAnalysis>
				{#snippet inspector()}
					<AnalysisInspector />
				{/snippet}
			</SystemAnalysis>
		</div>

		<div class="relative h-[35%] min-h-44 shrink-0">
			<IncidentTimeline />
		</div>
	{/if}
</div>

<EventDialog />
