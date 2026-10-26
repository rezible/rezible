<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import { Spinner } from "$components/ui/spinner";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { followUpPendingStatus, provisionalStatus } from "$features/situations/lib/status";
	import type { NextCheckView } from "$features/situations/lib/brief";
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import InvestigationRetry from "../InvestigationRetry.svelte";
	import InvestigationRunNotice from "../InvestigationRunNotice.svelte";
	import { useSituationOverviewController } from "./controller.svelte";

	type Props = { nextChecks: NextCheckView[] };
	const { nextChecks }: Props = $props();

	const controller = useSituationOverviewController();
	const understanding = $derived(controller.understanding);
</script>

{#snippet skeletonLines()}
	<div class="flex flex-col gap-2" aria-label="Loading current understanding">
		<Skeleton class="h-5 w-4/5" />
		<Skeleton class="h-5 w-3/5" />
	</div>
{/snippet}

{#snippet investigationActions(href: string)}
	<div class="flex flex-wrap items-start gap-3">
		<Button size="sm" {href}>
			View investigation
			<RiArrowRightLine aria-hidden="true" />
		</Button>
		<InvestigationRetry />
	</div>
{/snippet}

{#snippet raiseButton(label: string)}
	<Button disabled={controller.actionPending} onclick={controller.raise}>
		{#if controller.raisePending}
			<Spinner data-icon="inline-start" />
			Starting…
		{:else}
			{label}
		{/if}
	</Button>
{/snippet}

{#snippet retryLine(message: string, retry: () => void)}
	<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
		<span>{message}</span>
		<Button variant="outline" size="sm" onclick={retry}>Retry</Button>
	</div>
{/snippet}

{#snippet nextChecksList()}
	<div class="flex flex-col gap-2">
		<h3 class="text-sm font-semibold">Next checks</h3>
		<ol class="flex list-decimal flex-col gap-2 pl-5 text-sm">
			{#each nextChecks as nextCheck (nextCheck.key)}
				<li>
					<p class="font-medium wrap-anywhere">{nextCheck.check}</p>
					<p class="text-muted-foreground wrap-anywhere">{nextCheck.reason}</p>
				</li>
			{/each}
		</ol>
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
		<h2 id="understanding-title" class="region-label">Rezible's interpretation</h2>
		{#if understanding.kind === "investigation"}
			<div class="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
				{#if understanding.showRunStatus}
					<StatusBadge status={understanding.run} variant="inline" />
				{/if}
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
		{#if understanding.offer === "raise"}
			<p class="text-sm text-muted-foreground">
				Rezible is watching this. It has not been raised, so no investigation has run.
			</p>
			<div>{@render raiseButton("Raise and investigate")}</div>
		{:else if understanding.offer === "start"}
			<p class="text-sm text-muted-foreground">No investigation has been run for this situation.</p>
			<div>{@render raiseButton("Start investigation")}</div>
		{:else if understanding.offer === "muted"}
			<p class="text-sm text-muted-foreground">
				This situation is muted. It keeps gathering evidence but is not raised or investigated.
			</p>
		{:else}
			<p class="text-sm text-muted-foreground">No investigation was run for this situation.</p>
		{/if}
	{:else if understanding.kind === "unavailable"}
		<p class="text-sm text-muted-foreground">The investigation is unavailable.</p>
	{:else if understanding.kind === "error"}
		{@render retryLine("The investigation could not be loaded.", controller.retryInvestigation)}
	{:else}
		{@const report = understanding.report}
		<InvestigationRunNotice class="border-0 bg-transparent p-0" />
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
			{#if nextChecks.length > 0}
				{@render nextChecksList()}
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
				<InvestigationRetry />
			</div>
		{:else if report.kind === "loading"}
			{@render skeletonLines()}
		{:else if report.kind === "unavailable"}
			<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
			{@render investigationActions(understanding.href)}
		{:else if report.kind === "error"}
			{@render retryLine("The latest report could not be loaded.", controller.retryReport)}
		{:else}
			<p class="text-sm text-muted-foreground">{understanding.noReportMessage}</p>
			{@render investigationActions(understanding.href)}
		{/if}
	{/if}
</section>
