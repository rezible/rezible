<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTasksOptions, type ListTasksData, type Task, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import BacklogList from "./BacklogList.svelte";

	type Props = {
		teamId: string;
	}
	let { teamId }: Props = $props();

	let params = $state<ListTasksData>();
	const query = createQuery(() => listTasksOptions({ ...params, query: { teamId } }));
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(tasks: Task[])}
		<BacklogList {tasks} />
	{/snippet}
</LoadingQueryWrapper>
