import type { ConnectionRoute, ConnectionRouteSection } from "../flow-model";
import type { Point } from "$features/system/lib/system-map/geometry";

const pointToSvg = ({ x, y }: { x: number; y: number }): string => `${x} ${y}`;

const sectionToSvg = (section: ConnectionRouteSection): string => {
	const points = [section.start, ...section.bends, section.end];
	return `M${pointToSvg(points[0])}${points
		.slice(1)
		.map((point) => `L${pointToSvg(point)}`)
		.join("")}`;
};

/** Converts routed orthogonal sections into SVG subpaths and keeps the target section last for its marker. */
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

const samePoint = (left: Point, right: Point) => left.x === right.x && left.y === right.y;

/** Adds short orthogonal leads from moved compact nodes to the displayed route during a drag. */
export const connectionRouteWithNodeFeedback = (
	route: ConnectionRoute,
	sourceDelta?: Point,
	targetDelta?: Point
): ConnectionRoute => {
	const hasSourceDelta = sourceDelta && (sourceDelta.x !== 0 || sourceDelta.y !== 0);
	const hasTargetDelta = targetDelta && (targetDelta.x !== 0 || targetDelta.y !== 0);
	if (!hasSourceDelta && !hasTargetDelta) return route;

	const sections = route.sections.slice() as ConnectionRouteSection[];
	if (hasSourceDelta) {
		const section = sections[route.sourceSectionIndex];
		if (section) {
			const previousStart = section.start;
			const nextStart = { x: previousStart.x + sourceDelta.x, y: previousStart.y + sourceDelta.y };
			const firstAfterStart = section.bends[0] ?? section.end;
			const isHorizontal = previousStart.y === firstAfterStart.y;
			const corner = isHorizontal
				? { x: previousStart.x, y: nextStart.y }
				: { x: nextStart.x, y: previousStart.y };
			const lead = [corner, previousStart].filter(
				(point, index, points) =>
					!samePoint(point, nextStart) && (index === 0 || !samePoint(point, points[index - 1]))
			);
			sections[route.sourceSectionIndex] = {
				...section,
				start: nextStart,
				bends: [...lead, ...section.bends],
			};
		}
	}

	if (hasTargetDelta) {
		const section = sections[route.targetSectionIndex];
		if (section) {
			const previousEnd = section.end;
			const nextEnd = { x: previousEnd.x + targetDelta.x, y: previousEnd.y + targetDelta.y };
			const lastBeforeEnd = section.bends.at(-1) ?? section.start;
			const isHorizontal = previousEnd.y === lastBeforeEnd.y;
			const corner = isHorizontal
				? { x: nextEnd.x, y: previousEnd.y }
				: { x: previousEnd.x, y: nextEnd.y };
			const lead = [previousEnd, corner].filter(
				(point, index, points) =>
					!samePoint(point, nextEnd) && (index === 0 || !samePoint(point, points[index - 1]))
			);
			sections[route.targetSectionIndex] = {
				...section,
				bends: [...section.bends, ...lead],
				end: nextEnd,
			};
		}
	}

	return { ...route, sections };
};
