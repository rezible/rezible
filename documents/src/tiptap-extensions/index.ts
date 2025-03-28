import type { Extensions } from '@tiptap/core';

import StarterKit from '@tiptap/starter-kit';
import Link from '@tiptap/extension-link';
import Image from '@tiptap/extension-image';
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'

import { HighlightDraftDiscussionExtension } from './highlight-draft-discussions';
import { RezUserMentionExtension, type SuggestionExtensionType } from './user-mention';
import { AnnotationExtension } from './annotation';
import { BulletList } from '@tiptap/extension-bullet-list';

export const configureUserMentionExtension = (suggestion?: SuggestionExtensionType) => {
	return RezUserMentionExtension.configure({suggestion});
}

export const configureDraftDiscussionHighlightExtension = (sessionUserId = "") => {
	return HighlightDraftDiscussionExtension.configure({sessionUserId});
}

export const configureAnnotationExtension = (setActiveAnnotation?: (id?: string) => void) => {
	return AnnotationExtension.configure({setActiveAnnotation});
}

const configureTaskListExtensions = (nested = false) => {
	return [
		TaskList,
		TaskItem.configure({
			nested: false,
		})
	]
}

export const configureBaseExtensions = (history: boolean): Extensions => {
	return [
		StarterKit.configure({history: history ? undefined : false}),
		BulletList.configure({HTMLAttributes: {"class": "list-disc ml-4"}}),
		Link,
		Image,
	]
}

export const getHandoverExtensions = () => configureBaseExtensions(true);