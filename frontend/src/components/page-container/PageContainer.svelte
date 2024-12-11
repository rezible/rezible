<script lang="ts" module>
    import type { Breadcrumb } from "./PageBreadcrumbs.svelte";
    import type { PageTabsProps } from "./PageTabs.svelte";

	export type { Breadcrumb, PageTabsProps };
</script>

<script lang="ts">
    import type { Snippet } from "svelte";
    import PageBreadcrumbs from "./PageBreadcrumbs.svelte";
    import PageTabs from "./PageTabs.svelte";

	type Props = {
		breadcrumbs: Breadcrumb[];
		tabs?: PageTabsProps;
		children: Snippet;
	}
	const { breadcrumbs, tabs, children }: Props = $props();
</script>


<div class="w-full max-w-full h-full max-h-full min-h-0 overflow-hidden flex flex-col gap-2">
	<div class="flex flex-col border-b" class:gap-1={!!tabs}>
		<PageBreadcrumbs crumbs={breadcrumbs} />

		{#if tabs}
			<PageTabs {...tabs} />
		{/if}
	</div>

	<div class="flex flex-col flex-1 min-h-0 overflow-auto">
		{@render children()}
	</div>
</div>