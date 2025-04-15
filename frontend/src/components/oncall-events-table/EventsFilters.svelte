<script lang="ts" module>
	import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";

	type EventKind = OncallEventAttributes["kind"];

	export type FilterOptions = {
		rosterIds?: string[];
		eventKinds?: EventKind[];
		annotated?: boolean;
		dateRange?: DateRangeType;
	};

	export type DisabledFilters = {
		rosters?: boolean;
		kinds?: boolean;
		annotated?: boolean;
		dateRange?: boolean;
	};
</script>

<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
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
	import { listOncallRostersOptions, type OncallEventAttributes, type OncallRoster } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { debounce } from "$lib/utils.svelte";
	import { watch } from "runed";

	type Props = {
		filters: FilterOptions;
		disabled?: DisabledFilters;
		userRosters?: OncallRoster[];
	};
	let { filters = $bindable(), disabled, userRosters = [] }: Props = $props();

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

	let rostersSearch = $state<string>();
	const setRostersSearch = debounce((s?: string) => (rostersSearch = s), 500);

	const placeholderRostersData = $derived({data: userRosters, pagination: {total: userRosters.length}});
	const rostersQuery = createQuery(() => ({
		placeholderData: placeholderRostersData,
		...listOncallRostersOptions({query: {search: (!!rostersSearch ? rostersSearch : undefined)}}),
	}));
	const rosters = $derived(rostersQuery.data?.data ?? []);

	const queryRosterOptions = $derived(rosters.map(r => ({value: r.id, label: r.attributes.name})));

	const rosterOptions = $derived.by(() => {
		if (!!rostersSearch) return queryRosterOptions;

		const options: MenuOption<string>[] = [];
		const seenIds = new Set<string>();
		queryRosterOptions.forEach(r => {
			options.push(r);
			seenIds.add(r.value);
		});
		userRosters.forEach(r => {
			if (!seenIds.has(r.id)) options.push({value: r.id, label: r.attributes.name});
		});
		return options;
	});
	
	let selectedRosterOptions = $state<MenuOption<string>[]>([]);
	watch(() => filters.rosterIds, (ids) => {
		selectedRosterOptions = ids ? $state.snapshot(rosterOptions.filter(o => (ids.includes(o.value)))) : [];
	});

	let rosterMenuOpen = $state(false);
	const toggleRosterMenu = () => (rosterMenuOpen = !rosterMenuOpen);

	const today = new Date();
	const defaultDateRange: DateRangeType = {from: subDays(today, 7), to: today, periodType: PeriodType.Day};

	let kindMenuOpen = $state(false);
	const toggleKindMenu = () => (kindMenuOpen = !kindMenuOpen);
	const eventKindOptions: MenuOption<EventKind>[] = [
		{value: "alert", label: "Alerts"}
	]
</script>

<div class="flex flex-row items-center gap-2 justify-end">
	{#if !disabled?.dateRange}
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
	{/if}

	{#if !disabled?.annotated}
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
	{/if}

	{#if !disabled?.rosters}
		<Field
			label="Roster"
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
					search={async (s, o) => {setRostersSearch(s); return o}}
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
	{/if}

	{#if !disabled?.kinds}
		<Field
			label="Event Kind"
			labelPlacement="top"
			dense
			classes={{ root: "gap-0", container: "px-0 h-8 flex items-center", input: "my-0" }}
			let:id
		>
			<Button {id} on:click={toggleKindMenu} classes={{ root: "h-8" }}>
				<div class="flex gap-2">
					{#each (filters.eventKinds ?? []) as v}
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
					bind:value={filters.eventKinds}
					open={kindMenuOpen}
					maintainOrder
					placeholder="Event Kinds"
					on:change={(e) => (filters.eventKinds = (e.detail.value as string[]))}
					on:close={toggleKindMenu}
				/>
			</Button>
		</Field>
	{/if}
</div>
