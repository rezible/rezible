<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTasksOptions, type ListTasksData, type Task } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useTeamViewController } from "$features/teams/views/team";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

	const view = useTeamViewController();

	const paginator = new QueryPaginatorState();
	const queryParams = $derived<ListTasksData["query"]>({
		teamId: view.teamId,
		...paginator.queryParams,
	});
	const query = createQuery(() => ({
		...listTasksOptions({ query: queryParams }),
		enabled: !!view.teamId,
	}));
	paginator.watchQuery(query);
</script>

{#snippet tasksView(tasks: Task[])}
		{#each tasks as task}
			{@const attr = task.attributes}
			<a href="/tasks/{task.id}">
				<span>task: {attr.name}</span>
				<!-- <ListItem title={attr.name} classes={{ root: "hover:bg-surface-200", title: "text-lg" }}>
					<div slot="subheading">
						<span class="text-surface-content/80">{attr.description}</span>
					</div>
					<div slot="avatar">
						<span>-</span>
					</div>
					<div slot="actions">
						<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
					</div>
				</ListItem> -->
			</a>
		{/each}
{/snippet}

<div class="flex flex-col w-full">
	<LoadingQueryWrapper {query} view={tasksView} />
</div>