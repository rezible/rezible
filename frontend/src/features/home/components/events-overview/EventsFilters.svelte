<script lang="ts">
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import { mdiCalendarRange, mdiChevronDown } from "@mdi/js";
	import {
		Button,
		Checkbox,
		DateRangeField,
		Field,
		Icon,
		MultiSelectMenu,
		MultiSelectOption,
		type MenuOption,
	} from "svelte-ux";
	import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { v4 as uuidv4 } from "uuid";

	type Props = {
		rosters?: string[];
		actionRequired: boolean;
		dateRange: DateRangeType;
	};
	let { rosters = $bindable(), actionRequired = $bindable(), dateRange = $bindable() }: Props = $props();

	const rosterOptions: MenuOption<string>[] = [
		{ label: "One", value: uuidv4() },
		{ label: "Two", value: uuidv4() },
		{ label: "Three", value: uuidv4() },
		{ label: "Four", value: uuidv4() },
	];
	// let selectedRosters = $state<string[]>([rosterOptions[0].value]);
	const selectedSet = $derived(new Set(rosters));
	const selectedRosterOptions = $derived(rosterOptions.filter((o) => selectedSet.has(o.value)));

	let rosterMenuOpen = $state(false);
	const toggleRosterMenu = () => {
		rosterMenuOpen = !rosterMenuOpen;
	};
</script>

<div class="flex flex-row items-center gap-2">
	<DateRangeField
		label="Date Range"
		labelPlacement="top"
		dense
		classes={{
			field: { root: "gap-0", container: "pl-0 h-8 flex items-center", prepend: "[&>span]:mr-2" },
		}}
		icon={mdiCalendarRange}
		bind:value={dateRange}
	/>

	<Field
		label="Rosters"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0", container: "px-0 h-8 flex items-center", input: "my-0" }}
		let:id
	>
		<Button {id} on:click={toggleRosterMenu} classes={{ root: "h-8" }}>
			<div class="flex gap-2">
				{#each selectedRosterOptions as v, i (v.value)}
					<span class="flex items-center gap-1">
						<Avatar kind="roster" id={v.value} size={14} />
						{v.label + (i < selectedRosterOptions.length - 1 ? "," : "")}
					</span>
				{:else}
					<span>Any</span>
				{/each}
			</div>
			<Icon data={mdiChevronDown} />
			<MultiSelectMenu
				options={rosterOptions}
				value={rosters}
				open={rosterMenuOpen}
				search
				maintainOrder
				placeholder="Filter to roster"
				on:change={(e) => {
					// @ts-expect-error
					rosters = e.detail.value;
				}}
				on:close={toggleRosterMenu}
			>
				<MultiSelectOption
					slot="option"
					let:option
					let:label
					let:checked
					let:indeterminate
					let:onChange
					{checked}
					{indeterminate}
					on:change={onChange}
					classes={{
						checkbox: { label: "label flex-1 flex items-center pl-1 py-2" },
						container: "inline-flex items-center gap-2",
					}}
				>
					<Avatar kind="roster" id={(option as MenuOption<string>).value} size={22} />
					{label}
				</MultiSelectOption>
			</MultiSelectMenu>
		</Button>
	</Field>

	<Field
		label="Action"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0", container: "h-8 flex items-center" }}
		let:id
	>
		<Checkbox {id} bind:checked={actionRequired} classes={{ label: "pl-2" }}>Feedback Requested</Checkbox>
	</Field>
</div>
