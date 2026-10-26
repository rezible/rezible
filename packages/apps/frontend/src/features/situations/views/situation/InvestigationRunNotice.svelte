<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { cn } from "$lib/utils";
	import { statusToneTextClass } from "$components/common/status-badge/status";
	import { useSituationController } from "./controller.svelte";

	type Props = { class?: string };
	const { class: className }: Props = $props();

	const controller = useSituationController();

	const notice = $derived(controller.updateNotice);
</script>

{#if notice}
	<div role="status" class={cn("flex items-start gap-3 rounded-lg border bg-card p-4", className)}>
		<notice.status.icon
			class={cn("mt-0.5 size-5 shrink-0", statusToneTextClass[notice.status.tone])}
			aria-hidden="true"
		/>
		<div class="flex min-w-0 flex-1 flex-col gap-3 text-sm">
			<div class="flex flex-col gap-1">
				<p>
					{notice.message}
					{#if notice.at}
						<Timestamp value={notice.at} format="clock" />
					{/if}
				</p>
				{#if notice.failedAt}
					<p class="text-muted-foreground">
						Failed <Timestamp value={notice.failedAt} format="absolute" />
					</p>
				{/if}
			</div>
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
		</div>
	</div>
{/if}
