<script lang="ts">
	import { mdiCalendarRange, mdiChevronDown } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import RosterSelectField from "$components/roster-select-field/RosterSelectField.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import type { EventKind, EventsTableState } from "./eventsTableState.svelte";

	type Props = {
		tableState: EventsTableState;
	};
	const { tableState }: Props = $props();

	type AnnotationOption = "no" | "any" | "has";
	const annoOptions = [
		{value: "any", label: "Any"},
		{value: "has", label: "Yes"},
		{value: "no", label: "No"},
	];
	const annoValue = $derived(tableState.filters.annotated === undefined ? "any" : (tableState.filters.annotated ? "yes" : "no"));
	const setAnnotated = (v: string | null | undefined) => {
		if (v === "any") {
			tableState.filters.annotated = undefined;
		} else {
			tableState.filters.annotated = v === "has";
		}
	}

	let kindMenuOpen = $state(false);
	const toggleKindMenu = () => (kindMenuOpen = !kindMenuOpen);
	const eventKindOptions = [
		{value: "alert", label: "Alerts"}
	]
</script>

<div class="flex flex-row items-center justify-end gap-2">
	<span>filters</span>
	<!--
	<SelectField 
		label="Annotation"
		labelPlacement="top"
		resize
		dense
		classes={{ root: "w-28", field: { root: "gap-0", container: "h-8 flex items-center", input: "my-0" } }}
		options={annoOptions}
		clearable={false}
		bind:value={() => annoValue, setAnnotated}
	/>

	<RosterSelectField 
		selectedId={tableState.filters.rosterId} 
		onSelected={id => (tableState.filters.rosterId = id)}
		dense
		classes={{ root: "gap-0 w-44", field: {root: "gap-0", container: "h-8"} }}
	/>

	<Field
		label="Event Kind"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0", container: "px-0 h-8 flex items-center", input: "my-0" }}
		let:id
	>
		<Button {id} onclick={toggleKindMenu} classes={{ root: "h-8" }}>
			<div class="flex gap-2">
				{#each (tableState.filters.eventKinds ?? []) as v}
					<span class="flex items-center gap-1">
						{v}
					</span>
				{:else}
					<span>Any</span>
				{/each}
			</div>
			<Icon data={mdiChevronDown} />
			<MultiSelectMenu
				options={eventKindOptions}
				bind:value={tableState.filters.eventKinds}
				open={kindMenuOpen}
				maintainOrder
				placeholder="Event Kinds"
				on:change={(e) => (tableState.filters.eventKinds = (e.detail.value as string[]))}
				on:close={toggleKindMenu}
			/>
		</Button>
	</Field>

	{#if tableState.dateRangeOption === "custom"}
		<DateRangeField
			label="Custom Date Range"
			periodTypes={[PeriodType.Day]}
			getPeriodTypePresets={() => []}
			dense
			classes={{
				field: { root: "gap-0", container: "pl-0 py-[2px] flex items-center", prepend: "[&>span]:mr-2" },
			}}
			icon={mdiCalendarRange}
			bind:value={() => tableState.dateRange, d => (tableState.customDateRangeValue = d)}
		/>
	{/if}
	-->
</div>
