import Heading, { type HeadingOptions } from "@tiptap/extension-heading";
import { mergeAttributes } from "@tiptap/core";

export type OffsetHeadingOptions = HeadingOptions & {
	/** Added to the document heading level when rendering the HTML tag (clamped to 6). */
	offset: number;
};

/**
 * Heading that renders a semantically demoted tag so embedded documents nest under
 * the page outline. `data-level` keeps the document level for styling.
 */
export const OffsetHeading = Heading.extend<OffsetHeadingOptions>({
	addOptions() {
		return {
			...(this.parent?.() as HeadingOptions),
			offset: 0,
		};
	},

	renderHTML({ node, HTMLAttributes }) {
		const documentLevel: number = node.attrs.level;
		const level = Math.min(documentLevel + this.options.offset, 6);
		return [
			`h${level}`,
			mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, { "data-level": documentLevel }),
			0,
		];
	},
});
