<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiChevronRight, mdiMagnify, mdiFilter } from "@mdi/js";
	import { Button, ListItem, TextField } from "svelte-ux";
	import { listIncidentsOptions, type ListIncidentsData, type Incident } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import SectionHeader from "$src/components/section-header/SectionHeader.svelte";
	import Icon from "$src/components/icon/Icon.svelte";

	let allQueryParams = $state<ListIncidentsData["query"]>({});
	const allIncidentsQuery = createQuery(() => listIncidentsOptions({query: allQueryParams}));

	let userQueryParams = $state<ListIncidentsData["query"]>({});
	const userIncidentsQuery = createQuery(() => listIncidentsOptions({query: userQueryParams}));
</script>

{#snippet incidentList(incidents: Incident[])}
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

<SplitPage>
	{#snippet nav()}
		<SectionHeader title="Your Incidents">
			
		</SectionHeader>

		<LoadingQueryWrapper query={userIncidentsQuery} view={incidentList} />
	{/snippet}

	<SectionHeader title="All Incidents">
		{#snippet filters()}
			<TextField
				dense
				rounded
				classes={{ root: "w-full" }}
				label="Search For Incidents"
				labelPlacement="float"
				icon={mdiMagnify}
				debounceChange={500}
				on:change={({ detail }) =>
					console.log({
						search: detail.value ? String(detail.value) : undefined,
					})}
			/>
		{/snippet}
	</SectionHeader>

	<div class="flex flex-col min-h-0 w-full h-full gap-1 max-h-full">
		<LoadingQueryWrapper query={allIncidentsQuery} view={incidentList} />
	</div>
</SplitPage>