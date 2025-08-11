<script lang="ts">
	import { mdiCalendarRange, mdiChevronDown } from "@mdi/js";
	import {
		DateRangeField,
		Field,
		MultiSelectMenu,
		SelectField,
		ToggleGroup,
		ToggleOption,
		type MenuOption,
	} from "svelte-ux";
	import { PeriodType } from "@layerstack/utils";
	import Button from "$components/button/Button.svelte";
	import type { DateRangeOption, EventKind, EventsListFiltersState } from "./filters.svelte";
	import RosterSelectField from "$components/roster-select-field/RosterSelectField.svelte";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		filtersState: EventsListFiltersState;
	};
	const { filtersState }: Props = $props();

	const dateRangeOptions: DateRangeOption[] = [
		{ label: "Last 7 Days", value: "7d" },
		{ label: "Last Month", value: "30d" },
		{ label: "Custom", value: "custom" },
	];

	type AnnotationOption = "no" | "any" | "has";
	const annoOptions: MenuOption<AnnotationOption>[] = [
		{value: "any", label: "Any"},
		{value: "has", label: "Yes"},
		{value: "no", label: "No"},
	];
	const annoValue = $derived(filtersState.annotation === undefined ? "any" : (filtersState.annotation ? "yes" : "no"));
	const setAnnotated = (v: string | null | undefined) => {
		if (v === "any") {
			filtersState.annotation = undefined;
		} else {
			filtersState.annotation = v === "has";
		}
	}

	let kindMenuOpen = $state(false);
	const toggleKindMenu = () => (kindMenuOpen = !kindMenuOpen);
	const eventKindOptions: MenuOption<EventKind>[] = [
		{value: "alert", label: "Alerts"}
	]
</script>

<div class="flex flex-col gap-2">
	<Field label="Date Range" labelPlacement="top" dense base classes={{root: "", container: "px-0 border-none py-0", input: "my-0 gap-2"}}>
		<ToggleGroup variant="outline" inset classes={{root: "bg-surface-100 w-full"}} bind:value={filtersState.dateRangeOption}>
			{#if !!filtersState.activeShift}
				<ToggleOption value="shift">Active Shift</ToggleOption>
			{/if}

			{#each dateRangeOptions as opt}
				<ToggleOption value={opt.value}>{opt.label}</ToggleOption>
			{/each}
		</ToggleGroup>
	</Field>

	{#if filtersState.dateRangeOption === "custom"}
		<DateRangeField
			label="Custom Date Range"
			periodTypes={[PeriodType.Day]}
			getPeriodTypePresets={() => []}
			dense
			classes={{
				field: { root: "gap-0", container: "pl-0 py-[2px] flex items-center", prepend: "[&>span]:mr-2" },
			}}
			icon={mdiCalendarRange}
			bind:value={() => filtersState.dateRange, d => (filtersState.customDateRangeValue = d)}
		/>
	{/if}

				
	<SelectField 
		label="Annotated"
		labelPlacement="top"
		resize
		dense
		classes={{ root: "w-28", field: { root: "gap-0", container: "h-8 flex items-center", input: "my-0" } }}
		options={annoOptions}
		clearable={false}
		bind:value={() => annoValue, setAnnotated}
	/>

	<RosterSelectField 
		selectedId={filtersState.rosterId} 
		onSelected={id => (filtersState.rosterId = id)}
		dense
		classes={{ root: "gap-0 w-44", field: {root: "gap-0", container: "h-8"} }}
	/>

	<Field
		label="Event Kind"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0 w-fit", container: "px-0 h-8 flex items-center", input: "my-0" }}
		let:id
	>
		<Button {id} onclick={toggleKindMenu} classes={{ root: "h-8 px-2" }}>
			<div class="flex gap-2">
				{#each (filtersState.eventKinds ?? []) as v}
					<span class="flex items-center gap-1">
						{v}
					</span>
				{:else}
					<span class="text-sm font-normal">Any</span>
				{/each}
			</div>
			<Icon data={mdiChevronDown} />
			<MultiSelectMenu
				options={eventKindOptions}
				bind:value={filtersState.eventKinds}
				open={kindMenuOpen}
				maintainOrder
				placeholder="Event Kinds"
				on:change={(e) => (filtersState.eventKinds = (e.detail.value as string[]))}
				on:close={toggleKindMenu}
			/>
		</Button>
	</Field>
</div>
