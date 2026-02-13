<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import { listUsersOptions, type ListUsersData } from "$lib/api";
	import { useTeamViewState } from "$features/team";
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

	const view = useTeamViewState();
	const paginator = new QueryPaginatorState();
	
	const params = $derived<ListUsersData["query"]>({
		teamId: view.teamId,
		...paginator.queryParams,
	})
	const usersQuery = createQuery(() => ({
		...listUsersOptions({ query: params }),
		enabled: !!view.teamId,
	}));
	paginator.watchQuery(usersQuery);
	const users = $derived(usersQuery.data?.data ?? []);
</script>

<div class="flex flex-col max-h-full max-w-xl border p-2">
	<span class="uppercase font-semibold text-surface-content/90">Users</span>

	<div class="flex flex-col gap-2 min-h-0 flex-0 overflow-auto py-1">
		{#each users ?? [] as user (user.id)}
			<a class="flex gap-2 items-center rounded border border-surface-content/10 p-2" href="/users/{user.id}">
				<Avatar kind="user" size={20} id={user.id} />
				<span>{user.attributes.name}</span>
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
