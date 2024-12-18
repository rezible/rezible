<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { Button, Header } from 'svelte-ux';
	import { listRetrospectiveDiscussionsOptions, type RetrospectiveDiscussion } from '$lib/api';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import { draft } from '$features/incidents/views/retrospective/lib/discussions.svelte';
	import DiscussionThread from './DiscussionThread.svelte';
	import NewDiscussionDrafter from './NewDiscussionDrafter.svelte';
    import DebriefBox from './DebriefBox.svelte';

	type Props = {
		debriefId: string;
		retrospectiveId: string;
		showDebriefDialog: boolean;
	}
	let { debriefId, retrospectiveId, showDebriefDialog = $bindable() }: Props = $props();

	const queryClient = useQueryClient();

	const queryOptions = $derived(listRetrospectiveDiscussionsOptions({path: {id: retrospectiveId}}));
	const query = createQuery(() => queryOptions);

	const onDiscussionCreated = (d: RetrospectiveDiscussion) => {
		if (draft.editor) {
			draft.editor.commands.convertDraftToAnnotation(d.id);
			draft.clear(true);
		}
		const { queryKey } = queryOptions;
		queryClient.setQueryData(queryKey, data => {
			if (!data) return {data: [d], pagination: {total: 1}};
			const newData = structuredClone(data);
			newData.data.push(d);
			return newData;
		});
		queryClient.invalidateQueries({queryKey});
	}
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<Header title="Discussions">
		<svelte:fragment slot="actions">
			<Button>New</Button>
		</svelte:fragment>
	</Header>

	<DebriefBox {debriefId} bind:showDebriefDialog />

	{#if draft.open}
		<NewDiscussionDrafter {retrospectiveId} {onDiscussionCreated} />
	{/if}

	<div class="overflow-y-auto flex flex-col gap-2">
		<LoadingQueryWrapper {query}>
			{#snippet view(discussions: RetrospectiveDiscussion[])}
				{#each discussions as discussion (discussion.id)}
					<DiscussionThread {discussion} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
