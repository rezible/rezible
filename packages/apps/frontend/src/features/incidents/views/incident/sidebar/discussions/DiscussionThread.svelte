<script lang="ts">
	import {
		activeDiscussion,
		createDiscussionCommentsController,
		discussionCommentContent,
	} from "$features/incidents/views/incident/discussions.svelte";
	import type { DiscussionThread as DiscussionThreadType } from "$lib/api";
	import TiptapEditor, { Editor as SvelteEditor } from "$src/components/tiptap-editor/TiptapEditor.svelte";
	import { createDiscussionEditor } from "$src/components/tiptap-editor/editors";

	type Props = { discussion: DiscussionThreadType };
	let { discussion }: Props = $props();
	let editor = $state<SvelteEditor>();

	const commentsController = createDiscussionCommentsController(() => discussion.id);

	$effect(() => {
		const message = commentsController.query.data?.data?.[0];
		if (!message) return;
		const content = discussionCommentContent(message.attributes.content);
		const instance = createDiscussionEditor({ content, editable: false });
		editor = instance;
		return () => {
			instance.destroy();
		};
	});

	const setActiveDiscussion = () => activeDiscussion.set(discussion.id);
</script>

<button
	type="button"
	class="border p-2 rounded-lg flex flex-col gap-2 text-left"
	class:border-primary={activeDiscussion.id === discussion.id}
	onclick={setActiveDiscussion}
>
	<div class="flex flex-col">
		<span class="font-medium">{discussion.attributes.userId}</span>
		<span class="text-sm text-muted-foreground"
			>{new Date(discussion.attributes.createdAt).toLocaleString()}</span
		>
	</div>

	{#if commentsController.query.isPending}
		<span class="text-sm text-muted-foreground">Loading discussion…</span>
	{:else if commentsController.query.isError}
		<span class="text-sm text-destructive">Unable to load discussion.</span>
	{:else if editor}
		{#key editor}
			<TiptapEditor bind:editor />
		{/key}
	{:else}
		<span class="text-sm text-muted-foreground">No messages yet.</span>
	{/if}
</button>
