import { getNodesBounds, getViewportForBounds, type Node } from "@xyflow/svelte";

/** Retained with the map controller, including while its canvas is unmounted. */
export class GraphFraming {
	private framedFocus?: string;

	initial(nodes: Node[], width: number, height: number, focus = "") {
		if (this.framedFocus === focus) return;
		if (focus && !nodes.some((node) => node.id === focus)) return;
		const viewport = this.fit(nodes, width, height, focus || undefined);
		if (viewport) this.framedFocus = focus;
		return viewport;
	}

	fit(nodes: Node[], width: number, height: number, subjectId?: string) {
		const subject = subjectId ? nodes.find((node) => node.id === subjectId) : undefined;
		const visible = subject ? [subject] : nodes;
		if (
			width <= 0 ||
			height <= 0 ||
			!visible.length ||
			visible.some((node) => !node.measured?.width || !node.measured?.height)
		)
			return;
		return getViewportForBounds(getNodesBounds(visible), width, height, 0.01, 1, 0.2);
	}
}
