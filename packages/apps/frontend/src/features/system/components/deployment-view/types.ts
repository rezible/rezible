import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { MapSelection } from "$features/system/lib/system-map/presentation";

export type DeploymentViewProps = {
	graph: GraphSubset;
	selection?: MapSelection;
	onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
	/** A placement target to scroll to and highlight. */
	focusTargetId?: string;
	onShowInStructure: (entityId: string) => void;
};
