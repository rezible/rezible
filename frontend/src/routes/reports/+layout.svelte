<script lang="ts">
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import PageContainer from "$components/page-container/PageContainer.svelte";
	import type { Snippet } from "svelte";
	import SavedReportsNav from "./SavedReportsNav.svelte";
	
	type Props = {
		children: Snippet;
	}
	const { children }: Props = $props();
</script>

<PageContainer breadcrumbs={[{ label: "Reports" }]}>
	<SplitPage>
		{#snippet nav()}<SavedReportsNav />{/snippet}
		{@render children()}
	</SplitPage>
</PageContainer>