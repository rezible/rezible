<script lang="ts">
	import ClosedReviewBanner from "../review/ClosedReviewBanner.svelte";
	import RetrospectiveEmptyState from "../retrospective-empty-state/RetrospectiveEmptyState.svelte";
	import { Button } from "$components/ui/button";
	import * as Tooltip from "$components/ui/tooltip";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
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
	<ClosedReviewBanner content="analysis" />
	{#if incAttrs?.retrospective?.id && viewController.retrospectiveQuery.isPending}
		<p role="status" class="p-6 text-sm text-muted-foreground">Loading analysis...</p>
	{:else if retroQuery.error && !viewController.retrospective}
		<div role="alert" class="p-4">
			<ErrorAlert error={retroQuery.error} />
			<Button variant="outline" onclick={() => retroQuery.refetch()}>Retry analysis</Button>
		</div>
	{:else if !viewController.incidentRetrospectiveId}
		<RetrospectiveEmptyState />
	{:else if !controller.systemAnalysis.analysisId}
		<p class="p-6 text-sm text-muted-foreground">No analysis is associated with this incident.</p>
	{:else}
		{#if retroQuery.error}
			<ErrorAlert error={retroQuery.error} />
			<Button variant="outline" onclick={() => retroQuery.refetch()}>Retry analysis</Button>
		{/if}
		<SectionHeading id="analysis-title" title="Analysis" class="shrink-0 px-4 pt-3">
			{#snippet actions()}
				<Tooltip.Root>
					<Tooltip.Trigger onclick={controller.systemAnalysis.refreshAll}>
						{#snippet child({ props })}
							<Button {...props} variant="outline" size="sm">Refresh analysis</Button>
						{/snippet}
					</Tooltip.Trigger>
					<Tooltip.Content>Reload the map and timeline entries</Tooltip.Content>
				</Tooltip.Root>
				<Tooltip.Root>
					<Tooltip.Trigger onclick={controller.openInspector}>
						{#snippet child({ props })}
							<Button
								{...props}
								variant="outline"
								size="sm"
								bind:ref={controller.fallbackFocus}
							>
								Map inventory
							</Button>
						{/snippet}
					</Tooltip.Trigger>
					<Tooltip.Content class="max-w-64">
						List the components and relationships in this analysis, with details of the selection
					</Tooltip.Content>
				</Tooltip.Root>
				{#if controller.entryEditor.canEdit}
					<Button variant="outline" size="sm" onclick={() => controller.entryEditor.setCreating()}>
						Add timeline entry
					</Button>
				{/if}
			{/snippet}
		</SectionHeading>

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
