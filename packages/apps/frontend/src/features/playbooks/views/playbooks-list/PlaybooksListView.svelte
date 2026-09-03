<script lang="ts">
	import type { Playbook } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { resolve } from "$app/paths";
	import { initPlaybooksListController } from "./controller.svelte";

	setPageBreadcrumbs(() => [{ label: "Playbooks" }]);

	const controller = initPlaybooksListController();
</script>

{#snippet filters()}
	<SearchInput bind:value={controller.search} />
{/snippet}

{#snippet playbookListItem(pb: Playbook)}
	<a href={resolve(`/playbooks/${pb.id}`)}>
		<span>{pb.attributes.title}</span>
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedQueryListBox {...controller.paginatedPlaybooksQuery}>
		<LoadingQueryWrapper query={controller.query}>
			{#snippet view(playbooks: Playbook[])}
				{#each playbooks as pb (pb.id)}
					{@render playbookListItem(pb)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Playbooks Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
