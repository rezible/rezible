import { Mark as MarkExtension, mergeAttributes } from '@tiptap/core';
import { getMarkRanges } from './utils';

declare module '@tiptap/core' {
	interface Commands<ReturnType> {
		draft: {
			draftDiscussion: () => ReturnType;
			clearDraftDiscussion: () => ReturnType;
			navigateToDraftDiscussion: () => ReturnType;
			convertDraftToAnnotation: (id: string) => ReturnType;
		};
	}
}

export interface HighlightDraftDiscussionExtensionOptions {
	sessionUserId: string;
}

export const MARK_NAME = "rezDraftDiscussion";
const MARK_USER_ID_ATTR = "data-user-id";

export const HighlightDraftDiscussionExtension = MarkExtension.create<HighlightDraftDiscussionExtensionOptions>({
	name: MARK_NAME,
	keepOnSplit: false,
	exitable: true,
	inclusive: false,

	onCreate() {
		// cleanup old drafts
		setTimeout(() => {
			if (this.options.sessionUserId) this.editor.commands.clearDraftDiscussion();
		}, 500);
	},

	addOptions() {
		return {
			sessionUserId: "",
		}
	},

	addAttributes() {
		return {
			userId: {
				default: null,
				parseHTML: (el:HTMLElement) => el.getAttribute(MARK_USER_ID_ATTR),
				renderHTML: (attrs) => ({[MARK_USER_ID_ATTR]: attrs.userId})
			}
		};
	},

	parseHTML() {
		const tag = `span[${MARK_USER_ID_ATTR}]`;
		const getAttrs = (el: HTMLElement) => {
			const userId = el.getAttribute(MARK_USER_ID_ATTR)?.trim();
			if (!!userId) return null; // prosemirror expects null === success
			return false;
		};
		return [{ tag, getAttrs }];
	},

	renderHTML({ mark, HTMLAttributes }) {
		const id = String(mark.attrs.userId);
		let _class = "";
		if (id === this.options.sessionUserId) _class = "rez-discussion-draft";
		return ['span', mergeAttributes({class: _class}, HTMLAttributes), 0];
	},

	addCommands() {
		return {
			draftDiscussion: () => ({ commands }) => {
				const userId = this.options.sessionUserId;
				if (!userId) return false;
				return commands.setMark(MARK_NAME, { userId });
			},
			clearDraftDiscussion: () => ({ tr, dispatch }) => {
				const userId = this.options.sessionUserId;
				if (!userId) return false;
				getMarkRanges(tr.doc, (t, attrs) => (t.name === MARK_NAME && attrs.userId === userId))
					.forEach(({ mark, range }) => {
						tr.removeMark(range.from, range.to, mark);
					})
				if (dispatch) return dispatch(tr);
			},
			navigateToDraftDiscussion: () => ({editor, commands, tr, dispatch}) => {
				const userId = this.options.sessionUserId;
				if (!userId) return false;
				let nodePos: number | null = null;

				editor.state.doc.descendants((node, pos) => {
					if (nodePos !== null) return false;
					if (node.marks.length === 0 || node.childCount > 0) return;
					const hasDraftMark = node.marks.some((mark) => {
						return mark.type.name === MARK_NAME && mark.attrs.userId === userId;
					});
					if (!hasDraftMark) return;
					nodePos = pos + node.nodeSize;
					return false;
				});
				if (nodePos === null) {
					return false;
				}
				commands.focus();
				commands.setTextSelection(nodePos);
				if (dispatch) return dispatch(tr);
			},
			convertDraftToAnnotation: (id: string) => ({editor, commands, tr, dispatch}) => {
				const userId = this.options.sessionUserId;
				if (!userId || !editor) return false;
				
				let nodePos: number[] = [];

				editor.state.doc.descendants((node, pos) => {
					if (node.marks.length === 0 || node.childCount > 0) return;
					const hasDraftMark = node.marks.some((mark) => {
						return mark.type.name === MARK_NAME && mark.attrs.userId === userId;
					});
					if (hasDraftMark) {
						nodePos.push(pos);
						return false;
					}
				});

				nodePos.forEach(pos => {
					commands.setNodeSelection(pos)
					commands.createAnnotation(id)
					commands.clearDraftDiscussion()
				})
				if (dispatch) return dispatch(tr);
			}
		};
	}
});