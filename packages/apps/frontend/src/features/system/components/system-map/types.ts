import type { MapSelection } from "$features/system/lib/system-map/presentation";

export type SystemMapHandle = {
	fit(): void;
	recenter(): void;
	/** No-ops when the target is not drawn; commands are not queued. */
	reveal(target: MapSelection): void;
};
