<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listOncallRostersOptions, type ListOncallRostersData } from "$lib/api";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { useTeamViewController } from "$features/teams/views/team";
	import Avatar from "$components/avatar/Avatar.svelte";

	const view = useTeamViewController();
	const paginator = new QueryPaginatorState();
	
	const params = $derived<ListOncallRostersData["query"]>({
		teamId: view.teamId,
		...paginator.queryParams,
	});
	const rostersQuery = createQuery(() => ({
		...listOncallRostersOptions({ query: params }),
		enabled: !!view.teamId,
	}));
	paginator.watchQuery(rostersQuery);
	
	const rosters = $derived(rostersQuery.data?.data ?? []);
</script>

<div class="flex flex-col max-h-full max-w-xl border p-2">
	<span class="text-sm uppercase font-semibold text-surface-content/90">Rosters</span>

	<div class="flex flex-col gap-2 min-h-0 flex-0 overflow-auto py-1">
		{#each rosters ?? [] as roster (roster.id)}
			<a class="flex gap-2 items-center rounded border border-surface-content/10 p-2" href="/rosters/{roster.id}">
				<Avatar kind="roster" size={20} id={roster.id} />
				<span>{roster.attributes.name}</span>
			</a>
		{/each}
	</div>

	<div class="flex justify-end">
		<!-- <Pagination 
			{...paginator.paginationProps}
			hideSinglePage
		/> -->
	</div>
</div>
