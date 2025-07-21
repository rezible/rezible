import type { Content } from "@tiptap/core";
import { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
import {
	configureBaseExtensions,
	configureUserMentionExtension,
} from "@rezible/documents/tiptap-extensions";
import { RezUserSuggestion } from "$components/tiptap-editor/user-suggestions/user-suggestion.svelte";

export const createMentionEditor = (content: Content, classes = "") => {
	const baseExtensions = configureBaseExtensions(true);
	const userMentions = configureUserMentionExtension(RezUserSuggestion);
	return new SvelteEditor({
		content,
		extensions: [...baseExtensions, userMentions],
		editorProps: {
			attributes: {
				class: "focus:outline-none " + classes,
			},
		},
	});
};