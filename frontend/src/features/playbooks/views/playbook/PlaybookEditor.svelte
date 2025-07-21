<script lang="ts">
	import { onMount } from "svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import type { Editor } from "@tiptap/core";
	import { configureBaseExtensions } from "@rezible/documents/tiptap-extensions";

	let editor = $state<SvelteEditor>();
	const mountEditor = () => {
		editor = new SvelteEditor({
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
				// setIsActive(undefined)
			},
		});
		return () => {
			if (!editor?.isDestroyed) editor?.destroy();
		};
	};
	onMount(mountEditor);

	const onEditorContainerFocused = () => {
		if (!editor || editor.isFocused) return;
		editor.chain().focus("end").run();
	};
</script>

<div
	class="flex-1 border border-surface-content/15 bg-surface-content/5 p-2 px-3 rounded-lg mt-1 cursor-text w-full"
	tabindex="-1"
	spellcheck="false"
	onfocus={onEditorContainerFocused}
>
	{#if editor}
		<TiptapEditor bind:editor />
	{/if}
</div>
