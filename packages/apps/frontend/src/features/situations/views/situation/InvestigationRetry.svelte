<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { useSituationController } from "./controller.svelte";

	const controller = useSituationController();
	const retrying = $derived(controller.retryMutation.isPending || controller.retryReconciling);
	const visible = $derived(
		controller.retryAvailable || !!controller.retryMessage || !!controller.retryError
	);
</script>

{#if visible}
	<div class="flex flex-col items-start gap-2">
		{#if controller.retryAvailable && controller.retryNeedsRefresh}
			<Button
				variant="outline"
				size="sm"
				disabled={controller.retryReconciling}
				onclick={controller.refreshRetryStatus}
			>
				{#if controller.retryReconciling}
					<Spinner data-icon="inline-start" />
					Refreshing status
				{:else}
					Refresh status
				{/if}
			</Button>
		{:else if controller.retryAvailable}
			<Button
				variant="outline"
				size="sm"
				disabled={controller.retryDisabled}
				onclick={controller.retryLatestFailedTurn}
			>
				{#if retrying}
					<Spinner data-icon="inline-start" />
					Retrying
				{:else}
					Retry investigation
				{/if}
			</Button>
		{/if}
		{#if controller.retryMessage}
			<p class="text-sm text-muted-foreground">{controller.retryMessage}</p>
		{/if}
		{#if controller.retryError}
			<Alert.Root variant="destructive">
				<Alert.Title>Retry status needs attention</Alert.Title>
				<Alert.Description>{controller.retryError}</Alert.Description>
			</Alert.Root>
		{/if}
	</div>
{/if}
