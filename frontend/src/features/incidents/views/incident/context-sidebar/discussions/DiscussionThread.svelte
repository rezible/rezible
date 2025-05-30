<script lang="ts">
	import { Button, Tooltip } from "svelte-ux";
	import { onMount } from "svelte";
	import type { RetrospectiveDiscussion } from "$lib/api";
	import type { JSONContent } from "@tiptap/core";
	import { mdiCheck } from "@mdi/js";
	import { activeDiscussion, createReplyEditor } from "$features/incidents/lib/discussions.svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import Header from "$src/components/header/Header.svelte";

	type Props = {
		discussion: RetrospectiveDiscussion;
	}
	let { discussion }: Props = $props();

	const setActiveDiscussion = () => activeDiscussion.set(discussion.id);

	let editor = $state<SvelteEditor>();
	onMount(() => {
		const content = JSON.parse(discussion.attributes.content) as JSONContent;
		editor = createReplyEditor(content, false);
		return () => {
			if (editor) editor.destroy();
		};
	});
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="border p-2 cursor-pointer rounded-lg flex flex-col gap-2"
	class:border-primary={activeDiscussion.id === discussion.id}
	onclick={setActiveDiscussion}
>
	<Header title="User Name" subheading="date/time">
		{#snippet actions()}
			<Tooltip title="Mark Completed">
				<Button iconOnly size="sm" icon={mdiCheck} />
			</Tooltip>
		{/snippet}
	</Header>

	<div class="border-b-2"></div>

	<div class="">
		{#if editor}
			<TiptapEditor bind:editor />
		{/if}
	</div>

	<Button on:click={(e) => e.stopPropagation()}>Add Reply</Button>
</div>
