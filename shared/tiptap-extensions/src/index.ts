import type { Extensions } from '@tiptap/core';

import StarterKit from '@tiptap/starter-kit';
import Image from '@tiptap/extension-image';
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import Bold from "@tiptap/extension-bold";
import Italic from "@tiptap/extension-italic";

import { HighlightDraftDiscussionExtension } from './highlight-draft-discussions';
import { RezUserMentionExtension, type SuggestionExtensionType } from './user-mention';
import { AnnotationExtension } from './annotation';

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
