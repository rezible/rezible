<script lang="ts">
	import type { OncallEvent } from "$src/lib/api";
	import { mdiAlarmBell, mdiAlert, mdiCircleMedium, mdiDotsVertical, mdiFire } from "@mdi/js";
	import { Button, Checkbox, Icon, Menu, MenuItem, Toggle } from "svelte-ux";

	type Props = {
		event: OncallEvent;
		checked: boolean;
		onToggleChecked: () => void;
	}
	let { event, checked, onToggleChecked }: Props = $props();

	const eventIcon = $derived.by(() => {
		switch (event.kind) {
			case "incident": return mdiFire;
			case "alert": return mdiAlert;
		}
		return mdiCircleMedium;
	})
</script>

<div class="grid grid-cols-subgrid col-span-full hover:bg-surface-100/50 h-16 p-2">
	<div class="grid place-self-center">
		<Checkbox {checked} on:change={onToggleChecked} />
	</div>
	<div class="grid place-items-center">
		<Icon data={eventIcon} />
	</div>
	<div>
		<span>{event.title}</span>
	</div>
	<div class="grid place-items-center">
		<Toggle let:on={open} let:toggle let:toggleOff>
			<Button icon={mdiDotsVertical} iconOnly size="sm" on:click={toggle}>
				<Menu {open} on:close={toggleOff} placement="bottom-end">
					<MenuItem>Edit</MenuItem>
					<MenuItem class="text-danger">Delete</MenuItem>
				</Menu>
			</Button>
		</Toggle>
	</div>
</div>