<script lang="ts">
	import * as Empty from "$components/ui/empty";
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import SituationSourceSheet from "$features/situations/components/situation-source-sheet/SituationSourceSheet.svelte";
	import { initSituationOverviewController } from "./controller.svelte";
	import UnderstandingPanel from "./UnderstandingPanel.svelte";
	import WhyPanel from "./WhyPanel.svelte";
	import ObservationsSection from "./ObservationsSection.svelte";
	import BriefContext from "./BriefContext.svelte";

	const controller = initSituationOverviewController();
	const attributes = $derived(controller.situationAttributes);
	const situationQuery = $derived(controller.situationQuery);
</script>

{#snippet context()}
	<BriefContext />
{/snippet}

{#if controller.situationUnavailable}
	<PageCanvas width="reading">
		<Empty.Root>
			<Empty.Header>
				<Empty.Title>Situation unavailable</Empty.Title>
				<Empty.Description>It may have been removed or you may not have access.</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	</PageCanvas>
{:else if attributes}
	<PageCanvas {context} contextLabel="Situation context">
		<section aria-labelledby="situation-title" class="flex flex-col gap-3">
			<div class="flex flex-wrap items-center gap-x-3 gap-y-1">
				<h1
					id="situation-title"
					class="text-[28px] leading-9 font-semibold tracking-tight wrap-anywhere"
				>
					{attributes.title}
				</h1>
				{#if controller.longRunning}
					<StatusBadge status={controller.longRunning} />
				{/if}
			</div>
			{#if attributes.summary}
				<p
					class="max-w-[72ch] text-[15px] leading-6 whitespace-pre-wrap text-muted-foreground wrap-anywhere"
				>
					{attributes.summary}
				</p>
			{/if}
			<ul class="flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-muted-foreground">
				{#each controller.headerFacts as fact, index (fact.key)}
					<li class="flex items-center gap-x-1.5">
						{#if index > 0}
							<span aria-hidden="true">·</span>
						{/if}
						<span>{fact.label}</span>
						<Timestamp value={fact.at} />
						{#if fact.detail}
							<span>· {fact.detail}</span>
						{/if}
					</li>
				{/each}
			</ul>
		</section>

		<WhyPanel />
		<UnderstandingPanel />
		<ObservationsSection />
	</PageCanvas>
{:else if situationQuery.isError}
	<PageCanvas width="reading">
		<div role="alert" class="flex flex-col items-start gap-3 text-sm">
			<p>The situation could not be loaded.</p>
			<Button variant="outline" size="sm" onclick={() => situationQuery.refetch()}>Retry</Button>
		</div>
	</PageCanvas>
{:else}
	<PageCanvas>
		<div class="flex flex-col gap-3" aria-label="Loading situation">
			<Skeleton class="h-9 w-2/3" />
			<Skeleton class="h-5 w-full" />
			<Skeleton class="h-5 w-4/5" />
		</div>
	</PageCanvas>
{/if}

<SituationSourceSheet target={controller.inspection.target} onClose={controller.inspection.close} />
