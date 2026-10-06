<script lang="ts">
	import * as Empty from "$components/ui/empty";
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import { Spinner } from "$components/ui/spinner";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import SituationSourceSheet from "$features/situations/components/situation-source-sheet/SituationSourceSheet.svelte";
	import { initSituationInvestigationController } from "./controller.svelte";
	import InvestigationRunNotice from "./InvestigationRunNotice.svelte";
	import InvestigationOutputList from "./InvestigationOutputList.svelte";
	import InvestigationReportSection from "./InvestigationReportSection.svelte";
	import InvestigationContext from "./InvestigationContext.svelte";
	import SituationInvestigationQuestions from "./SituationInvestigationQuestions.svelte";

	const controller = initSituationInvestigationController();
	const investigationAttributes = $derived(controller.investigationAttributes);
</script>

{#snippet context()}
	<InvestigationContext />
{/snippet}

{#snippet headline()}
	<header class="flex flex-col gap-2">
		<h1 class="text-[28px] leading-9 font-semibold tracking-tight">Investigation</h1>
		<p class="text-sm text-muted-foreground">
			Evidence, findings and hypotheses gathered for this situation.
		</p>
	</header>
{/snippet}

{#if !controller.investigationId}
	<PageCanvas width="reading">
		<Empty.Root>
			<Empty.Header>
				<Empty.Title>No investigation yet</Empty.Title>
				<Empty.Description>
					Start an investigation to gather evidence, findings and hypotheses for this situation.
				</Empty.Description>
			</Empty.Header>
			{#if controller.investigationOffer === "raise" || controller.investigationOffer === "start"}
				<Empty.Content>
					<Button disabled={controller.actionPending} onclick={controller.startInvestigation}>
						{#if controller.startPending}
							<Spinner data-icon="inline-start" />
							Starting…
						{:else if controller.investigationOffer === "raise"}
							Raise and investigate
						{:else}
							Start investigation
						{/if}
					</Button>
				</Empty.Content>
			{/if}
		</Empty.Root>
	</PageCanvas>
{:else if controller.investigationUnavailable}
	<PageCanvas width="reading">
		<Empty.Root>
			<Empty.Header>
				<Empty.Title>Investigation unavailable</Empty.Title>
				<Empty.Description>
					Investigation details could not be found or are no longer accessible.
				</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	</PageCanvas>
{:else if investigationAttributes}
	<PageCanvas {context} contextLabel="Investigation context">
		<div class="flex flex-col gap-4">
			{@render headline()}
			{#if controller.investigationQuery.isError}
				<div role="status" class="flex items-center gap-3 text-xs text-muted-foreground">
					<span>Refresh failed. Showing previously loaded investigation details.</span>
					<Button variant="ghost" size="sm" onclick={() => controller.investigationQuery.refetch()}>
						Retry
					</Button>
				</div>
			{/if}
			<InvestigationRunNotice />
		</div>

		{#if controller.conclusion}
			<section aria-labelledby="conclusion-title" class="flex flex-col gap-2">
				<h2 id="conclusion-title" class="region-label">Conclusion</h2>
				<p class="max-w-[65ch] text-lg leading-7 font-medium">{controller.conclusion}</p>
			</section>
		{/if}

		<section aria-labelledby="findings-title" class="flex flex-col gap-3">
			<div class="flex flex-col gap-1">
				<SectionHeading id="findings-title" title="Findings" count={controller.findings.length} />
				<p class="text-sm text-muted-foreground">
					Observations the investigation established, with the evidence they cite.
				</p>
			</div>
			<InvestigationOutputList kind="findings" />
		</section>

		<section aria-labelledby="hypotheses-title" class="flex flex-col gap-3">
			<div class="flex flex-col gap-1">
				<SectionHeading
					id="hypotheses-title"
					title="Hypotheses"
					count={controller.hypotheses.length}
				/>
				<p class="text-sm text-muted-foreground">
					Possible explanations and how well the evidence supports them.
				</p>
			</div>
			<InvestigationOutputList kind="hypotheses" />
		</section>

		<InvestigationReportSection />
		<SituationInvestigationQuestions />
	</PageCanvas>
{:else if controller.investigationQuery.isError}
	<PageCanvas width="reading">
		<div role="alert" class="flex flex-col items-start gap-3 text-sm">
			<p>Investigation details could not be loaded.</p>
			<Button variant="outline" size="sm" onclick={() => controller.investigationQuery.refetch()}>
				Retry
			</Button>
		</div>
	</PageCanvas>
{:else}
	<PageCanvas>
		{@render headline()}
		<div class="flex flex-col gap-3" aria-label="Loading investigation">
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
		</div>
	</PageCanvas>
{/if}

<SituationSourceSheet target={controller.inspection.target} onClose={controller.inspection.close} />
