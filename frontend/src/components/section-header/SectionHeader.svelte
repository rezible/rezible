<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import { mdiFilter } from "@mdi/js";
	import type { Snippet } from "svelte";
	import { Button } from "svelte-ux";

	type Props = {
		title: string;
		subheading?: string;
		filters?: Snippet;
		actions?: Snippet;
	};
	let {
		title,
		subheading,
		filters,
		actions: propActions,
	}: Props = $props();

	let showFilters = $state(false);
</script>

{#snippet filterActions()}
	<Button icon={mdiFilter} iconOnly on:click={() => {showFilters = !showFilters}} />
{/snippet}

<div class="flex flex-col">
	<Header {title} {subheading} classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			{#if filters}{@render filterActions()}{/if}
			{#if propActions}{@render propActions()}{/if}
		{/snippet}
	</Header>

	{#if showFilters && !!filters}
		{@render filters()}
	{/if}
</div>