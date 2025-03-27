<script lang="ts">
	import { mdiCalendar } from "@mdi/js";
	import { Button, Icon, Menu, MenuItem, Toggle, type MenuOption } from "svelte-ux";


	const defaultPeriods = [
		{ value: 7, label: "Last 7 Days" },
		{ value: 30, label: "Last 30 Days" },
		{ value: 90, label: "Last 90 Days" },
	];

	type Props = {
		selected: number;
		periodOptions?: MenuOption[];
	}
	let { 
		selected = $bindable(30),
		periodOptions = defaultPeriods,
	}: Props = $props();

	let selectedOption = $state(periodOptions.find(o => o.value === selected) ?? periodOptions[0]);
	const selectPeriod = (idx: number) => {
		selectedOption = periodOptions[idx];
		selected = selectedOption.value;
	}
</script>

<Toggle let:on={open} let:toggle let:toggleOff>
	<Button on:click={toggle} classes={{root: "flex gap-2 items-center"}}>
		{selectedOption.label}
		<Menu {open} on:close={toggleOff}>
			{#each periodOptions as period, idx}
				<MenuItem on:click={() => selectPeriod(idx)}>{period.label}</MenuItem>
			{/each}
		</Menu>
		<Icon data={mdiCalendar} />
	</Button>
</Toggle>