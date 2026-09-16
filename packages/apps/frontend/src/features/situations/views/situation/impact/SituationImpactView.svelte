<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Empty from "$components/ui/empty";
	import { useSituationController } from "../controller.svelte";
	import SituationImpactAnalysis from "./SituationImpactAnalysis.svelte";

	const controller = useSituationController();
	const situationId = $derived(controller.situationId);
	const investigationAttributes = $derived(controller.investigation?.attributes);
</script>

<section class="flex min-h-0 min-w-0 flex-1 flex-col gap-4 p-4">
	<header class="flex flex-col gap-2">
		<h1 class="text-[28px] leading-9 font-semibold">Analysis</h1>
		<p class="text-sm text-muted-foreground">
			Systems and recorded entries from the selected investigation.
		</p>
	</header>
	<div class="min-h-0 min-w-0 flex-1">
		<div class="flex min-h-[28rem] min-w-0 flex-col gap-3">
			{#if situationId && investigationAttributes}
				<SituationImpactAnalysis {situationId} {investigationAttributes} />
			{:else}
				<Empty.Root>
					<Empty.Header>
						<Empty.Title>No investigation yet</Empty.Title>
						<Empty.Description>
							Run an investigation to build its system analysis.
						</Empty.Description>
					</Empty.Header>
					<Empty.Content>
						<Button variant="outline" href={controller.investigationHref}>
							Go to investigation
						</Button>
					</Empty.Content>
				</Empty.Root>
			{/if}
		</div>
	</div>
</section>
