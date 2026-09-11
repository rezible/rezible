<script lang="ts">
	import type { DiscussionThread as DiscussionThreadModel } from "$lib/api";
	import Header from "$src/components/layout/header/Header.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import DiscussionThread from "./DiscussionThread.svelte";

	import {
		createDiscussionsController,
		draft,
		type DiscussionsController,
	} from "$features/incidents/views/incident/discussions.svelte";
	import NewDiscussionDrafter from "./NewDiscussionDrafter.svelte";

	type Props = {
		retrospectiveId: string;
	};
	let { retrospectiveId }: Props = $props();

	const controller: DiscussionsController = createDiscussionsController(() => retrospectiveId);
	const query = controller.query;
</script>

<div class="col-span-3 flex flex-col gap-2 overflow-y-auto border p-2">
	<Header title="Discuss" />

	<div class="flex flex-row gap-2">
		<span class="rounded-lg border px-3 py-1 bg-primary cursor-pointer">All</span>
		<span class="rounded-lg border px-3 py-1">Comments</span>
		<span class="rounded-lg border px-3 py-1">Action Items</span>
	</div>

	{#if draft.open}
		<NewDiscussionDrafter {controller} />
	{/if}

	<div class="overflow-y-auto flex flex-col gap-2">
		<LoadingQueryWrapper {query}>
			{#snippet view(discussions: DiscussionThreadModel[])}
				{#each discussions as discussion (discussion.id)}
					<DiscussionThread {discussion} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
