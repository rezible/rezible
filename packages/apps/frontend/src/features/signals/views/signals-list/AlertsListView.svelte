<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initSignalsListViewController } from "./controller.svelte";
	import SignalRow from "./SignalRow.svelte";

	const controller = initSignalsListViewController();
	registerPageDescriptor(() => ({ title: "Signals" }));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
	<div class="w-full max-w-sm">
		<SearchInput bind:value={controller.search} label="Search signals" />
	</div>
	<LoadingQueryWrapper query={controller.query} feedbackOnly />
	<PaginatedQueryListBox {...controller.paginatedAlertsQuery}>
		{#if controller.query.data}
			<ul class="divide-y rounded-lg border bg-card">
				{#each controller.query.data.data as definition (definition.id)}
					<li>
						<SignalRow {definition} />
					</li>
				{:else}
					<li class="p-6 text-sm text-muted-foreground">No matching signals.</li>
				{/each}
			</ul>
		{/if}
	</PaginatedQueryListBox>
</div>
