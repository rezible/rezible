<script lang="ts">
	import type { Task } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initTeamBacklogController } from "./controller.svelte";

	const controller = initTeamBacklogController();
</script>

{#snippet tasksView(tasks: Task[])}
	{#each tasks as task (task.id)}
		{@const attr = task.attributes}
		<div>
			<span>task: {attr.name}</span>
			<!-- <ListItem title={attr.name} classes={{ root: "hover:bg-surface-200", title: "text-lg" }}>
					<div slot="subheading">
						<span class="text-surface-content/80">{attr.description}</span>
					</div>
					<div slot="avatar">
						<span>-</span>
					</div>
					<div slot="actions">
						<Button icon={RiArrowRightSLine} class="p-2 text-surface-content/50" />
					</div>
				</ListItem> -->
		</div>
	{/each}
{/snippet}

<div class="flex flex-col w-full">
	<PaginatedQueryListBox {...controller.paginatedTasksQuery}>
		<LoadingQueryWrapper query={controller.query} view={tasksView} />
	</PaginatedQueryListBox>
</div>
