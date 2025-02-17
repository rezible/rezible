<script lang="ts">
	import { } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Header } from "svelte-ux";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";
	
</script>

<SplitPage nav={savedReportsNav}>
	<Header title="(title)" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />
</SplitPage>

{#snippet savedReportsNav()}
	<Header title="Saved" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />
{/snippet}