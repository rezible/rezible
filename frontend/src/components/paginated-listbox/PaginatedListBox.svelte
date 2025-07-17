<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";
	import { Pagination } from "svelte-ux";

	type Props = {
		title?: string;
		children: Snippet;	
		pagination: ComponentProps<Pagination>["pagination"];
	};
	const { 
		title = "Results",
		children,
		pagination,
	}: Props = $props();
</script>

<div class="flex flex-col max-h-full max-w-xl">
	<span class="text-sm uppercase font-semibold text-surface-content/60">{title}</span>

	<div class="flex flex-col gap-2 min-h-0 flex-0 overflow-auto pb-1 pr-1 border-b">
		{@render children()}
	</div>

	<div class="pt-1 flex justify-end">
		<Pagination {pagination} show={["perPage", "prevPage", "pagination", "nextPage"]} />
	</div>
</div>