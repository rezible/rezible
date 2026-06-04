<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";

	type Props = {
		children: Snippet;
		dense?: boolean;
		pagination?: any;//ComponentProps<Pagination>["pagination"];
	};
	const { 
		children,
		dense,
		pagination,
	}: Props = $props();
</script>

<div class="flex flex-col max-h-full" class:max-w-xl={dense}>
	<div class="flex flex-col gap-1 min-h-0 flex-0 overflow-auto pb-1 pr-1">
		{@render children()}
	</div>

	<div class="pt-1 flex justify-end">
		<!-- <Pagination {pagination} show={["perPage", "prevPage", "pagination", "nextPage"]} /> -->
	</div>
</div>