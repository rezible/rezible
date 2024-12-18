<script lang="ts">
	import { onMount } from 'svelte';
	import { EditorContent, Editor as SvelteEditor } from 'svelte-tiptap';
	import { type Extensions, type Editor, type EditorOptions } from '@tiptap/core';
    import { type RetrospectiveSection } from '$lib/api';
    import MenuBar from './MenuBar.svelte';
    import { activeEditor, configureEditorExtensions } from './editor.svelte';
	import BubbleMenu, { type AnnotationType } from './BubbleMenu.svelte';
    import type { HocuspocusProvider } from '@hocuspocus/provider';

	type Props = { 
		section: RetrospectiveSection;
		provider: HocuspocusProvider;
		setIsActive: (e: Editor, field: string) => void;
		onCreateAnnotation: (e: Editor, t: AnnotationType) => void;
		focusEditor: () => void;
	};
	let { section, provider, setIsActive, onCreateAnnotation, focusEditor = $bindable() }: Props = $props();

	const editor = new SvelteEditor({
		extensions: configureEditorExtensions(section.field, provider),
		editable: true,
		autofocus: false,
		editorProps: {
			attributes: {
				class: 'max-w-none focus:outline-none min-h-20'
			}
		},
		onFocus({ editor }) {
			setIsActive(editor, section.field);
		},
		onBlur() {
			// setIsActive(undefined)
		}
	});
	onMount(() => {
		return () => {
			if (!editor?.isDestroyed) editor?.destroy()
		};
	});

	const onEditorContainerFocused = () => {
		if (!editor || editor.isFocused) return;
		editor.chain().focus('end').run();
	};
	focusEditor = onEditorContainerFocused;
</script>

<div class="flex h-10">
	<div class="flex-1 flex h-10 items-end">
		<span class="text-lg text-surface-content/80">{section.title}</span>
	</div>
	<div class="">
		{#if activeEditor.field === section.field}
			<MenuBar />
		{/if}
	</div>
</div>

<div class="border border-surface-content/15 bg-surface-content/5 p-2 px-3 rounded-lg mt-1" tabindex="-1" spellcheck="false" onfocus={onEditorContainerFocused}>
	{#if editor}
		<BubbleMenu {editor} field={section.field} onCreate={t => onCreateAnnotation(editor as Editor, t)} />
		<EditorContent {editor} />
	{/if}
</div>