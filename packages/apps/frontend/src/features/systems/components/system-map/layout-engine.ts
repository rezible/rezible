import ELK, { type ELK as ElkInstance } from "elkjs/lib/elk-api.js";

import type { GraphSubset } from "$features/systems/lib/system-map/graph";
import type { MapProjection } from "$features/systems/lib/system-map/presentation";
import { layoutWithElk } from "./layout";
import type { LayoutResult } from "./flow-model";

export type SystemMapLayoutEngine = {
	layout: (graph: GraphSubset, projection: MapProjection) => Promise<LayoutResult>;
	dispose: () => void;
};

/** Owns the single lazy ELK worker used by a controller and rejects work on disposal. */
export const createSystemMapLayoutEngine = (): SystemMapLayoutEngine => {
	let elk: ElkInstance | undefined;
	let initializing: Promise<ElkInstance> | undefined;
	let disposed = false;
	const pendingLayouts = new Set<{ reject: (reason?: unknown) => void }>();

	const getElk = (): Promise<ElkInstance> => {
		if (disposed) return Promise.reject(new Error("System map layout engine has been disposed"));
		if (elk) return Promise.resolve(elk);
		if (initializing) return initializing;

		initializing = Promise.resolve()
			.then(() => {
				if (disposed) throw new Error("System map layout engine has been disposed");

				const instance = new ELK({
					algorithms: ["layered"],
					workerFactory: () =>
						new Worker(new URL("./elk.worker.ts", import.meta.url), { type: "module" }),
				});
				elk = instance;
				return instance;
			})
			.finally(() => {
				initializing = undefined;
			});

		return initializing;
	};

	return {
		layout: async (graph, projection) => {
			const instance = await getElk();
			if (disposed) throw new Error("System map layout engine has been disposed");

			return new Promise<LayoutResult>((resolve, reject) => {
				const pending = { reject };
				pendingLayouts.add(pending);
				void layoutWithElk(instance, graph, projection).then(
					(result) => {
						pendingLayouts.delete(pending);
						resolve(result);
					},
					(error: unknown) => {
						pendingLayouts.delete(pending);
						reject(error);
					}
				);
			});
		},
		dispose: () => {
			if (disposed) return;
			disposed = true;
			const error = new Error("System map layout engine has been disposed");
			for (const pending of pendingLayouts) pending.reject(error);
			pendingLayouts.clear();
			elk?.terminateWorker();
			elk = undefined;
		},
	};
};
