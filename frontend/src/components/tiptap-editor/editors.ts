import type { Content, EditorOptions } from "@tiptap/core";
import { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
import { configureBaseExtensions, configureUserMentionExtension, getHandoverExtensions, getDiscussionExtensions, getPlaybookExtensions } from "@rezible/tiptap-extensions";
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

export const createHandoverEditor = (options: Partial<EditorOptions>) => {
	return new SvelteEditor({
		extensions: getHandoverExtensions(RezUserSuggestion),
		editorProps: {
			attributes: {
				// class: "focus:outline-none " + classes,
			},
		},
		...options,
	});
};

export const createDiscussionEditor = (options: Partial<EditorOptions>) => {
	return new SvelteEditor({
		extensions: getDiscussionExtensions(RezUserSuggestion),
		editorProps: {
			attributes: {
				// class: "focus:outline-none " + classes,
			},
		},
		...options,
	});
};

export const createPlaybookEditor = (options: Partial<EditorOptions>) => {
	return new SvelteEditor({
		extensions: getPlaybookExtensions(RezUserSuggestion),
		editorProps: {
			attributes: {
				// class: "focus:outline-none " + classes,
			},
		},
		...options,
	});
};
