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
		children: Snippet;
	};
	const { breadcrumbs, children }: Props = $props();
</script>

<div class="w-full max-w-full h-full max-h-full min-h-0 overflow-hidden flex flex-col gap-2 p-2 border rounded-lg bg-surface-300">
	<div class="border-b">
		<PageBreadcrumbs crumbs={breadcrumbs} />
	</div>

	<div class="flex flex-col flex-1 min-h-0 overflow-auto">
		{@render children()}
	</div>
</div>
