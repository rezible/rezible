<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import { Spinner } from "$components/ui/spinner";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { followUpPendingStatus, provisionalStatus } from "$features/situations/lib/status";
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import { useSituationOverviewController } from "./controller.svelte";

	const controller = useSituationOverviewController();
	const understanding = $derived(controller.understanding);
</script>

{#snippet skeletonLines()}
	<div class="flex flex-col gap-2" aria-label="Loading current understanding">
		<Skeleton class="h-5 w-4/5" />
		<Skeleton class="h-5 w-3/5" />
	</div>
{/snippet}

{#snippet investigationLink(href: string)}
	<Button variant="link" size="inline" {href}>
		View investigation
		<RiArrowRightLine aria-hidden="true" />
	</Button>
{/snippet}

{#snippet retryLine(message: string, retry: () => void)}
	<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
		<span>{message}</span>
		<Button variant="outline" size="sm" onclick={retry}>Retry</Button>
	</div>
{/snippet}

{#snippet refreshFailedLine(retry: () => void)}
	<div role="status" class="flex items-center gap-3 text-xs text-muted-foreground">
		<span>Refresh failed. Showing previously loaded details.</span>
		<Button variant="ghost" size="sm" onclick={retry}>Retry</Button>
	</div>
{/snippet}

<section aria-labelledby="understanding-title" class="flex flex-col gap-4 rounded-lg border bg-card p-5">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<h2 id="understanding-title" class="region-label">Current understanding</h2>
		{#if understanding.kind === "investigation"}
			<div class="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
				<StatusBadge status={understanding.run} variant="inline" />
				{#if understanding.report.kind === "published"}
					<span>
						Published
						<Timestamp value={understanding.report.attributes.createdAt} />
					</span>
				{/if}
			</div>
		{/if}
	</div>

	{#if understanding.kind === "loading"}
		{@render skeletonLines()}
	{:else if understanding.kind === "none"}
		<p class="text-sm text-muted-foreground">No investigation has been run for this situation.</p>
		<div>
			<Button disabled={controller.startPending} onclick={controller.startInvestigation}>
				{#if controller.startPending}
					<Spinner data-icon="inline-start" />
					Starting…
				{:else}
					Start investigation
				{/if}
			</Button>
		</div>
	{:else if understanding.kind === "unavailable"}
		<p class="text-sm text-muted-foreground">The investigation is unavailable.</p>
	{:else if understanding.kind === "error"}
		{@render retryLine("The investigation could not be loaded.", controller.retryInvestigation)}
	{:else}
		{@const report = understanding.report}
		{#if report.kind === "published"}
			{#if report.summary}
				<p class="max-w-[65ch] text-lg leading-7 font-medium">{report.summary}</p>
			{:else}
				<p class="text-sm text-muted-foreground">This report has no summary. Read the full report.</p>
			{/if}
			{#if understanding.refreshFailed}
				{@render refreshFailedLine(controller.retryInvestigation)}
			{:else if report.refreshFailed}
				{@render refreshFailedLine(controller.retryReport)}
			{/if}
			<div class="flex flex-wrap items-center gap-x-6 gap-y-2 border-t pt-4 text-sm">
				<Button variant="link" size="inline" href={understanding.href}>
					Read report
					<RiArrowRightLine aria-hidden="true" />
				</Button>
				{#if report.attributes.references.length > 0}
					<span class="text-muted-foreground tabular-nums">
						{report.attributes.references.length}
						{report.attributes.references.length === 1
							? "evidence reference"
							: "evidence references"}
					</span>
				{/if}
				{#if report.provisional}
					<StatusBadge status={provisionalStatus} variant="inline" />
				{/if}
				{#if understanding.pendingFollowUp}
					<StatusBadge status={followUpPendingStatus} variant="inline" />
				{/if}
			</div>
		{:else if report.kind === "loading"}
			{@render skeletonLines()}
		{:else if report.kind === "unavailable"}
			<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
			<div>{@render investigationLink(understanding.href)}</div>
		{:else if report.kind === "error"}
			{@render retryLine("The latest report could not be loaded.", controller.retryReport)}
		{:else}
			<p class="text-sm text-muted-foreground">{understanding.noReportMessage}</p>
			<div>{@render investigationLink(understanding.href)}</div>
		{/if}
	{/if}
</section>
