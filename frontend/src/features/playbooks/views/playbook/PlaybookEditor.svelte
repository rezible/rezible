<script lang="ts">
	import { onMount } from "svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { configureBaseExtensions } from "@rezible/documents/tiptap-extensions";
	import { playbookViewStateCtx } from "./viewState.svelte";

	const viewState = playbookViewStateCtx.get();

	const mountEditor = () => {
		viewState.editor = new SvelteEditor({
			content: viewState.playbookContent,
			extensions: configureBaseExtensions(false),
			editable: true,
			autofocus: false,
			editorProps: {
				attributes: {
					class: "max-w-none focus:outline-none min-h-20",
				},
			},
			onFocus({ editor }) {

			},
			onBlur() {

			},
		});
		return () => {
			if (!viewState.editor?.isDestroyed) viewState.editor?.destroy();
		};
	};
	onMount(mountEditor);

	const onEditorContainerFocused = () => {
		if (!viewState.editor || viewState.editor.isFocused) return;
		viewState.editor.chain().focus("end").run();
	};
</script>

<div
	class="flex-1 border border-surface-content/15 bg-surface-content/5 p-2 px-3 rounded-lg mt-1 cursor-text w-full"
	tabindex="-1"
	spellcheck="false"
	onfocus={onEditorContainerFocused}
>
	{#if viewState.editor}
		<TiptapEditor bind:editor={viewState.editor} />
	{/if}
</div>
