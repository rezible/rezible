<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTasksOptions, type ListTasksData, type Task } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import { Button, ListItem } from "svelte-ux";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useTeamViewState } from "../viewState.svelte";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";

	const viewState = useTeamViewState();

	const paginator = new QueryPaginatorState();
	const queryParams = $derived<ListTasksData["query"]>({
		teamId: viewState.teamId,
		...paginator.queryParams,
	})
	const query = createQuery(() => listTasksOptions({ query: queryParams }));
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