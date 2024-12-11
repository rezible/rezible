import type { Range } from '@tiptap/core';
import type { Attrs, Mark, Node, NodeType, MarkType } from '@tiptap/pm/model';

type NodeMarkRange = {mark: Mark; range: Range;}
export type MarkFilterFn = (t: MarkType, attrs: Attrs) => boolean;
export const getMarkRanges = (root: Node, filterFn: MarkFilterFn): NodeMarkRange[] => {
	const findMark = (node: Node) => 
		node.marks.find(({type, attrs}) => 
			filterFn(type, attrs));

	const marks: NodeMarkRange[] = [];
	root.descendants((node: Node, pos: number) => {
		const mark = findMark(node);
		const range = {from: pos, to: pos + node.nodeSize};
		if (mark) marks.push({mark, range});
	});

	return marks;
}