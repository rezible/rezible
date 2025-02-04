<script lang="ts">
	import type { Task } from "$lib/api";
	import { mdiChevronRight } from "@mdi/js";
	import { Button, ListItem } from "svelte-ux";

	interface Props {
		tasks: Task[];
	}
	let { tasks }: Props = $props();
</script>

<div class="flex flex-col w-full">
	{#each tasks as task}
		{@const attr = task.attributes}
		<a href="/tasks/{task.id}">
			<ListItem
				title={attr.name}
				classes={{ root: "hover:bg-surface-200", title: "text-lg" }}
			>
				<div slot="subheading">
					<span class="text-surface-content/80"
						>{attr.description}</span
					>
				</div>
				<div slot="avatar">
					<span>-</span>
				</div>
				<div slot="actions">
					<Button
						icon={mdiChevronRight}
						class="p-2 text-surface-content/50"
					/>
				</div>
			</ListItem>
		</a>
	{/each}
</div>
