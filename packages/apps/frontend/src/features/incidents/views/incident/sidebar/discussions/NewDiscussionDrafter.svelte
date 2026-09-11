<script lang="ts">
	import { onMount } from "svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$src/components/tiptap-editor/TiptapEditor.svelte";
	import { draft, type DiscussionsController } from "$features/incidents/views/incident/discussions.svelte";
	import ConfirmChangeButtons from "$components/forms/confirm-buttons/ConfirmButtons.svelte";
	import Header from "$src/components/layout/header/Header.svelte";
	import { createDiscussionEditor } from "$src/components/tiptap-editor/editors";

	type Props = {
		controller: DiscussionsController;
	};
	const { controller }: Props = $props();

	let draftEditor = $state<SvelteEditor>();
	let contentSize = $state(0);

	const saveDraft = async () => {
		if (!draft.open || !draftEditor) return;

		controller.saveDraft(draftEditor);
	};

	const cancelDraft = () => {
		draft.clear(true);
	};

	onMount(() => {
		draftEditor = createDiscussionEditor({ editable: true });
		draftEditor.on("update", ({ editor }) => {
			contentSize = editor.$doc.content.size;
		});
		return () => {
			if (draftEditor) draftEditor.destroy();
		};
	});
</script>

<div class="border border-primary rounded-lg p-2 flex flex-col gap-2">
	<Header title="New Discussion" subheading="drafting" />

	<div class="border border-neutral-200 bg-background cursor-text p-1">
		{#if draftEditor}
			<TiptapEditor bind:editor={draftEditor} />
		{/if}
	</div>

	<ConfirmChangeButtons
		alignRight
		confirmText="Save"
		saveEnabled={contentSize > 1}
		loading={controller.createDiscussion.isPending}
		onClose={cancelDraft}
		onConfirm={saveDraft}
	/>

	{#if controller.createDiscussion.isError}
		<span>error: {controller.createDiscussion.error}</span>
	{/if}
</div>
