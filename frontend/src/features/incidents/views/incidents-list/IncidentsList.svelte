<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiChevronRight, mdiMagnify, mdiFilter } from "@mdi/js";
	import { Button, ListItem, TextField } from "svelte-ux";
	import { listIncidentsOptions, type Incident } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";

	const query = createQuery(() => listIncidentsOptions({}));
</script>

<div class="flex flex-col min-h-0 w-full h-full gap-1 max-h-full">
	<div class="flex flex-row gap-2 pb-1">
		<Button icon={mdiFilter} iconOnly />

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
	</div>

	{#snippet incidentListRow(inc: Incident)}
		<a href="/incidents/{inc.attributes.slug}">
			<ListItem title={inc.attributes.title} classes={{ root: "hover:bg-secondary-900" }}>
				<svelte:fragment slot="avatar">
					<Avatar id={inc.id} square kind="incident" />
				</svelte:fragment>
				<div slot="actions">
					<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
				</div>
			</ListItem>
		</a>
	{/snippet}

	<LoadingQueryWrapper {query}>
		{#snippet view(incidents: Incident[])}
			<div class="min-h-0 flex flex-col gap-2 overflow-y-auto flex-0">
				{#each incidents as inc (inc.id)}
					{@render incidentListRow(inc)}
				{/each}
			</div>
		{/snippet}
	</LoadingQueryWrapper>
</div>