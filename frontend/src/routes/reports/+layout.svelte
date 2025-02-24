<script lang="ts">
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import type { Snippet } from "svelte";
	import SavedReportsNav from "./SavedReportsNav.svelte";
	
	type Props = {
		children: Snippet;
	}
	const { children }: Props = $props();
</script>

<SplitPage>
	{#snippet nav()}<SavedReportsNav />{/snippet}
	{@render children()}
</SplitPage>