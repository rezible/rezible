<script lang="ts">
	import { onMount } from "svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { configureBaseExtensions } from "@rezible/documents/tiptap-extensions";
	import { usePlaybookViewState } from "$features/playbook";

	const view = usePlaybookViewState();

	const mountEditor = () => {
		view.editor = new SvelteEditor({
			content: view.playbookContent,
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
			if (!view.editor?.isDestroyed) view.editor?.destroy();
		};
	};
	onMount(mountEditor);

	const onEditorContainerFocused = () => {
		if (!view.editor || view.editor.isFocused) return;
		view.editor.chain().focus("end").run();
	};
</script>

<div
	class="flex-1 border border-surface-content/15 bg-surface-content/5 p-2 rounded-lg cursor-text w-full"
	tabindex="-1"
	spellcheck="false"
	onfocus={onEditorContainerFocused}
>
	{#if view.editor}
		<TiptapEditor bind:editor={view.editor} />
	{/if}
</div>
