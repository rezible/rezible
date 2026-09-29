<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();
</script>

<section aria-labelledby="original-question-title" class="flex flex-col gap-2">
	<h2 id="original-question-title" class="text-lg font-semibold">Original question</h2>
	<p class="max-w-[75ch] whitespace-pre-wrap text-sm leading-relaxed wrap-anywhere">
		{controller.investigationAttributes?.query || "Original question unavailable."}
	</p>
</section>

<section aria-labelledby="execution-status-title" class="flex flex-col gap-2">
	<div class="flex flex-wrap items-center gap-3">
		<h2 id="execution-status-title" class="text-lg font-semibold">Execution</h2>
		<Badge variant="secondary">{controller.execution.label}</Badge>
	</div>
	{#if controller.investigationAttributes?.hasPendingWork}
		<p class="text-sm text-muted-foreground">Follow-up work is waiting.</p>
	{/if}
	{#if controller.execution.message}
		<p class="text-sm text-muted-foreground">{controller.execution.message}</p>
	{/if}
	{#if controller.retryMessage}
		<p role="status" class="text-sm text-muted-foreground">{controller.retryMessage}</p>
	{/if}
	{#if controller.retryError}
		<Alert.Root variant="destructive">
			<Alert.Title>Retry status needs attention</Alert.Title>
			<Alert.Description>{controller.retryError}</Alert.Description>
		</Alert.Root>
	{/if}
	{#if controller.retryAvailable}
		{#if controller.retryNeedsRefresh}
			<Button
				variant="outline"
				size="sm"
				disabled={controller.retryReconciling}
				onclick={controller.refreshRetryStatus}
			>
				{#if controller.retryReconciling}
					<Spinner data-icon="inline-start" /> Refreshing status
				{:else}
					Refresh status
				{/if}
			</Button>
		{:else}
			<Button size="sm" disabled={controller.retryDisabled} onclick={controller.retryLatestFailedTurn}>
				{#if controller.retryMutation.isPending || controller.retryReconciling}
					<Spinner data-icon="inline-start" /> Retrying
				{:else}
					Retry failed turn
				{/if}
			</Button>
		{/if}
	{/if}
</section>
