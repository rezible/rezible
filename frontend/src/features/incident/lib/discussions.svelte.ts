import type { Editor } from "@tiptap/core";

const createActiveDiscussion = () => {
	let value = $state<string>();

	return {
		get id() {
			return value;
		},
		set: (id?: string) => {
			value = id;
		},
	};
};
export const activeDiscussion = createActiveDiscussion();

type Draft = {
	editor: Editor;
};
const createDraft = () => {
	let value = $state<Draft>();

	const set = (val?: Draft) => {
		value = val;
	};

	const clear = (navigate: boolean) => {
		if (value) {
			if (navigate) value.editor.commands.navigateToDraftDiscussion();
			value.editor.commands.clearDraftDiscussion();
		}
		set();
	};

	const create = (editor: Editor) => {
		clear(false);
		set({ editor });
		editor.commands.draftDiscussion();
	};

	return {
		get open() {
			return value !== undefined;
		},
		get editor() {
			return value?.editor;
		},
		set,
		create,
		clear,
	};
};
export const draft = createDraft();
