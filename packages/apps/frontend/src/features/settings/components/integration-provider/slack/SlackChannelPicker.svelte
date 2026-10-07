<script lang="ts">
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import { tick } from "svelte";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import { Spinner } from "$components/ui/spinner";

	import { SlackChannelPickerController } from "./slackChannelPicker.svelte";

	type Props = {
		installationId: string;
		value: string;
		id?: string;
		"aria-invalid"?: boolean;
	};
	let { installationId, value = $bindable(), id, ...aria }: Props = $props();

	const CLEAR_VALUE = "__clear__";

	let triggerRef = $state<HTMLButtonElement>(null!);

	const picker = new SlackChannelPickerController(() => installationId);

	function select(channelId: string) {
		value = channelId;
		picker.open = false;
		tick().then(() => {
			triggerRef?.focus();
		});
	}
</script>

<Popover.Root bind:open={picker.open}>
	<Popover.Trigger bind:ref={triggerRef} {id}>
		{#snippet child({ props })}
			<Button
				variant="outline"
				class="w-full min-w-0 justify-between font-normal"
				{...props}
				{...aria}
				role="combobox"
				aria-expanded={picker.open}
			>
				<span class="truncate">{picker.labelFor(value)}</span>
			</Button>
		{/snippet}
	</Popover.Trigger>

	<Popover.Content class="w-[var(--bits-popover-anchor-width)] p-0">
		<Command.Root>
			<Command.Input placeholder="Filter loaded channels…" />
			<Command.List>
				<Command.Empty>No loaded channels match.</Command.Empty>
				<Command.Group>
					<Command.Item
						value={CLEAR_VALUE}
						keywords={["none", "clear"]}
						onSelect={() => select("")}
					>
						<RiCloseLine />
						No announcement channel
					</Command.Item>
					{#each picker.channels as channel (channel.id)}
						<Command.Item
							value={channel.id}
							keywords={[channel.name]}
							onSelect={() => select(channel.id)}
						>
							#{channel.name}
							{#if value === channel.id}
								<RiCheckLine class="ml-auto" />
							{/if}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>

			{#if picker.loadError}
				<div class="border-t p-2">
					<InlineAlert error={picker.loadError} dismissable={false} />
				</div>
			{:else if !picker.loaded}
				<div
					class="flex items-center justify-center gap-2 border-t py-4 text-sm text-muted-foreground"
				>
					<Spinner />
					Loading channels…
				</div>
			{:else if picker.channels.length === 0 && !picker.hasMore}
				<div class="border-t py-4 text-center text-sm text-muted-foreground">
					No public channels found.
				</div>
			{/if}

			{#if picker.hasMore}
				<div class="border-t p-1">
					<Button
						variant="ghost"
						size="sm"
						class="w-full"
						disabled={picker.loadingMore}
						onclick={picker.loadMore}
					>
						{#if picker.loadingMore}
							<Spinner />
						{/if}
						Load more channels
					</Button>
				</div>
			{/if}
		</Command.Root>
	</Popover.Content>
</Popover.Root>
