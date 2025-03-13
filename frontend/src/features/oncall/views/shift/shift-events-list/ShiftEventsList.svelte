<script lang="ts">
	import { mdiPlus, mdiPhoneAlert, mdiFire } from "@mdi/js";
	import { Icon, Button, Header } from "svelte-ux";
	import type { ShiftEvent } from "$features/oncall/lib/utils";

	type Props = {
		shiftEvents: ShiftEvent[];
	};
	const { shiftEvents }: Props = $props();

	const eventKindIcons: Record<ShiftEvent["eventType"], string> = {
		["incident"]: mdiFire,
		["alert"]: mdiPhoneAlert,
	};
</script>

<div class="h-10 flex w-full gap-4 items-center px-2">
	<Header title="Shift Events" classes={{ root: "w-full", title: "text-xl", container: "flex-1" }}>
		<!--div slot="actions">
			<Button
				color="primary"
				variant="fill"
				on:click={() => {}}
			>
				Filters <Icon data={mdiPlus} />
			</Button>
		</div-->
	</Header>
</div>

<div class="flex-1 min-h-0 flex flex-col gap-4 overflow-y-auto bg-surface-200 p-3">
	{#each shiftEvents as ev}
		{@render eventListItem(ev)}
	{/each}
</div>

{#snippet eventListItem(ev: ShiftEvent)}
	{@const occurredAt = ev.timestamp.toDate()}
	<div class="grid grid-cols-[100px_auto_minmax(0,1fr)] place-items-center border p-2">
		<div class="justify-self-start">
			<span class="flex items-center">
				{occurredAt.toLocaleString()}
			</span>
		</div>

		<div class="items-center static z-10">
			<Icon
				data={eventKindIcons[ev.eventType]}
				classes={{ root: "bg-accent-900 rounded-full p-2 w-auto h-10" }}
			/>
		</div>

		<div class="w-full justify-self-start grid grid-cols-[auto_40px] items-center px-2">
			<div class="leading-none">{ev.eventType}</div>
		</div>

		<div
			class="row-start-3 col-start-3 overflow-y-auto max-h-20 overflow-y-auto border rounded p-2 w-full"
		>
			notes
		</div>
	</div>
{/snippet}
