<script lang="ts">
	import { onMount } from "svelte";
	import { watch } from "runed";
	import {
		getViewportForBounds,
		useStore,
		useSvelteFlow,
		type CoordinateExtent,
		type FitViewOptions,
		type Viewport,
	} from "@xyflow/svelte";
	import type { Box } from "$features/system/lib/system-map/geometry";
	import type { FlowNode, FlowEdge } from "./projection";
	import type { SystemMapController } from "./controller.svelte";

	type Props = { controller: SystemMapController };
	let { controller }: Props = $props();
	const flow = useSvelteFlow<FlowNode, FlowEdge>();
	const store = useStore<FlowNode, FlowEdge>();

	const FIT_PADDING = 0.1;
	const MAX_FIT_ZOOM = 1;

	/** Whether the current map has had its first fit; reset when the map empties. */
	let fitted = false;

	/** The whole map in view: the opening viewport and the zoom floor. */
	function landscapeViewport(bounds: Box): Viewport {
		return getViewportForBounds(bounds, store.width, store.height, 0, MAX_FIT_ZOOM, FIT_PADDING);
	}

	/** The map plus the fit padding of its tighter dimension, so panning never leaves the map behind. */
	function panExtent(bounds: Box, zoom: number): CoordinateExtent {
		const spareWidth = store.width / zoom - bounds.width;
		const spareHeight = store.height / zoom - bounds.height;
		const margin = Math.max(0, Math.min(spareWidth, spareHeight) / 2);
		return [
			[bounds.x - margin, bounds.y - margin],
			[bounds.x + bounds.width + margin, bounds.y + bounds.height + margin],
		];
	}

	async function showLandscape(viewport: Viewport) {
		await flow.setViewport(viewport, { duration: 0 });
		controller.viewportZoom(viewport.zoom);
	}

	/**
	 * Keeps the zoom floor at the whole-map fit as the graph or container size changes. The map opens at that
	 * fit, and returns to it when the user was at the floor or the floor rose above their zoom.
	 */
	async function syncFloor(bounds: Box | undefined) {
		if (!bounds) {
			fitted = false;
			return;
		}

		const landscape = landscapeViewport(bounds);
		const zoom = flow.getViewport().zoom;
		const atFloor = zoom <= store.minZoom;
		store.setMinZoom(landscape.zoom);
		store.setTranslateExtent(panExtent(bounds, landscape.zoom));

		if (!fitted || atFloor || zoom < landscape.zoom) {
			fitted = true;
			await showLandscape(landscape);
		}
		// After the fit, so the threshold is never judged against the zoom from before it.
		controller.setFitZoom(landscape.zoom);
	}

	watch(
		() => [controller.bounds, store.width, store.height, store.viewportInitialized] as const,
		([bounds, width, height, initialized]) => {
			if (width > 0 && height > 0 && initialized) {
				void syncFloor(bounds);
			}
		}
	);

	onMount(() => {
		controller.setViewportApi({
			fit: async (ids, request) => {
				const bounds = controller.bounds;
				if (!ids && !request?.preserveZoom && bounds) {
					await showLandscape(landscapeViewport(bounds));
					return;
				}

				// The flow's minimum zoom is the floor, so a fit to a subset never zooms out past it.
				const options: FitViewOptions<FlowNode> = {
					padding: FIT_PADDING,
					maxZoom: MAX_FIT_ZOOM,
					duration: 0,
				};

				if (ids) {
					options.nodes = ids.map((id) => ({ id }));
					// Cards are hidden at the landscape band; their layout boxes still locate them.
					options.includeHiddenNodes = true;
				}

				if (request?.minZoom) {
					options.minZoom = request.minZoom;
					options.maxZoom = Math.max(MAX_FIT_ZOOM, request.minZoom);
				}

				if (request?.preserveZoom) {
					const zoom = flow.getViewport().zoom;
					options.minZoom = zoom;
					options.maxZoom = zoom;
				}

				await flow.fitView(options);
				controller.viewportZoom(flow.getViewport().zoom);
			},
		});

		return () => {
			controller.setViewportApi(undefined);
		};
	});
</script>
