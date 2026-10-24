<script lang="ts">
	import type { Snippet } from "svelte";
	import RiCheckLine from "remixicon-svelte/icons/check-line";

	import type { User } from "$lib/api";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import { Spinner } from "$components/ui/spinner";

	import { MemberPickerController } from "./controller.svelte";

	type Props = {
		/** Chosen members, marked in the list. */
		selectedIds?: string[];
		/** Members who are not offered. */
		excludeIds?: string[];
		/** Keeps the list open after a choice, so several members can be chosen. */
		multiple?: boolean;
		onSelect: (user: User) => void;
		trigger: Snippet<[Record<string, unknown>]>;
	};
	const { selectedIds = [], excludeIds = [], multiple = false, onSelect, trigger }: Props = $props();

	let open = $state(false);

	const picker = new MemberPickerController();
	const users = $derived(picker.users.filter((user) => !excludeIds.includes(user.id)));

	const select = (user: User) => {
		if (!multiple) {
			open = false;
		}
		onSelect(user);
	};
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
			{@render trigger(props)}
		{/snippet}
	</Popover.Trigger>

	<Popover.Content class="w-72 p-0" align="start">
		<Command.Root shouldFilter={false}>
			<Command.Input
				placeholder="Search people"
				value={picker.search}
				oninput={(event) => picker.setSearch(event.currentTarget.value)}
			/>
			<Command.List>
				{#if picker.isLoading}
					<div class="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
						<Spinner />
						Loading people…
					</div>
				{:else if picker.isError}
					<Command.Empty>Could not load people.</Command.Empty>
				{:else}
					<Command.Empty>No matching people.</Command.Empty>
					<Command.Group>
						{#each users as user (user.id)}
							<Command.Item value={user.id} onSelect={() => select(user)}>
								<span class="truncate">{user.attributes.name}</span>
								{#if selectedIds.includes(user.id)}
									<RiCheckLine class="ml-auto" />
								{/if}
							</Command.Item>
						{/each}
					</Command.Group>
				{/if}
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
