<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "System Components" }]);

	let searchValue = $state<string>();
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<div class="flex flex-col gap-2 min-h-0 flex-1 overflow-auto">
		components
	</div>
</FilterPage>
