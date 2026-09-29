<script lang="ts">
	import * as Empty from "$components/ui/empty";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import SituationSourceSheet from "$features/situations/components/situation-source-sheet/SituationSourceSheet.svelte";
	import { initSituationInvestigationController } from "./controller.svelte";
	import SituationInvestigationExecution from "./SituationInvestigationExecution.svelte";
	import SituationInvestigationReport from "./SituationInvestigationReport.svelte";
	import SituationInvestigationQuestions from "./SituationInvestigationQuestions.svelte";

	const controller = initSituationInvestigationController();
	const investigationAttributes = $derived(controller.investigationAttributes);
</script>

<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4">
	<div class="mx-auto flex max-w-4xl flex-col gap-6">
		<h1 class="text-[28px] leading-9 font-semibold">Investigation</h1>

		{#if !controller.investigationId}
			<Empty.Root>
				<Empty.Header>
					<Empty.Title>Investigation unavailable</Empty.Title>
					<Empty.Description>This situation has no linked investigation.</Empty.Description>
				</Empty.Header>
			</Empty.Root>
		{:else if controller.investigationUnavailable}
			<Empty.Root>
				<Empty.Header>
					<Empty.Title>Investigation unavailable</Empty.Title>
					<Empty.Description>
						Investigation details could not be found or are no longer accessible.
					</Empty.Description>
				</Empty.Header>
			</Empty.Root>
		{:else if controller.investigationQuery.isPending && !investigationAttributes}
			<div role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
				<Spinner /> Loading investigation
			</div>
		{:else if controller.investigationQuery.isError && !investigationAttributes}
			<div role="alert" class="flex flex-col items-start gap-3 text-sm">
				<p>Investigation details could not be loaded.</p>
				<Button variant="outline" size="sm" onclick={() => controller.investigationQuery.refetch()}>
					Retry
				</Button>
			</div>
		{:else if investigationAttributes}
			<article class="flex min-w-0 flex-col gap-5 rounded-lg border bg-card p-4 md:p-6">
				{#if controller.investigationQuery.isError}
					<div
						role="status"
						class="flex items-center justify-between gap-3 text-xs text-muted-foreground"
					>
						<span>Refresh failed. Showing previously loaded investigation details.</span>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => controller.investigationQuery.refetch()}
						>
							Retry
						</Button>
					</div>
				{/if}

				<SituationInvestigationExecution />
				<SituationInvestigationReport />
				<SituationInvestigationQuestions />
			</article>
		{/if}
	</div>
</div>

<SituationSourceSheet target={controller.inspection.target} onClose={controller.inspection.close} />
