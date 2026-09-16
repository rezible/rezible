<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import * as Empty from "$components/ui/empty";
	import { initSituationInvestigationController } from "./controller.svelte";
	import SituationInvestigationContent from "./SituationInvestigationContent.svelte";
	import StartInvestigationForm from "./StartInvestigationForm.svelte";

	const controller = initSituationInvestigationController();
	const investigationAttributes = $derived(controller.investigation?.attributes);

	const startInvErr = $derived(controller.startInvestigationMutation.error);
</script>

<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4">
	<div class="mx-auto flex max-w-7xl flex-col gap-6">
		<h1 class="text-[28px] leading-9 font-semibold">Investigation</h1>
		{#if !controller.investigation}
			<StartInvestigationForm />
		{/if}

		{#if startInvErr}
			<Alert.Root variant="destructive">
				<Alert.Title>Could not request investigation</Alert.Title>
				<Alert.Description>
					{startInvErr.detail || startInvErr.title}
				</Alert.Description>
			</Alert.Root>
		{/if}

		<div class="min-w-0">
			<article class="flex min-w-0 flex-col gap-5 rounded-lg border bg-card p-4 md:p-6">
				{#if investigationAttributes}
					<SituationInvestigationContent {investigationAttributes} />
				{:else}
					<Empty.Root>
						<Empty.Header>
							<Empty.Title>No investigation yet</Empty.Title>
							<Empty.Description>
								Run an investigation using the form above.
							</Empty.Description>
						</Empty.Header>
					</Empty.Root>
				{/if}
			</article>
		</div>
	</div>
</div>
