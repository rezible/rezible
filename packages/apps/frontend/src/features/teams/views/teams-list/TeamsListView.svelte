<script lang="ts">
	import type { Team } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import SearchInput from "$src/components/forms/search-input/SearchInput.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { resolve } from "$app/paths";
	import { initTeamsListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Teams" }));

	const controller = initTeamsListController();
</script>

{#snippet filters()}
	<SearchInput bind:value={controller.search} />
{/snippet}

{#snippet teamCard(team: Team)}
	<a href={resolve(`/teams/${team.attributes.slug}`)}>
		<span>team card</span>
		<!--ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
			<svelte:fragment slot="avatar">
				<Avatar kind="team" size={32} id={team.id} />
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
			</div>
		</ListItem-->
	</a>
{/snippet}

<FilterPage {filters}>
	<PaginatedQueryListBox {...controller.paginatedTeamsQuery}>
		<LoadingQueryWrapper query={controller.query}>
			{#snippet view(teams: Team[])}
				{#each teams as team (team.id)}
					{@render teamCard(team)}
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Teams Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
