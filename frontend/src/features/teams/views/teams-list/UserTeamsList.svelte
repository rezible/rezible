<script lang="ts">
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import { Button, ListItem } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";

	let params = $state<ListTeamsData>();
	const userTeamsQuery = createQuery(() => listTeamsOptions(params));
</script>

<div class="flex flex-col h-full">
	<LoadingQueryWrapper query={userTeamsQuery}>
		{#snippet view(userTeams: Team[])}
			{#each userTeams as team}
				<a href="/teams/{team.attributes.slug}">
					<ListItem title={team.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
						<svelte:fragment slot="avatar">
							<Avatar kind="team" size={32} id={team.id} />
						</svelte:fragment>
						<div slot="actions">
							<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
						</div>
					</ListItem>
				</a>
			{/each}
		{/snippet}
	</LoadingQueryWrapper>
</div>
