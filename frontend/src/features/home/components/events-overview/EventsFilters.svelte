<script lang="ts" module>
	import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";

	export type FilterOptions = {
		rosterIds?: string[];
		annotated?: boolean;
		dateRange?: DateRangeType;
	};
</script>

<script lang="ts">
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import { mdiCalendarRange, mdiChevronDown } from "@mdi/js";
	import {
		Button,
		DateRangeField,
		Field,
		Icon,
		MultiSelectMenu,
		MultiSelectOption,
		SelectField,
		type MenuOption,
	} from "svelte-ux";
	import { v4 as uuidv4 } from "uuid";
	import { PeriodType } from "@layerstack/utils";
	import { subDays } from "date-fns";

	type Props = {
		filters: FilterOptions;
	};
	let { filters = $bindable() }: Props = $props();

	type AnnotationOption = "no" | "any" | "has";
	const annoOptions: MenuOption<AnnotationOption>[] = [
		{value: "any", label: "Any"},
		{value: "has", label: "Yes"},
		{value: "no", label: "No"},
	];
	const annoValue = $derived(filters.annotated === undefined ? "any" : (filters.annotated ? "yes" : "no"));
	const setAnnotated = (v: string | null | undefined) => {
		if (v === "any") {
			filters.annotated = undefined;
		} else {
			filters.annotated = v === "has";
		}
	}

	const rosterOptions: MenuOption<string>[] = [
		{ label: "One", value: uuidv4() },
	];
	const selectedRostersSet = $derived(new Set(filters.rosterIds));
	const selectedRosterOptions = $derived(rosterOptions.filter((o) => selectedRostersSet.has(o.value)));
	let rosterMenuOpen = $state(false);
	const toggleRosterMenu = () => (rosterMenuOpen = !rosterMenuOpen);

	let serviceIds = $state<string[]>();
	const serviceOptions: MenuOption<string>[] = [
		{ label: "Foo", value: uuidv4() },
	];
	const selectedServicesSet = $derived(new Set());
	const selectedServiceOptions = $derived(serviceOptions.filter((o) => selectedServicesSet.has(o.value)));
	let serviceMenuOpen = $state(false);
	const toggleServiceMenu = () => (serviceMenuOpen = !serviceMenuOpen);

	const today = new Date();
	const defaultDateRange: DateRangeType = {from: subDays(today, 7), to: today, periodType: PeriodType.Day};
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
		value={filters.dateRange || defaultDateRange}
		on:change={(e) => {filters.dateRange = (e.detail as DateRangeType)}}
	/>

	<SelectField 
		label="Annotation"
		labelPlacement="top"
		resize
		dense
		classes={{ root: "gap-0 w-32", field: { root: "gap-0", container: "h-8 flex items-center", input: "my-0" } }}
		options={annoOptions}
		clearable={false}
		value={annoValue}
		on:change={e => setAnnotated(e.detail.value)}
	/>

	<!-- <Field
		label="Action"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0", container: "h-8 flex items-center" }}
		let:id
	>
		<Checkbox {id} bind:checked={annotated} classes={{ label: "pl-2" }}>Annotated</Checkbox>
	</Field> -->

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
				value={filters.rosterIds}
				open={rosterMenuOpen}
				search
				maintainOrder
				placeholder="Filter to roster"
				on:change={(e) => (filters.rosterIds = (e.detail.value as string[]))}
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
		label="Services"
		labelPlacement="top"
		dense
		classes={{ root: "gap-0", container: "px-0 h-8 flex items-center", input: "my-0" }}
		let:id
	>
		<Button {id} on:click={toggleServiceMenu} classes={{ root: "h-8" }}>
			<div class="flex gap-2">
				{#each selectedServiceOptions as v, i (v.value)}
					<span class="flex items-center gap-1">
						<Avatar kind="service" id={v.value} size={14} />
						{v.label + (i < selectedServiceOptions.length - 1 ? "," : "")}
					</span>
				{:else}
					<span>Any</span>
				{/each}
			</div>
			<Icon data={mdiChevronDown} />
			<MultiSelectMenu
				options={serviceOptions}
				value={serviceIds}
				open={serviceMenuOpen}
				search
				maintainOrder
				placeholder="Filter to service"
				on:change={(e) => (serviceIds = (e.detail.value as string[]))}
				on:close={toggleServiceMenu}
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
					<Avatar kind="service" id={(option as MenuOption<string>).value} size={22} />
					{label}
				</MultiSelectOption>
			</MultiSelectMenu>
		</Button>
	</Field>
</div>
