<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Alert from "$components/ui/alert";
	import * as Empty from "$components/ui/empty";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SituationInvestigationList from "../SituationInvestigationList.svelte";
	import { useSituationController } from "../controller.svelte";
	import SituationImpactAnalysis from "./SituationImpactAnalysis.svelte";

	const controller = useSituationController();
</script>

<section class="flex min-h-0 min-w-0 flex-1 flex-col gap-4 p-4">
	<header class="flex flex-col gap-2">
		<h1 class="text-[28px] leading-9 font-semibold">Analysis</h1>
		<p class="text-sm text-muted-foreground">
			Systems and recorded entries from the selected investigation.
		</p>
	</header>
	<div class="grid min-h-0 min-w-0 flex-1 gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
		<div class="flex min-h-[28rem] min-w-0 flex-col gap-3">
			{#if controller.selectedInvestigationId}
				<LoadingQueryWrapper query={controller.selectedInvestigationQuery} feedbackOnly />
				{#if controller.selectedInvestigation}
					<SituationImpactAnalysis investigation={controller.selectedInvestigation} />
				{/if}
			{:else}
				<Empty.Root>
					<Empty.Header>
						<Empty.Title>No investigation yet</Empty.Title>
						<Empty.Description>
							Run an investigation to build its system analysis.
						</Empty.Description>
					</Empty.Header>
					<Empty.Content>
						<Button variant="outline" href={controller.investigationsHref}>
							Go to investigations
						</Button>
					</Empty.Content>
				</Empty.Root>
			{/if}
		</div>

		<SituationInvestigationList />
	</div>
</section>
