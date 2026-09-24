import type { MapSelection } from "$features/system/lib/system-map/presentation";

export type SystemMapHandle = {
	/** No-ops before a usable layout exists. */
	fit(): void;
	recenter(): void;
	/** No-ops before a usable layout or when the target is unavailable; commands are not queued. */
	reveal(target: MapSelection): void;
};
