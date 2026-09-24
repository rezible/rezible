import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";

export type SystemMapProps = {
	graph: GraphSubset;
	/** Optional sparse starting hints; the map may adjust final coordinates. */
	positions?: Readonly<Partial<Record<string, Point>>>;
	selection?: MapSelection;
	highlights?: MapHighlights;
	/** Emits the user's selection intention without owning an independent selection. */
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
	/** Enables compact-node dragging; called once with the requested position when a drag completes. */
	onNodeMove?: (entityId: string, position: Point) => void;
};

export type SystemMapHandle = {
	/** No-ops before a usable layout exists. */
	fit(): void;
	recenter(): void;
	/** No-ops before a usable layout or when the target is unavailable; commands are not queued. */
	reveal(target: MapSelection): void;
};
