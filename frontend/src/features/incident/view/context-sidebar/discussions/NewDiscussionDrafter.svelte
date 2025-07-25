<script lang="ts">
	import { onMount } from "svelte";
	import { createMutation } from "@tanstack/svelte-query";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { createReplyEditor, draft } from "$features/incident/lib/discussions.svelte";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { createRetrospectiveDiscussionMutation, type RetrospectiveDiscussion } from "$lib/api";
	import Header from "$components/header/Header.svelte";

	type Props = {
		retrospectiveId: string;
		onDiscussionCreated: (discussion: RetrospectiveDiscussion) => void;
	};
	const { retrospectiveId, onDiscussionCreated }: Props = $props();

	let draftEditor = $state<SvelteEditor>();
	let contentSize = $state(0);

	const createDiscussion = createMutation(() => ({
		...createRetrospectiveDiscussionMutation(),
		onSuccess({ data }) {
			onDiscussionCreated(data);
		},
		onError(error, variables, context) {
			console.error(error);
		},
	}));

	const saveDraft = async () => {
		if (!draft.open || !draftEditor) return;

		const content = draftEditor.getJSON();
		createDiscussion.mutate({
			path: { id: retrospectiveId },
			body: { attributes: { content } },
		});
	};

	const cancelDraft = () => {
		draft.clear(true);
	};

	onMount(() => {
		draftEditor = createReplyEditor(null, true);
		draftEditor.on("update", ({ editor }) => {
			contentSize = editor.$doc.content.size;
		});
		return () => {
			console.log("clearing");
			// draft.clear(false);
			if (draftEditor) draftEditor.destroy();
		};
	});
</script>

<div class="border border-accent rounded-lg p-2 flex flex-col gap-2">
	<Header title="New Discussion" subheading="drafting" />

	<div class="border border-neutral-200 bg-surface-300 cursor-text p-1">
		{#if draftEditor}
			<TiptapEditor bind:editor={draftEditor} />
		{/if}
	</div>

	<ConfirmChangeButtons
		alignRight
		confirmText="Save"
		saveEnabled={contentSize > 1}
		loading={createDiscussion.isPending}
		onClose={cancelDraft}
		onConfirm={saveDraft}
	/>

	{#if createDiscussion.isError}
		<span>error: {createDiscussion.error}</span>
	{/if}
</div>
