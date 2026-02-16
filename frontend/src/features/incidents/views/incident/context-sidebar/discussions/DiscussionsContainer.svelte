<script lang="ts">
	import { createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { listRetrospectiveCommentsOptions, type RetrospectiveComment } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { draft } from "$src/features/incidents/lib/discussions.svelte";
	import DiscussionThread from "./DiscussionThread.svelte";
	import NewDiscussionDrafter from "./NewDiscussionDrafter.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		retrospectiveId: string;
	};
	let { retrospectiveId }: Props = $props();

	const queryClient = useQueryClient();

	const queryOptions = $derived(listRetrospectiveCommentsOptions({ path: { id: retrospectiveId } }));
	const query = createQuery(() => queryOptions);

	const onDiscussionCreated = (d: RetrospectiveComment) => {
		if (draft.editor) {
			draft.editor.commands.convertDraftToAnnotation(d.id);
			draft.clear(true);
		}
		const { queryKey } = queryOptions;
		queryClient.setQueryData(queryKey, (data) => {
			if (!data) return { data: [d], pagination: { total: 1 } };
			const newData = structuredClone(data);
			newData.data.push(d);
			return newData;
		});
		queryClient.invalidateQueries({ queryKey });
	};
</script>

<div class="col-span-3 flex flex-col gap-2 overflow-y-auto border p-2">
	<Header title="Discuss" />

	<div class="flex flex-row gap-2">
		<span class="rounded-lg border px-3 py-1 bg-primary cursor-pointer">All</span>
		<span class="rounded-lg border px-3 py-1">Comments</span>
		<span class="rounded-lg border px-3 py-1">Action Items</span>
	</div>

	{#if draft.open}
		<NewDiscussionDrafter {retrospectiveId} {onDiscussionCreated} />
	{/if}

	<div class="overflow-y-auto flex flex-col gap-2">
		<LoadingQueryWrapper {query}>
			{#snippet view(discussions: RetrospectiveComment[])}
				{#each discussions as discussion (discussion.id)}
					<DiscussionThread {discussion} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
