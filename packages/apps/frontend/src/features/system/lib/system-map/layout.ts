import type { Box } from "./geometry";
import type { StructureNode, StructureTree } from "./structure";

export const CARD_WIDTH = 200;
export const CARD_HEIGHT = 144;
export const REGION_PADDING = 32;
export const REGION_LABEL_HEIGHT = 40;
export const SIBLING_GAP = 40;
export const ROOT_GAP = 96;
export const MIN_REGION_WIDTH = 160;
/** A region's box is its padded content scaled by this, so the content's corners lie inside its outline. */
export const REGION_SHAPE_SCALE = 1.25;

/** The padded content of a region (label band and children), centred in its inflated box. */
export function regionContentBox(box: Box): Box {
	const width = box.width / REGION_SHAPE_SCALE;
	const height = box.height / REGION_SHAPE_SCALE;
	return {
		x: box.x + (box.width - width) / 2,
		y: box.y + (box.height - height) / 2,
		width,
		height,
	};
}

/** Absolute boxes for every region and card, and the extent of the whole map from the origin. */
export type MapLayout = {
	readonly boxes: ReadonlyMap<string, Box>;
	readonly width: number;
	readonly height: number;
};

/** A packed node or group: its size and the boxes inside it, relative to its own origin. */
type Packed = {
	width: number;
	height: number;
	boxes: Map<string, Box>;
};

function placeInto(target: Map<string, Box>, packed: Packed, x: number, y: number) {
	for (const [id, box] of packed.boxes) {
		target.set(id, { ...box, x: box.x + x, y: box.y + y });
	}
}

/** Fills rows left to right, top-aligned; the column count grows with the square root of the count. */
function packRows(items: readonly Packed[], gap: number): Packed {
	const columns = Math.max(1, Math.round(Math.sqrt(items.length * 1.2)));
	const boxes = new Map<string, Box>();
	let width = 0;
	let y = 0;

	for (let start = 0; start < items.length; start += columns) {
		const row = items.slice(start, start + columns);
		let x = 0;
		let rowHeight = 0;
		for (const item of row) {
			placeInto(boxes, item, x, y);
			x += item.width + gap;
			rowHeight = Math.max(rowHeight, item.height);
		}

		width = Math.max(width, x - gap);
		y += rowHeight + gap;
	}

	const height = items.length > 0 ? y - gap : 0;
	return { width, height, boxes };
}

function packNode(node: StructureNode): Packed {
	if (node.type === "card") {
		const box = { x: 0, y: 0, width: CARD_WIDTH, height: CARD_HEIGHT };
		return { width: CARD_WIDTH, height: CARD_HEIGHT, boxes: new Map([[node.id, box]]) };
	}

	const content = packRows(node.children.map(packNode), SIBLING_GAP);
	const paddedWidth = Math.max(MIN_REGION_WIDTH, content.width + 2 * REGION_PADDING);
	const paddedHeight = content.height + 2 * REGION_PADDING + REGION_LABEL_HEIGHT;
	const width = paddedWidth * REGION_SHAPE_SCALE;
	const height = paddedHeight * REGION_SHAPE_SCALE;
	const insetX = (width - paddedWidth) / 2;
	const insetY = (height - paddedHeight) / 2;

	const boxes = new Map<string, Box>([[node.id, { x: 0, y: 0, width, height }]]);
	placeInto(boxes, content, insetX + REGION_PADDING, insetY + REGION_PADDING + REGION_LABEL_HEIGHT);
	return { width, height, boxes };
}

/** Packs the tree bottom-up; the same tree always produces the same boxes. */
export function layoutStructure(tree: StructureTree): MapLayout {
	const root = packRows(tree.roots.map(packNode), ROOT_GAP);
	return { boxes: root.boxes, width: root.width, height: root.height };
}
