<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import RichTextView from "$components/rich-text-view/RichTextView.svelte";
	import { provisionalStatus } from "$features/situations/lib/status";
	import CitationChips from "./CitationChips.svelte";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();
	const report = $derived(controller.reportState);
</script>

{#snippet headingActions()}
	{#if report.kind === "published"}
		<span class="text-xs text-muted-foreground">
			Published
			<Timestamp value={report.attributes.createdAt} />
		</span>
		{#if report.provisional}
			<StatusBadge status={provisionalStatus} variant="inline" />
		{/if}
	{/if}
{/snippet}

<section aria-labelledby="report-title" class="flex flex-col gap-4">
	<SectionHeading id="report-title" title="Report" actions={headingActions} />

	{#if report.kind === "published"}
		{#if report.refreshFailed}
			<div role="status" class="flex items-center gap-3 text-xs text-muted-foreground">
				<span>Refresh failed. Showing the previously loaded report.</span>
				<Button variant="ghost" size="sm" onclick={controller.retryReport}>Retry</Button>
			</div>
		{/if}
		{#if report.attributes.text.trim()}
			<RichTextView markdown={report.attributes.text} label="Investigation report" />
		{:else}
			<p class="text-sm text-muted-foreground">Report text unavailable.</p>
		{/if}
		<div class="flex flex-col gap-2">
			<h3 class="region-label">Evidence cited</h3>
			{#if controller.reportCitations.length}
				<CitationChips citations={controller.reportCitations} />
			{:else}
				<p class="text-sm text-muted-foreground">No evidence cited.</p>
			{/if}
		</div>
	{:else if report.kind === "unavailable"}
		<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
	{:else if report.kind === "loading"}
		<div class="flex flex-col gap-2" aria-label="Loading report">
			<Skeleton class="h-4 w-4/5" />
			<Skeleton class="h-4 w-3/5" />
		</div>
	{:else if report.kind === "error"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
			<span>The latest report could not be loaded.</span>
			<Button variant="outline" size="sm" onclick={controller.retryReport}>Retry</Button>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">No report has been published yet.</p>
	{/if}
</section>
