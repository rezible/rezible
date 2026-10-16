<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import * as AlertDialog from "$components/ui/alert-dialog";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import type { ConfirmQuestion, ImpliedProposalView } from "$features/system/lib/catalogue/proposals";
	import { watch } from "runed";
	import { useCatalogueController } from "../controller.svelte";

	const controller = useCatalogueController();

	// Keep the last question while the dialog animates closed; a newly selected proposal replaces it.
	let question = $state.raw<ConfirmQuestion>();
	let implies = $state.raw<ImpliedProposalView[]>([]);
	watch(
		() => [controller.confirmCandidateId, controller.confirmCandidateQuestion] as const,
		([candidateId, candidateQuestion]) => {
			if (!candidateId) {
				return;
			}
			question = candidateQuestion;
			implies = controller.confirmCandidateImplies;
		}
	);
</script>

<AlertDialog.Root
	bind:open={
		() => !!controller.confirmCandidateId,
		(open) => {
			if (!open) {
				controller.cancelConfirm();
			}
		}
	}
>
	<AlertDialog.Content>
		{#if question}
			<AlertDialog.Header>
				<AlertDialog.Title>{question.title}</AlertDialog.Title>
				<AlertDialog.Description>{question.consequence}</AlertDialog.Description>
			</AlertDialog.Header>

			{#if implies.length > 0}
				<div class="space-y-1 text-sm">
					<p class="font-medium">This also settles:</p>
					<ul class="list-disc space-y-0.5 pl-5">
						{#each implies as implied (implied.key)}
							<li>
								{implied.values}
								{#if implied.rejection}
									<span class="text-muted-foreground">({implied.rejection})</span>
								{/if}
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		{:else if controller.confirmCandidateLoading}
			<AlertDialog.Header>
				<AlertDialog.Title>Confirm match</AlertDialog.Title>
			</AlertDialog.Header>
			<LoadingIndicator />
		{:else}
			<AlertDialog.Header>
				<AlertDialog.Title>Confirm match</AlertDialog.Title>
			</AlertDialog.Header>
			<InlineAlert error={controller.confirmCandidateLoadError} dismissable={false} />
		{/if}

		{#if controller.confirmError}
			<InlineAlert error={controller.confirmError} dismissable={false} />
		{/if}

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={controller.confirming}>Cancel</AlertDialog.Cancel>
			<Button disabled={!controller.canMerge} onclick={controller.confirm}>
				{#if controller.confirming}
					<Spinner />
				{/if}
				Merge
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
