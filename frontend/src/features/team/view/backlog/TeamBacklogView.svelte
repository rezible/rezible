<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTasksOptions, type ListTasksData, type Task } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import { ListItem } from "svelte-ux";
	import Button from "$components/button/Button.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useTeamViewState } from "$features/team";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

	const view = useTeamViewState();

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
				<ListItem title={attr.name} classes={{ root: "hover:bg-surface-200", title: "text-lg" }}>
					<div slot="subheading">
						<span class="text-surface-content/80">{attr.description}</span>
					</div>
					<div slot="avatar">
						<span>-</span>
					</div>
					<div slot="actions">
						<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
					</div>
				</ListItem>
			</a>
		{/each}
{/snippet}

<div class="flex flex-col w-full">
	<LoadingQueryWrapper {query} view={tasksView} />
</div>