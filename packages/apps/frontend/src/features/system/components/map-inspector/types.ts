import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { MapSelection } from "$features/system/lib/system-map/presentation";

export type MapInspectorProps = {
	graph: GraphSubset;
	selection?: MapSelection;
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
};
