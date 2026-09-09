<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initSituationsListController } from "./controller.svelte";
	import Filters from "$features/situations/components/situation-filters/SituationFilters.svelte";
	import Row from "$features/situations/components/situation-row/SituationRow.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const controller = initSituationsListController();
	registerPageDescriptor(() => ({ title: "Situations" }));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
	<div class="flex flex-wrap items-end justify-between gap-2">
		<Filters filters={controller.filters} onchange={controller.setFilters} />
		<Button variant="ghost" size="sm" onclick={controller.resetFilters}>Reset filters</Button>
	</div>
	<LoadingQueryWrapper query={controller.query} feedbackOnly />
	<PaginatedQueryListBox {...controller.paginatedSituationsQuery}>
		{#if controller.query.data}
			{#each controller.query.data.data as item (item.id)}
				<Row situation={item} />
			{:else}
				<p class="p-6 text-sm text-muted-foreground">No matching situations.</p>
			{/each}
		{/if}
	</PaginatedQueryListBox>
</div>
