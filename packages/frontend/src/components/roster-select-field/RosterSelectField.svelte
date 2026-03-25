<script lang="ts">
	import { debounce } from "$lib/utils.svelte";
	import { type ListOncallRostersData, listOncallRostersOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { useUserOncallInformation } from "$lib/userOncall.svelte";

	type Props = {
		onSelected: (id: string | undefined) => void;
		selectedId?: string;
		dense?: boolean;
		clearable?: boolean;
	};
	const {
		onSelected,
		selectedId,
		dense,
		clearable,
	}: Props = $props();

	const userInfo = useUserOncallInformation();

	const userRosters = $derived(userInfo.rosters);

	let searchValue = $state<string>();
	const setSearchValue = debounce((s?: string) => (searchValue = s), 500);

	let menuOpen = $state(false);
	const queryEnabled = $derived(menuOpen);
	let queryParams = $derived<ListOncallRostersData["query"]>({search: searchValue});
	const query = createQuery(() => ({
		...listOncallRostersOptions({query: queryParams}),
		enabled: queryEnabled,
	}));
	const rosters = $derived(query.data?.data || userRosters);

	const queryOptions = $derived(rosters.map(r => ({value: r.id, label: r.attributes.name})));

	/*
	const rosterOptions = $derived.by(() => {
		if (!!searchValue) return queryOptions;

		const options: MenuOption<string>[] = [];
		const seenIds = new Set<string>();
		queryOptions.forEach(r => {
			seenIds.add(r.value);
			options.push(r);
		});
		userRosters.forEach(r => {
			if (seenIds.has(r.id)) return;
			options.push({value: r.id, label: r.attributes.name});
		});
		return options;
	});

	const onRosterSelected = (value?: string | null) => {
		onSelected(!!value ? value : undefined);
	}

	const searchFn = async (input: string, opts: string[]) => {
		setSearchValue(input);
		return opts
	}
	*/
</script>

<!--SelectField 
	label="Roster"
	labelPlacement="top"
	loading={query.isLoading}
	bind:open={menuOpen}
	bind:value={() => selectedId, onRosterSelected}
	search={searchFn}
	maintainOrder
	{dense}
	{classes}
	{clearable}
	options={rosterOptions}
>
	<div slot="prepend" class:hidden={menuOpen} class="mr-2">
		{#if !!selectedId}
			<Avatar kind="roster" id={selectedId} size={18} />
		{:else}
			<span class="text-sm">Any</span>
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
</SelectField-->