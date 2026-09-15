<script lang="ts">
	import { onMount } from "svelte";
	import type { Editor } from "@tiptap/core";
	import type { HocuspocusProvider } from "@hocuspocus/provider";
	import {
		configureBaseExtensions,
		configureUserMentionExtension,
		configureAnnotationExtension,
		configureDraftDiscussionHighlightExtension,
	} from "@rezible/tiptap-extensions";
	import Collaboration from "@tiptap/extension-collaboration";
	import CollaborationCaret from "@tiptap/extension-collaboration-caret";

	import type { RetrospectiveReportSection } from "$lib/api";
	import { useUserSessionState } from "$lib/user-session.svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { RezUserSuggestion } from "$components/tiptap-editor/user-suggestions/user-suggestion.svelte";
	import { activeAnnotation, activeEditor } from "../activeEditor.svelte";
	import { useIncidentCollaboration } from "../../collaboration.svelte";
	import MenuBar from "./MenuBar.svelte";
	import { watch } from "runed";

	type Props = {
		section: RetrospectiveReportSection;
		focusEditor: () => void;
	};
	let {section, focusEditor = $bindable()}: Props = $props();

	const session = useUserSessionState();
	const collab = useIncidentCollaboration();

	const isEditable = $derived(true);

	// TODO: load this
	const userAccentColor = "#a33333";

	const configureEditorExtensions = (field: string, provider: HocuspocusProvider) => {
		const user = { name: session.user?.attributes.name, color: userAccentColor };
		return [
			...configureBaseExtensions(false),
			configureUserMentionExtension(RezUserSuggestion),
			configureAnnotationExtension(activeAnnotation.set),
			configureDraftDiscussionHighlightExtension(session.user?.id),
			Collaboration.configure({ document: provider.document, field }),
			CollaborationCaret.configure({ provider, user }),
		];
	};

	let editor = $state<SvelteEditor>();
	const createEditor = (provider?: HocuspocusProvider) => {
		if (!provider) return;
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
				activeEditor.set(editor, section.field);
			},
			onBlur() {
				// setIsActive(undefined)
			},
		});
	};
	watch(() => collab.provider, createEditor);
	watch(() => isEditable, (editable) => {
		editor?.setEditable(editable);
	});
	onMount(() => {
		return () => {
			if (!editor) return;
			if (activeEditor.editor === editor) activeEditor.clear();
			if (!editor.isDestroyed) editor.destroy();
		};
	});

	const onEditorContainerFocused = () => {
		if (!editor || editor.isFocused) return;
		editor.chain().focus("end").run();
	};
	focusEditor = onEditorContainerFocused;
</script>

<div class="flex h-8">
	<div class="flex-1 flex h-8 items-end">
		<span class="text-lg text-foreground/80">{section.title}</span>
	</div>
	<div class="">
		{#if isEditable && activeEditor.field === section.field}
			<MenuBar />
		{/if}
	</div>
</div>

<div
	class="border-t border-border py-4 first:border-t-0"
	tabindex="-1"
	spellcheck="false"
	onfocus={onEditorContainerFocused}
>
	{#if editor}
		<TiptapEditor bind:editor />
	{/if}
</div>
