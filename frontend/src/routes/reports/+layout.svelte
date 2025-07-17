<script lang="ts">
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import type { Snippet } from "svelte";
	import SavedReportsNav from "./SavedReportsNav.svelte";
	
	type Props = {
		children: Snippet;
	}
	const { children }: Props = $props();
</script>

{#snippet filters()}
	<SavedReportsNav />
{/snippet}

<FilterPage {filters}>
	{@render children()}
</FilterPage>