<script lang="ts">
	import { Background, BackgroundVariant, Controls, MiniMap, Panel, SvelteFlow, type ColorMode, type SvelteFlowProps, type ControlsProps, type BackgroundProps, type MiniMapProps } from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";
	import * as Alert from "$components/ui/alert";
	import * as Button from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import RiHome4Line from "remixicon-svelte/icons/home-4-line";
	import RiRestartLine from "remixicon-svelte/icons/restart-line";
	import { initSystemMapViewController } from "./controller.svelte";
	import SystemMapDetails from "./SystemMapDetails.svelte";
	import SystemMapEntityNode from "./SystemMapEntityNode.svelte";
	import SystemMapSearch from "./SystemMapSearch.svelte";

	const view = initSystemMapViewController();

	const colorMode = $derived<ColorMode>("dark");
	const flowSettings: SvelteFlowProps = {
		nodeTypes: {
			default: SystemMapEntityNode,
			entity: SystemMapEntityNode,
		},
		snapGrid: [25, 25],
		connectionRadius: 40,
		fitView: true,
		proOptions: { hideAttribution: true },
	};

	const backgroundSettings: BackgroundProps = {
		variant: BackgroundVariant.Dots,
	};

	const controlsSettings: ControlsProps = {
		position: "top-left",
	};

	const minimapSettings: MiniMapProps = {
		position: "top-right",
	};
</script>

<section class="flex min-h-0 flex-1 flex-col overflow-hidden">
	<header class="border-border flex items-center justify-between gap-4 border-b px-4 py-3">
		<div>
			<h1 class="text-lg font-semibold">Explore</h1>
			<p class="text-muted-foreground text-sm">System Map</p>
		</div>
		<div class="flex items-center gap-2">
			<SystemMapSearch />
		</div>
	</header>

	<div class="bg-muted/20 relative min-h-0 flex-1">
		{#if view.error}
			<Alert.Root variant="destructive" class="absolute top-3 left-1/2 z-20 w-fit -translate-x-1/2">
				<Alert.Title>Could not load the system map</Alert.Title>
				<Alert.Description>{view.error.detail}</Alert.Description>
			</Alert.Root>
		{:else}
			<SvelteFlow
				{...flowSettings}
				{colorMode}
				bind:nodes={view.nodes}
				bind:edges={view.edges}
				bind:viewport={view.viewport}
				nodesDraggable={true}
				nodesConnectable={false}
				elementsSelectable
				minZoom={0.2}
				maxZoom={2}
				onedgeclick={({ edge }) => {
					// const relationship = edge.data?.relationship;
					// if (relationship) view.selectRelationship(relationship);
				}}
				onpaneclick={() => view.clearSelection()}
			>
				<Background {...backgroundSettings} />
				<Controls {...controlsSettings} />
				<MiniMap {...minimapSettings} />
				<Panel position="bottom-left">
					<div class="flex flex-col gap-2">
						{#if view.loading}
							<div
								class="bg-background border-border flex items-center gap-2 border px-3 py-2 text-sm shadow-sm"
							>
								<Spinner />
								Loading graph…
							</div>
						{/if}
					</div>
				</Panel>
				<Panel position="bottom-right">
					<Button.Root variant="outline" size="sm" onclick={() => view.reset()}>
						<RiRestartLine />
						Reset
					</Button.Root>
				</Panel>
			</SvelteFlow>
			<SystemMapDetails />
		{/if}
	</div>
</section>
