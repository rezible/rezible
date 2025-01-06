import { isActive, type ChainedCommands, type Editor, type Extensions, type Content } from '@tiptap/core';
import type { EditorState, Transaction } from '@tiptap/pm/state';
import { Editor as SvelteEditor } from 'svelte-tiptap';

import { debounce } from '$lib/utils.svelte';
import { session } from '$lib/auth.svelte';

import {
	configureBaseExtensions,
	configureUserMentionExtension,
	configureAnnotationExtension,
	configureDraftDiscussionHighlightExtension,
} from '@rezible/documents/tiptap-extensions';

import type { HocuspocusProvider } from '@hocuspocus/provider';
import Collaboration from '@tiptap/extension-collaboration';
import CollaborationCursor from '@tiptap/extension-collaboration-cursor';

import { RezUserSuggestion } from './user-suggestions/user-suggestion.svelte';

export const createMentionEditor = (content: Content, classes = "") => {
	const userMentions = configureUserMentionExtension(RezUserSuggestion);
	const baseExtensions = configureBaseExtensions(true);
	return new SvelteEditor({
		content,
		extensions: [...baseExtensions, userMentions],
		editorProps: {
			attributes: {
				class: 'focus:outline-none ' + classes
			}
		}
	})
};

export type ActiveStatus = {
	focused?: boolean;
	bold?: boolean;
	italic?: boolean;
	code?: boolean;
	codeBlock?: boolean;
	bulletList?: boolean;
	orderedList?: boolean;
	taskList?: boolean;
	blockquote?: boolean;
	heading1?: boolean;
	heading2?: boolean;
	paragraph?: boolean;
}
const createActiveStatus = () => {
	let status = $state<ActiveStatus>({});

	return {
		set: (s: ActiveStatus) => {status = s},
		get focused() { return !!status.focused },
		get bold() { return !!status.bold },
		get italic() { return !!status.italic },
		get code() { return !!status.code },
		get codeBlock() { return !!status.codeBlock },
		get bulletList() { return !!status.bulletList },
		get orderedList() { return !!status.orderedList },
		get taskList() { return !!status.taskList },
		get blockquote() { return !!status.blockquote },
		get heading1() { return !!status.heading1 },
		get heading2() { return !!status.heading2 },
		get paragraph() { return !!status.paragraph },
	};
};
export const activeStatus = createActiveStatus();

const updateActiveStatus = (state: EditorState, focused: boolean) => {
	const selection = state.selection;
	if (selection.ranges.length > 0) {
		const range = selection.ranges[0];
		const rangeSize = range.$to.pos - range.$from.pos;
		if (rangeSize > 1000) {
			activeStatus.set({});
			return;
		}
	}

	if (!focused) {
		activeStatus.set({focused: false});
		return;
	}

	// TODO: this can probably be quicker
	let newActiveMarks: ActiveStatus = {
		focused: true,
		bold: isActive(state, 'bold'),
		italic: isActive(state, 'italic'),
		code: isActive(state, 'code'),
		codeBlock: isActive(state, 'codeBlock'),
		bulletList: isActive(state, 'bulletList'),
		orderedList: isActive(state, 'orderedList'),
		taskList: isActive(state, 'taskList'),
		blockquote: isActive(state, 'blockquote')
	};

	if (isActive(state, 'heading', { level: 1 })) {
		newActiveMarks['heading1'] = true;
	} else if (isActive(state, 'heading', { level: 2 })) {
		newActiveMarks['heading2'] = true;
	} else {
		newActiveMarks['paragraph'] = true;
	}

	activeStatus.set(newActiveMarks);
};

const onEditorTransaction = debounce(
	({ editor, transaction }: { editor: Editor; transaction: Transaction }) => {
		updateActiveStatus(editor.state, editor.isFocused);
	},
);

type RunCommandFn = (cmd: ChainedCommands) => void;
const createActiveEditorState = () => {
	let editor = $state<Editor>();
	let field = $state<string>();

	const clear = () => {
		if (editor) editor.off('transaction', onEditorTransaction);
		field = undefined;
	};

	const set = (e: Editor, f: string) => {
		clear();
		field = f;
		editor = e;
		editor.on('transaction', onEditorTransaction);
	};

	const tryRunCommand = (fn: RunCommandFn) => {
		return () => {
			if (!editor) return;
			const chain = editor.chain().focus();
			fn(chain);
			chain.run();
		};
	};

	return {
		set,
		clear,
		tryRunCommand,
		get field() { return field },
		get editor() { return editor },
	};
};
export let activeEditor = createActiveEditorState();

const createActiveAnnotationIdState = () => {
	let id = $state<string>();

	return {
		get id() { return id },
		set: (newId?: string) => {id = newId},
	};
};

export const activeAnnotation = createActiveAnnotationIdState();

export const configureEditorExtensions = (field: string, provider: HocuspocusProvider) => {
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