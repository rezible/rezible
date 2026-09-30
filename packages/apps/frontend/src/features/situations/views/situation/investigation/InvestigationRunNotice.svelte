<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { cn } from "$lib/utils";
	import { statusToneTextClass } from "$components/common/status-badge/status";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();

	const run = $derived(controller.run);
</script>

{#if controller.runNotice}
	<div role="status" class="flex items-start gap-3 rounded-lg border bg-card p-4">
		<run.icon class={cn("mt-0.5 size-5 shrink-0", statusToneTextClass[run.tone])} aria-hidden="true" />
		<div class="flex min-w-0 flex-1 flex-col gap-3 text-sm">
			<p>{controller.runNotice}</p>
			{#if controller.retryMessage}
				<p class="text-muted-foreground">{controller.retryMessage}</p>
			{/if}
			{#if controller.retryError}
				<Alert.Root variant="destructive">
					<Alert.Title>Retry status needs attention</Alert.Title>
					<Alert.Description>{controller.retryError}</Alert.Description>
				</Alert.Root>
			{/if}
			{#if controller.retryAvailable}
				<div>
					{#if controller.retryNeedsRefresh}
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
					{:else}
						<Button
							size="sm"
							disabled={controller.retryDisabled}
							onclick={controller.retryLatestFailedTurn}
						>
							{#if controller.retryMutation.isPending || controller.retryReconciling}
								<Spinner data-icon="inline-start" />
								Retrying
							{:else}
								Retry failed turn
							{/if}
						</Button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}
