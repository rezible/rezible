<script lang="ts">
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "System Components" }]);

	const pagination = createPaginationStore();
	let searchValue = $state<string>();
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox {pagination}>
		<span></span>
	</PaginatedListBox>
</FilterPage>
