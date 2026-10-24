<script lang="ts">
	import type { Snippet } from "svelte";

	import type { Situation } from "$lib/api";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import { Spinner } from "$components/ui/spinner";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";

	import { SituationPickerController } from "../situation-picker/controller.svelte";

	type Props = {
		/** Situations already chosen or linked; they are listed but can't be chosen again. */
		linkedIds: string[];
		onSelect: (situation: Situation) => void;
		/** Opens the list in a popover. Without one, the list is shown inline. */
		trigger?: Snippet<[Record<string, unknown>]>;
	};
	const { linkedIds, onSelect, trigger }: Props = $props();

	let open = $state(false);

	// Open and closed situations, muted or not, can be linked to an incident.
	const picker = new SituationPickerController(() => undefined, {
		stages: ["candidate", "raised", "closed"],
	});

	const select = (situation: Situation) => {
		open = false;
		onSelect(situation);
	};
</script>

{#snippet list()}
	<Command.Root shouldFilter={false}>
		<Command.Input
			placeholder="Search situations"
			value={picker.search}
			oninput={(event) => picker.setSearch(event.currentTarget.value)}
		/>
		<Command.List>
			{#if picker.isLoading}
				<div class="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
					<Spinner />
					Loading situations…
				</div>
			{:else if picker.isError}
				<Command.Empty>Could not load situations.</Command.Empty>
			{:else}
				<Command.Empty>No matching situations.</Command.Empty>
				<Command.Group>
					{#each picker.options as option (option.id)}
						{@const linked = linkedIds.includes(option.id)}
						<Command.Item
							value={option.id}
							disabled={linked}
							onSelect={() => select(option.situation)}
						>
							<div class="flex min-w-0 flex-1 flex-col gap-0.5">
								<div class="flex min-w-0 items-center gap-2">
									<span class="truncate text-sm">{option.title}</span>
									<StatusBadge status={option.status} variant="inline" />
									{#if linked}
										<span class="ml-auto text-xs text-muted-foreground">Linked</span>
									{/if}
								</div>
								<span class="truncate text-xs text-muted-foreground">
									{#if option.services.length}
										{option.services.join(", ")} ·
									{/if}
									opened
									<Timestamp value={option.openedAt} />
								</span>
							</div>
						</Command.Item>
					{/each}
				</Command.Group>
			{/if}
		</Command.List>
	</Command.Root>
{/snippet}

{#if trigger}
	<Popover.Root bind:open>
		<Popover.Trigger>
			{#snippet child({ props })}
				{@render trigger(props)}
			{/snippet}
		</Popover.Trigger>

		<Popover.Content class="w-96 p-0" align="start">
			{@render list()}
		</Popover.Content>
	</Popover.Root>
{:else}
	<div class="rounded-md border">
		{@render list()}
	</div>
{/if}
