export type Point = { x: number; y: number };

/** An absolute world-space rectangle. */
export type Box = { x: number; y: number; width: number; height: number };

export type BoxSide = "top" | "right" | "bottom" | "left";

export type ConnectionAnchor = Point & { side: BoxSide };

/**
 * Both ends, and the unit direction an opposed connection was slid along its sides; zero when not opposed.
 * Labels shift along it so the two labels of an opposed pair sit side by side.
 */
export type ConnectionAnchors = { source: ConnectionAnchor; target: ConnectionAnchor; separation: Point };

/** Separation between two connections that run in opposite directions between the same boxes. */
const OPPOSING_CONNECTION_OFFSET = 8;

const centre = (box: Box): Point => ({ x: box.x + box.width / 2, y: box.y + box.height / 2 });

function sideMidpoint(box: Box, side: BoxSide): Point {
	const middle = centre(box);
	if (side === "top") {
		return { x: middle.x, y: box.y };
	}
	if (side === "bottom") {
		return { x: middle.x, y: box.y + box.height };
	}
	if (side === "left") {
		return { x: box.x, y: middle.y };
	}
	return { x: box.x + box.width, y: middle.y };
}

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max);

/** Moves a point along a box side by a signed distance without leaving the side. */
function slideAlongSide(box: Box, side: BoxSide, point: Point, distance: number): Point {
	if (side === "left" || side === "right") {
		return { x: point.x, y: clamp(point.y + distance, box.y, box.y + box.height) };
	}
	return { x: clamp(point.x + distance, box.x, box.x + box.width), y: point.y };
}

/**
 * Attaches each end to the midpoint of the side of its box that faces the other box, chosen by the larger of
 * the horizontal and vertical centre distance. An opposed connection slides both ends along their sides
 * towards the right of its direction of travel, so the two directions of a pair separate.
 */
export function connectionAnchors(source: Box, target: Box, opposed: boolean): ConnectionAnchors {
	const from = centre(source);
	const to = centre(target);
	const dx = to.x - from.x;
	const dy = to.y - from.y;

	let sourceSide: BoxSide;
	let targetSide: BoxSide;
	let rightOfTravel: Point;
	if (Math.abs(dx) >= Math.abs(dy)) {
		sourceSide = dx >= 0 ? "right" : "left";
		targetSide = dx >= 0 ? "left" : "right";
		rightOfTravel = { x: 0, y: dx >= 0 ? 1 : -1 };
	} else {
		sourceSide = dy >= 0 ? "bottom" : "top";
		targetSide = dy >= 0 ? "top" : "bottom";
		rightOfTravel = { x: dy >= 0 ? -1 : 1, y: 0 };
	}

	let separation: Point = { x: 0, y: 0 };
	if (opposed) {
		separation = rightOfTravel;
	}
	// One component is zero; the other is the signed slide along both sides.
	const offset = (separation.x + separation.y) * OPPOSING_CONNECTION_OFFSET;

	const sourcePoint = slideAlongSide(source, sourceSide, sideMidpoint(source, sourceSide), offset);
	const targetPoint = slideAlongSide(target, targetSide, sideMidpoint(target, targetSide), offset);
	return {
		source: { ...sourcePoint, side: sourceSide },
		target: { ...targetPoint, side: targetSide },
		separation,
	};
}

/** The superellipse exponent of a region outline: 2 is an ellipse, higher values approach the box. */
const REGION_EXPONENT = 4;
/** How far the outline's radius swells and shrinks around the superellipse. */
const REGION_WOBBLE = 0.03;
const REGION_WOBBLE_LOBES = 3;
/** Points on the curve; the parametric form spaces them closer together where it turns at the corners. */
const REGION_CURVE_POINTS = 96;
/** Straight segments drawn for each spline span between two curve points. */
const REGION_SPAN_SEGMENTS = 4;

/** A stable phase in [0, 2π) for an id, so each region wobbles differently but always the same way. */
function wobblePhase(id: string): number {
	let hash = 2166136261;
	for (let index = 0; index < id.length; index++) {
		hash ^= id.charCodeAt(index);
		hash = Math.imul(hash, 16777619);
	}
	return ((hash >>> 0) / 2 ** 32) * 2 * Math.PI;
}

/** `sign(value) · |value|^power`, which keeps the parametric superellipse in the value's quadrant. */
const signedPower = (value: number, power: number) => Math.sign(value) * Math.abs(value) ** power;

/**
 * Points on a superellipse inscribed in the box, each pushed out from the centre by `1 + 0.03 · sin(3θ + φ)`
 * at its polar angle θ, with the phase φ taken from the id. The swell can reach slightly past the box edge.
 */
function regionCurve(box: Box, id: string): Point[] {
	const middle = centre(box);
	const phase = wobblePhase(id);
	const power = 2 / REGION_EXPONENT;
	const points: Point[] = [];

	for (let index = 0; index < REGION_CURVE_POINTS; index++) {
		const t = (index / REGION_CURVE_POINTS) * 2 * Math.PI;
		const x = (box.width / 2) * signedPower(Math.cos(t), power);
		const y = (box.height / 2) * signedPower(Math.sin(t), power);
		const angle = Math.atan2(y, x);
		const wobble = 1 + REGION_WOBBLE * Math.sin(REGION_WOBBLE_LOBES * angle + phase);
		points.push({ x: middle.x + x * wobble, y: middle.y + y * wobble });
	}

	return points;
}

/**
 * The region outline as drawn: a closed Catmull-Rom spline through the curve points, flattened to straight
 * segments so containment can be checked against exactly what is on screen.
 */
export function regionOutline(box: Box, id: string): Point[] {
	const curve = regionCurve(box, id);
	const count = curve.length;
	const at = (index: number) => curve[(index + count) % count];
	const outline: Point[] = [];

	for (let index = 0; index < count; index++) {
		const p0 = at(index - 1);
		const p1 = at(index);
		const p2 = at(index + 1);
		const p3 = at(index + 2);
		for (let step = 0; step < REGION_SPAN_SEGMENTS; step++) {
			const s = step / REGION_SPAN_SEGMENTS;
			const s2 = s * s;
			const s3 = s2 * s;
			const blend = (a: number, b: number, c: number, d: number) =>
				0.5 * (2 * b + (c - a) * s + (2 * a - 5 * b + 4 * c - d) * s2 + (3 * b - a - 3 * c + d) * s3);
			outline.push({ x: blend(p0.x, p1.x, p2.x, p3.x), y: blend(p0.y, p1.y, p2.y, p3.y) });
		}
	}

	return outline;
}

const round = (value: number) => Math.round(value * 100) / 100;

/** The region outline as a closed SVG path. */
export function regionPath(box: Box, id: string): string {
	const [first, ...rest] = regionOutline(box, id);
	const commands = [`M ${round(first.x)} ${round(first.y)}`];
	for (const point of rest) {
		commands.push(`L ${round(point.x)} ${round(point.y)}`);
	}

	commands.push("Z");
	return commands.join(" ");
}
