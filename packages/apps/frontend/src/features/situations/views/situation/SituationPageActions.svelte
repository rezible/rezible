<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as DropdownMenu from "$components/ui/dropdown-menu";
	import { Spinner } from "$components/ui/spinner";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import RiMoreLine from "remixicon-svelte/icons/more-line";
	import type { SituationController } from "./controller.svelte";

	type Props = { controller: SituationController };
	const { controller }: Props = $props();

	const actions = $derived(controller.actions);
	const pending = $derived(controller.actionPending);
	const hasMoreActions = $derived(actions.mute || actions.merge || actions.keepOpen || actions.close);
</script>

{#snippet raiseButton(label: string)}
	<Button size="sm" disabled={pending} onclick={controller.raise}>
		{#if controller.raiseMutation.isPending}
			<Spinner data-icon="inline-start" />
		{/if}
		{label}
	</Button>
{/snippet}

{#if actions.raise}
	{@render raiseButton("Raise")}
{/if}
{#if actions.startInvestigation}
	{@render raiseButton("Start investigation")}
{/if}
{#if actions.unmute}
	<Button size="sm" variant="outline" disabled={pending} onclick={controller.unmute}>Unmute</Button>
{/if}
{#if actions.release && controller.holdUntil}
	<span class="text-xs text-muted-foreground">
		Held until
		<Timestamp value={controller.holdUntil} />
	</span>
	<Button size="sm" variant="outline" disabled={pending} onclick={controller.release}>Release</Button>
{/if}

{#if hasMoreActions}
	<DropdownMenu.Root>
		<DropdownMenu.Trigger>
			{#snippet child({ props })}
				<Button {...props} size="icon-sm" variant="outline" aria-label="More situation actions">
					<RiMoreLine />
				</Button>
			{/snippet}
		</DropdownMenu.Trigger>
		<DropdownMenu.Content align="end" class="w-48">
			<DropdownMenu.Group>
				{#if actions.mute}
					<DropdownMenu.Item disabled={pending} onSelect={controller.openMuteDialog}>
						Mute…
					</DropdownMenu.Item>
				{/if}
				{#if actions.merge}
					<DropdownMenu.Item disabled={pending} onSelect={controller.openMergeDialog}>
						Merge into…
					</DropdownMenu.Item>
				{/if}
				{#if actions.keepOpen}
					<DropdownMenu.Item disabled={pending} onSelect={controller.keepOpen}>
						Keep open
					</DropdownMenu.Item>
				{/if}
				{#if actions.close}
					<DropdownMenu.Item disabled={pending} onSelect={controller.openCloseDialog}>
						Close
					</DropdownMenu.Item>
				{/if}
			</DropdownMenu.Group>
		</DropdownMenu.Content>
	</DropdownMenu.Root>
{/if}
