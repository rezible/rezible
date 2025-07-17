<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Alerts" }]);

	let searchValue = $state<string>();
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<div class="min-h-0 flex flex-col gap-2 overflow-y-auto">
		list
	</div>
</FilterPage>
