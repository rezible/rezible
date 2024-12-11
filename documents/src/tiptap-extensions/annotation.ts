import { Mark as MarkExtension, mergeAttributes, type Range } from '@tiptap/core';
import { getMarkRanges } from './utils';

declare module '@tiptap/core' {
	interface Commands<ReturnType> {
		annotations: {
			createAnnotation: (id: string) => ReturnType;
			removeAnnotation: (id: string) => ReturnType;
			navigateToAnnotation: (id: string) => ReturnType;
		};
	}
}

export interface AnnotationExtensionOptions {
	setActiveAnnotation: (id?: string) => void;		
}

export const MARK_NAME = "rezAnnotation";
export const MARK_ANNOTATION_ID_ATTR = "data-annotation-id";

export const AnnotationExtension = MarkExtension.create<AnnotationExtensionOptions>({
	name: MARK_NAME,
	keepOnSplit: false,
	exitable: true,
	inclusive: false,

	onCreate() {
		
	},

	addAttributes() {
		return {
			annotationId: {
				default: null,
				parseHTML: (el: HTMLElement) => el.getAttribute(MARK_ANNOTATION_ID_ATTR),
				renderHTML: (attrs) => ({[MARK_ANNOTATION_ID_ATTR]: attrs.annotationId})
			}
		};
	},

	parseHTML() {
		const tag = `span[${MARK_ANNOTATION_ID_ATTR}]`;
		const getAttrs = (el: HTMLElement) => {
			const annotationId = el.getAttribute(MARK_ANNOTATION_ID_ATTR)?.trim();
			if (!!annotationId) return null; // prosemirror expects null === success
			return false;
		};
		return [{ tag, getAttrs }];
	},

	renderHTML({ mark, HTMLAttributes }) {
		const id = String(mark.attrs.annotationId);
		let _class = "rez-annotation";
		return ['span', mergeAttributes({class: _class}, HTMLAttributes), 0];
	},

	onSelectionUpdate() {
		const marks = this.editor.state.selection.$from.marks();

		const setActiveAnnotation = (annId?: string) => {
			if (this.options.setActiveAnnotation) this.options.setActiveAnnotation(annId);
		}

		if (marks.length === 0) {
			setActiveAnnotation();
			return;
		}

		const annotationMark = this.editor.schema.marks[MARK_NAME];
		const activeAnnotationMark = marks.find((mark) => mark.type === annotationMark);
		const id = activeAnnotationMark?.attrs.annotationId;
		if (id) setActiveAnnotation(id);
	},

	addCommands() {
		return {
			createAnnotation: (id: string) => ({ commands }) => {
				return commands.setMark(MARK_NAME, { annotationId: id });
			},
			removeAnnotation: (id: string) => ({ tr, dispatch }) => {
				if (!id) return false;
				getMarkRanges(tr.doc, (t, attrs) => (t.name === MARK_NAME && attrs.annotationId === id))
					.forEach(({ mark, range }) => 
						tr.removeMark(range.from, range.to, mark));
				if (dispatch) return dispatch(tr);
			},
			navigateToAnnotation: (id: string) => ({editor, commands}) => {
				let nodePos: number | null = null;
				editor.state.doc.descendants((node, pos) => {
					if (nodePos !== null) return false;
					if (node.marks.length === 0 || node.childCount > 0) return;
					const hasAnnotation = node.marks.some((mark) => {
						return mark.type.name === MARK_NAME && mark.attrs.annotationId === id;
					});
					if (!hasAnnotation) return;
					nodePos = pos + node.nodeSize;
					return false;
				});
				if (nodePos === null) {
					return false;
				}
				// commands.focus();
				return commands.setTextSelection(nodePos);
			},
		};
	}
});