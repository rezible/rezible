import type { Extensions } from '@tiptap/core';

import StarterKit from '@tiptap/starter-kit';
import Image from '@tiptap/extension-image';
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import Bold from "@tiptap/extension-bold";
import Italic from "@tiptap/extension-italic";
import { TableKit } from "@tiptap/extension-table";
import { Markdown } from "@tiptap/markdown";

import { HighlightDraftDiscussionExtension } from './highlight-draft-discussions';
import { RezUserMentionExtension, type SuggestionExtensionType } from './user-mention';
import { AnnotationExtension } from './annotation';
import { OffsetHeading } from './offset-heading';

export { OffsetHeading };

export const configureUserMentionExtension = (suggestion?: SuggestionExtensionType) => {
	return RezUserMentionExtension.configure({suggestion});
}

export const configureDraftDiscussionHighlightExtension = (sessionUserId = "") => {
	return HighlightDraftDiscussionExtension.configure({sessionUserId});
}

export const configureAnnotationExtension = (setActiveAnnotation?: (id?: string) => void) => {
	return AnnotationExtension.configure({setActiveAnnotation});
}

export const configureBaseExtensions = (undoRedo?: boolean): Extensions => {
	if (undoRedo) undoRedo = undefined;
	const kit = StarterKit.configure({
		undoRedo,
		bulletList: {HTMLAttributes: {"class": "list-disc ml-4"}},
	});
	return [kit, Image];
}

export const getUserMentionExtension = (suggestion?: SuggestionExtensionType) => RezUserMentionExtension.configure({suggestion});

export const getHandoverExtensions = (suggestion?: SuggestionExtensionType) => [
	...configureBaseExtensions(true), 
	getUserMentionExtension(suggestion),
];

export const getDiscussionExtensions = (suggestion?: SuggestionExtensionType) => [
	Document, Paragraph, Text, Bold, Italic, getUserMentionExtension(suggestion),
];

export const getPlaybookExtensions = (suggestion?: SuggestionExtensionType) => [
	...configureBaseExtensions(true), 
	getUserMentionExtension(suggestion),
];

export type ReadOnlyDocumentOptions = {
	/** Added to every Markdown heading level when rendering the HTML tag (clamped to 6). Default 2. */
	headingOffset?: 0 | 1 | 2;
};

const allowedLinkProtocol = /^(https?:|mailto:)/i;

/** Extensions for rendering Markdown documents read-only: no history or editing affordances, safe links, no remote images. */
export const getReadOnlyDocumentExtensions = (options?: ReadOnlyDocumentOptions): Extensions => {
	const offset = options?.headingOffset ?? 2;
	const kit = StarterKit.configure({
		undoRedo: false,
		heading: false,
		dropcursor: false,
		gapcursor: false,
		trailingNode: false,
		link: {
			openOnClick: true,
			autolink: false,
			linkOnPaste: false,
			HTMLAttributes: { rel: "noopener noreferrer nofollow", target: "_blank" },
			isAllowedUri: (url, ctx) => allowedLinkProtocol.test(url) && ctx.defaultValidate(url),
		},
	});
	return [
		kit,
		OffsetHeading.configure({ levels: [1, 2, 3, 4, 5, 6], offset }),
		TableKit.configure({ table: { resizable: false } }),
		Markdown,
	];
};
