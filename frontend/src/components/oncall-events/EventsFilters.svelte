<script lang="ts" module>
	import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";

	type EventKind = OncallEventAttributes["kind"];

	export type FilterOptions = {
		rosterId?: string;
		eventKinds?: EventKind[];
		annotated?: boolean;
		dateRange?: DateRangeType;
	};

	export type DisabledFilters = {
		roster?: boolean;
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
		MenuItem,
		MultiSelectMenu,
		MultiSelectOption,
		SelectField,
		type MenuOption,
	} from "svelte-ux";
	import { PeriodType } from "@layerstack/utils";
	import { subDays } from "date-fns";
	import { listOncallRostersOptions, type OncallEventAttributes, type OncallRoster } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { debounce } from "$lib/utils.svelte";
	import { watch } from "runed";
	import { cls } from "@layerstack/tailwind";

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
			seenIds.add(r.value);
			options.push(r);
		});
		userRosters.forEach(r => {
			if (seenIds.has(r.id)) return;
			options.push({value: r.id, label: r.attributes.name});
		});
		return options;
	});
	
	let rosterMenuOpen = $state(false);
	let selectedRosterOption = $state<MenuOption<string>>();
	watch(() => filters.rosterId, id => {
		selectedRosterOption = id ? $state.snapshot(rosterOptions.find(o => (o.value === id))) : undefined;
	});

	const onRosterSelected = (value?: string | null) => {
		filters.rosterId = !!value ? value : undefined;
	}

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
			classes={{ root: "w-28", field: { root: "gap-0", container: "h-8 flex items-center", input: "my-0" } }}
			options={annoOptions}
			clearable={false}
			value={annoValue}
			on:change={e => setAnnotated(e.detail.value)}
		/>
	{/if}

	{#if !disabled?.roster}
		<SelectField 
			label="Roster"
			labelPlacement="top"
			value={filters.rosterId}
			bind:open={rosterMenuOpen}
			on:change={e => onRosterSelected(e.detail.value)}
			search={async (s, o) => {setRostersSearch(s); return o}}
			maintainOrder
			dense
			classes={{ root: "gap-0 w-44", field: {root: "gap-0", container: "h-8"} }}
			options={rosterOptions}
		>
			<div slot="prepend" class:hidden={rosterMenuOpen} class="mr-2">
				{#if !!selectedRosterOption}
					<Avatar kind="roster" id={selectedRosterOption.value} size={18} />
				{:else}
					<span>Any</span>
				{/if}
			</div>

			<svelte:fragment
				slot="option"
				let:option
				let:index
				let:selected
				let:highlightIndex
			>
				<MenuItem
					class={cls(
						index === highlightIndex && "bg-surface-content/5",
						option === selected && "font-semibold",
						option.group ? "px-4" : "px-2",
					)}
					scrollIntoView={index === highlightIndex}
					disabled={option.disabled}
				>
					<span class="flex items-center gap-2">
						<Avatar kind="roster" id={option.value} size={18} />
						{option.label}
					</span>
				</MenuItem>
			</svelte:fragment>
		</SelectField>
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
