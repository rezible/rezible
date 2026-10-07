<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { cn } from "$lib/utils";
	import { statusToneTextClass } from "$components/common/status-badge/status";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();

	const notice = $derived(controller.updateNotice);
</script>

{#if notice}
	<div role="status" class="flex items-start gap-3 rounded-lg border bg-card p-4">
		<notice.status.icon
			class={cn("mt-0.5 size-5 shrink-0", statusToneTextClass[notice.status.tone])}
			aria-hidden="true"
		/>
		<div class="flex min-w-0 flex-1 flex-col gap-3 text-sm">
			<p>
				{notice.message}
				{#if notice.at}
					<Timestamp value={notice.at} format="clock" />
				{/if}
			</p>
			{#if notice.pendingMessage}
				<p class="text-muted-foreground">{notice.pendingMessage}</p>
			{/if}
			{#if controller.updateError}
				<Alert.Root variant="destructive">
					<Alert.Title>Update could not start</Alert.Title>
					<Alert.Description>{controller.updateError}</Alert.Description>
				</Alert.Root>
			{/if}
			{#if notice.offerUpdate}
				<div>
					<Button
						variant="outline"
						size="sm"
						disabled={controller.updateMutation.isPending}
						onclick={controller.updateNow}
					>
						{#if controller.updateMutation.isPending}
							<Spinner data-icon="inline-start" />
							Updating
						{:else}
							Update now
						{/if}
					</Button>
				</div>
			{/if}
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
