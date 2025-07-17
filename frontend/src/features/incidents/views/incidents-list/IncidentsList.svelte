<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiChevronRight, mdiMagnify } from "@mdi/js";
	import { ListItem, TextField } from "svelte-ux";
	import { listIncidentsOptions, type ListIncidentsData, type Incident } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";

	let searchValue = $state<string>();

	let allQueryParams = $state<ListIncidentsData["query"]>({});
	const allIncidentsQuery = createQuery(() => listIncidentsOptions({query: allQueryParams}));
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<div class="flex flex-col min-h-0 w-full h-full gap-1 max-h-full">
		<LoadingQueryWrapper query={allIncidentsQuery}>
			{#snippet view(incidents: Incident[])}
				<div class="min-h-0 flex flex-col gap-2 overflow-y-auto flex-0">
					{#each incidents as inc (inc.id)}
						<a href="/incidents/{inc.attributes.slug}">
							<ListItem title={inc.attributes.title} classes={{ root: "hover:bg-secondary-900" }}>
								<svelte:fragment slot="avatar">
									<Avatar id={inc.id} square kind="incident" />
								</svelte:fragment>
								<div slot="actions">
									<Icon data={mdiChevronRight} size={24} classes={{root: "text-surface-content/50"}} />
								</div>
							</ListItem>
						</a>
					{/each}
				</div>
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</FilterPage>