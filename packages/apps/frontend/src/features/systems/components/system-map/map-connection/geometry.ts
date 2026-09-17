import type { ConnectionRoute, ConnectionRouteSection } from "../flow-model";

const pointToSvg = ({ x, y }: { x: number; y: number }): string => `${x} ${y}`;

const sectionToSvg = (section: ConnectionRouteSection): string => {
	const points = [section.start, ...section.bends, section.end];
	return `M${pointToSvg(points[0])}${points.slice(1).map((point) => `L${pointToSvg(point)}`).join("")}`;
};

/**
 * Converts ELK's already-routed orthogonal sections into SVG subpaths.
 * Sections remain separate so no renderer-owned bridge or alternate route is introduced.
 * The target section is emitted last so SVG's marker-end is placed on the actual target.
 */
export const connectionRouteToSvgPath = (route: ConnectionRoute): string => {
	const targetSection = route.sections[route.targetSectionIndex];
	if (!targetSection) throw new Error("System map connection route has no target section");

	return route.sections
		.map((section, index) => ({ section, index }))
		.sort((left, right) => {
			if (left.index === route.targetSectionIndex) return 1;
			if (right.index === route.targetSectionIndex) return -1;
			return left.index - right.index;
		})
		.map(({ section }) => sectionToSvg(section))
		.join("");
};
