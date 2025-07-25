<script lang="ts">
	import { onMount } from "svelte";
	import { type RetrospectiveReportSection } from "$lib/api";
	import { session } from "$lib/auth.svelte";

	import { activeAnnotation, activeEditor } from "$features/incident/lib/activeEditor.svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { RezUserSuggestion } from "$components/tiptap-editor/user-suggestions/user-suggestion.svelte";
	import type { Editor, Extensions } from "@tiptap/core";
	import type { HocuspocusProvider } from "@hocuspocus/provider";
	import {
		configureBaseExtensions,
		configureUserMentionExtension,
		configureAnnotationExtension,
		configureDraftDiscussionHighlightExtension,
	} from "@rezible/documents/tiptap-extensions";
	import Collaboration from "@tiptap/extension-collaboration";
	import CollaborationCursor from "@tiptap/extension-collaboration-cursor";
	import BubbleMenu, { type AnnotationType } from "./BubbleMenu.svelte";
	import MenuBar from "./MenuBar.svelte";

	type Props = {
		section: RetrospectiveReportSection;
		provider: HocuspocusProvider;
		setIsActive: (e: Editor, field: string) => void;
		onCreateAnnotation: (e: Editor, t: AnnotationType) => void;
		focusEditor: () => void;
	};
	let { section, provider, setIsActive, onCreateAnnotation, focusEditor = $bindable() }: Props = $props();

	const configureEditorExtensions = (field: string, provider: HocuspocusProvider) => {
		const user = { name: session.username, color: session.accentColor };
		const extensions: Extensions = [
			...configureBaseExtensions(false),
			configureUserMentionExtension(RezUserSuggestion),
			configureAnnotationExtension(activeAnnotation.set),
			configureDraftDiscussionHighlightExtension(session.user?.id),
			Collaboration.configure({ document: provider.document, field }),
			CollaborationCursor.configure({ provider, user }),
		];

		return extensions;
	};


	let editor = $state<SvelteEditor>();
	const mountEditor = () => {
		editor = new SvelteEditor({
			extensions: configureEditorExtensions(section.field, provider),
			editable: true,
			autofocus: false,
			editorProps: {
				attributes: {
					class: "max-w-none focus:outline-none min-h-20",
				},
			},
			onFocus({ editor }) {
				setIsActive(editor, section.field);
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
	focusEditor = onEditorContainerFocused;
</script>

<div class="flex h-8">
	<div class="flex-1 flex h-8 items-end">
		<span class="text-lg text-surface-content/80">{section.title}</span>
	</div>
	<div class="">
		{#if activeEditor.field === section.field}
			<MenuBar />
		{/if}
	</div>
</div>

<div
	class="border border-surface-content/15 bg-surface-content/5 p-2 px-3 rounded-lg mt-1"
	tabindex="-1"
	spellcheck="false"
	onfocus={onEditorContainerFocused}
>
	{#if editor}
		<BubbleMenu
			{editor}
			field={section.field}
			onCreate={(t) => onCreateAnnotation(editor as Editor, t)}
		/>
		<TiptapEditor bind:editor />
	{/if}
</div>
