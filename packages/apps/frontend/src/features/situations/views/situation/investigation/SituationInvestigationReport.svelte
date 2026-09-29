<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();
	const report = $derived(controller.reportAttributes);
</script>

<section aria-labelledby="latest-report-title" class="flex flex-col gap-3">
	<h2 id="latest-report-title" class="text-lg font-semibold">Latest report</h2>
	{#if report}
		{#if controller.reportQuery.isError}
			<div role="status" class="flex items-center justify-between gap-3 text-xs text-muted-foreground">
				<span>Refresh failed. Showing the previously loaded report.</span>
				<Button variant="ghost" size="sm" onclick={() => controller.reportQuery.refetch()}>
					Retry
				</Button>
			</div>
		{/if}
		<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
			<time datetime={controller.reportPublishedAt.iso}>
				Published {controller.reportPublishedAt.label}
			</time>
			{#if report.turnStatus === "running"}
				<Badge variant="outline">In progress</Badge>
			{/if}
		</div>
		<p class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
			{report.text || "Report text unavailable."}
		</p>
		<div class="flex flex-col gap-2">
			<h3 class="text-sm font-medium">Evidence references</h3>
			{#if controller.reportReferences.length}
				<ul class="flex flex-wrap gap-x-3 gap-y-1">
					{#each controller.reportReferences as reference, index (reference.id)}
						<li>
							<Button
								variant="link"
								class="h-auto p-0"
								aria-label={`Inspect Evidence ${index + 1}`}
								onclick={(event) =>
									controller.inspection.openEvidence(reference.id, event.currentTarget)}
							>
								Evidence {index + 1}
							</Button>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-muted-foreground">No evidence references.</p>
			{/if}
		</div>
	{:else if controller.reportAccessLost}
		<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
	{:else if controller.reportQuery.isPending}
		<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
			<Spinner /> Loading latest report
		</p>
	{:else if controller.reportQuery.isError && controller.reportQuery.error?.status !== 404}
		<div role="alert" class="flex flex-col items-start gap-3 text-sm">
			<p>The latest report could not be loaded.</p>
			<Button variant="outline" size="sm" onclick={() => controller.reportQuery.refetch()}>
				Retry
			</Button>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">No report was published yet.</p>
	{/if}
</section>
