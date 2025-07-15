<script lang="ts" module>
	type EventKind = OncallEventAttributes["kind"];

	export type FilterOptions = {
		rosterId?: string;
		eventKinds?: EventKind[];
		annotated?: boolean;
	};

	export type DisabledFilters = {
		roster?: boolean;
		kinds?: boolean;
		annotated?: boolean;
	};
</script>

<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronDown } from "@mdi/js";
	import {
		Button,
		Field,
		MenuItem,
		MultiSelectMenu,
		SelectField,
		type MenuOption,
	} from "svelte-ux";
	import { listOncallRostersOptions, type OncallEventAttributes, type OncallRoster } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { debounce } from "$lib/utils.svelte";
	import { cls } from "@layerstack/tailwind";
	import type { Snippet } from "svelte";

	type Props = {
		filters: FilterOptions;
		disabled?: DisabledFilters;
		extra?: Snippet;
		userRosters?: OncallRoster[];
	};
	let { filters = $bindable(), disabled, extra, userRosters = [] }: Props = $props();

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

	const rostersQuery = createQuery(() => listOncallRostersOptions({query: {search: (!!rostersSearch ? rostersSearch : undefined)}}));
	const rosters = $derived(rostersQuery.data?.data ?? userRosters);

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

	const onRosterSelected = (value?: string | null) => {
		filters.rosterId = !!value ? value : undefined;
	}

	let kindMenuOpen = $state(false);
	const toggleKindMenu = () => (kindMenuOpen = !kindMenuOpen);
	const eventKindOptions: MenuOption<EventKind>[] = [
		{value: "alert", label: "Alerts"}
	]
</script>

<div class="flex flex-row items-center justify-end gap-2">
	{#if !disabled?.annotated}
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
	{/if}

	{#if !disabled?.roster}
		<SelectField 
			label="Roster"
			labelPlacement="top"
			loading={rostersQuery.isLoading}
			bind:open={rosterMenuOpen}
			bind:value={() => filters.rosterId, onRosterSelected}
			search={async (s, o) => {setRostersSearch(s); return o}}
			maintainOrder
			dense
			classes={{ root: "gap-0 w-44", field: {root: "gap-0", container: "h-8"} }}
			options={rosterOptions}
		>
			<div slot="prepend" class:hidden={rosterMenuOpen} class="mr-2">
				{#if !!filters.rosterId}
					<Avatar kind="roster" id={filters.rosterId} size={18} />
				{:else}
					<span>Any</span>
				{/if}
			</div>

			<svelte:fragment slot="option" let:option let:index let:selected let:highlightIndex>
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

	{@render extra?.()}
</div>
