<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listOncallRostersOptions, type ListOncallRostersData } from "$lib/api";
	import RosterCard from "./RosterCard.svelte";

	let params = $state<ListOncallRostersData["query"]>({});
	const query = createQuery(() => listOncallRostersOptions({ query: { ...params, pinned: true } }));

	const updateSearch = (value: any) => {
		console.log(value);
	};
</script>

<div class="flex flex-col gap-2 overflow-x-hidden">
	<!--TextField
        label="Search"
        dense
        on:change={(e) => updateSearch(e.detail)}
        debounceChange
        iconRight={mdiMagnify}
        labelPlacement="float"
    /-->

	<div class="w-full border-b"></div>

	{#if query.isLoading}
		<span>loading pinned rosters</span>
	{:else if query.isError}
		<span>error getting pinned rosters: {query.error}</span>
	{:else if query.isSuccess}
		{@const rosters = query.data.data}
		{#each rosters as roster}
			<RosterCard {roster} />
		{/each}
		{#if rosters.length === 0}
			<span>No Pinned Rosters</span>
		{/if}
	{/if}
</div>
