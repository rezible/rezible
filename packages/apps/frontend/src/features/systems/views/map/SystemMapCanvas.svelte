<script lang="ts">
	import {
		Background,
		BackgroundVariant,
		Controls,
		MiniMap,
		Panel,
		SvelteFlow,
		type ColorMode,
		type SvelteFlowProps,
		type ControlsProps,
		type BackgroundProps,
		type MiniMapProps,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import * as ToggleGroup from "$components/ui/toggle-group";
	import { cn } from "$lib/utils";
	import * as Select from "$components/ui/select";
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";
	import * as Table from "$components/ui/table";
	import RiRestartLine from "remixicon-svelte/icons/restart-line";
	import RiListView from "remixicon-svelte/icons/list-view";
	import RiPieChartLine from "remixicon-svelte/icons/pie-chart-line";
	import { mode } from "mode-watcher";
	import {
		SystemMapViewController,
		provideSystemMapViewController,
		makeEntityLabel,
		type SystemMapEdgeData,
	} from "./controller.svelte";
	import SystemMapDetails from "./SystemMapDetails.svelte";
	import SystemMapEntityNode from "./SystemMapEntityNode.svelte";
	import SystemMapSearch from "./SystemMapSearch.svelte";

	type Props = { defaultFocus?: () => string | undefined };
	let { defaultFocus }: Props = $props();
	const localController = new SystemMapViewController(() => defaultFocus?.() ?? "");
	const view = provideSystemMapViewController(localController);
	let mapRoot = $state<HTMLElement>();
	$effect(() => view.setFocusFallback(mapRoot));

	const colorMode = $derived<ColorMode>(mode.current === "dark" ? "dark" : "light");
	const flowSettings: SvelteFlowProps = {
		nodeTypes: {
			default: SystemMapEntityNode,
			entity: SystemMapEntityNode,
		},
		snapGrid: [25, 25],
		connectionRadius: 40,
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

	const formatFreshness = (iso: string) => {
		const date = new Date(iso);
		const diff = Date.now() - date.getTime();
		const minutes = Math.round(diff / 60_000);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.round(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		return `${Math.round(hours / 24)}d ago`;
	};
</script>

<section
	class="flex min-h-0 flex-1 flex-col overflow-hidden"
	bind:this={mapRoot}
	tabindex="-1"
	aria-label="System map"
>
	<header class="border-border flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3">
		<div class="flex min-w-0 flex-wrap items-center gap-2">
			<SystemMapSearch />
			<Select.Root
				type="single"
				value={String(view.depth)}
				onValueChange={(value) => value && view.setDepth(Number(value))}
			>
				<Select.Trigger class="w-36" aria-label="Neighborhood depth">
					Depth {view.depth}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each [1, 2, 3, 4] as depthOption (depthOption)}
							<Select.Item value={String(depthOption)}>Depth {depthOption}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
			<Select.Root
				type="single"
				value={view.kindFilter === "" ? "all" : view.kindFilter}
				onValueChange={(value) => view.setKindFilter(value === "all" ? "" : (value ?? ""))}
			>
				<Select.Trigger class="w-44" aria-label="Filter by subject type">
					{view.kindFilter === "" ? "All subject types" : view.kindFilter.replaceAll("_", " ")}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						<Select.Item value="all">All subject types</Select.Item>
						{#each view.availableKinds as kindOption (kindOption)}
							<Select.Item value={kindOption}>{kindOption.replaceAll("_", " ")}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
			<Select.Root
				type="single"
				value={view.predicateFilter === "" ? "all" : view.predicateFilter}
				onValueChange={(value) => view.setPredicateFilter(value === "all" ? "" : (value ?? ""))}
			>
				<Select.Trigger class="w-44" aria-label="Filter by relationship type">
					{view.predicateFilter === ""
						? "All relationships"
						: view.predicateFilter.replaceAll("_", " ")}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						<Select.Item value="all">All relationships</Select.Item>
						{#each view.availablePredicates as predicateOption (predicateOption)}
							<Select.Item value={predicateOption}>
								{predicateOption.replaceAll("_", " ")}
							</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
			{#if view.hasFilters}
				<Button variant="ghost" size="sm" onclick={() => view.clearFilters()}>Clear filters</Button>
			{/if}
		</div>
		<ToggleGroup.Root
			type="single"
			value={view.displayMode}
			onValueChange={(value) => value && view.setDisplayMode(value as "graph" | "list")}
			aria-label="Map display"
		>
			<ToggleGroup.Item value="graph"><RiPieChartLine />Graph</ToggleGroup.Item>
			<ToggleGroup.Item value="list"><RiListView />List</ToggleGroup.Item>
		</ToggleGroup.Root>
	</header>

	{#if view.error && view.hasGraph}
		<Alert.Root variant="destructive"
			><Alert.Title>Map refresh failed</Alert.Title><Alert.Description
				>{view.error.detail}</Alert.Description
			><Alert.Action
				><Button variant="outline" size="sm" onclick={view.retry}>Retry</Button></Alert.Action
			></Alert.Root
		>
	{/if}
	{#if view.selectionLoading}
		<p role="status" class="px-4 py-2 text-sm text-muted-foreground">Loading selected subject…</p>
	{:else if view.selectionError}
		<Alert.Root variant="destructive">
			<Alert.Title>Selected subject unavailable</Alert.Title>
			<Alert.Description>{view.selectionError.detail || view.selectionError.title}</Alert.Description>
			<Alert.Action
				><Button variant="outline" size="sm" onclick={view.retrySelection}>Retry</Button
				></Alert.Action
			>
		</Alert.Root>
	{/if}
	{#if view.focusMissing && view.hasGraph}
		<Alert.Root
			><Alert.Title>Focused subject unavailable</Alert.Title><Alert.Description
				>The requested subject was not returned. Previously loaded relationships remain available.</Alert.Description
			><Alert.Action
				><Button variant="outline" size="sm" onclick={view.retry}>Retry</Button></Alert.Action
			></Alert.Root
		>
	{/if}
	<div
		class="bg-muted/20 relative min-h-64 flex-1"
		bind:clientWidth={view.canvasWidth}
		bind:clientHeight={view.canvasHeight}
	>
		{#if view.error && !view.hasGraph}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root variant="destructive" class="w-fit max-w-md">
					<Alert.Title>Could not load the system map</Alert.Title>
					<Alert.Description>{view.error.detail}</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" onclick={view.retry}>Retry</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if view.focusMissing && !view.hasGraph}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root class="w-fit max-w-md">
					<Alert.Title>That subject is not on the map</Alert.Title>
					<Alert.Description>
						The linked subject could not be found or you are not authorized to see it. Search for
						a subject to explore its neighborhood.
					</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" onclick={() => view.reset()}>
							Start from the default view
						</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if !view.hasGraph && !view.loading}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root class="w-fit max-w-md">
					<Alert.Title>No mapped systems yet</Alert.Title>
					<Alert.Description>
						Once providers and mappings are connected, systems and their relationships appear
						here. Search for a subject to focus the map on it.
					</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" href="/settings/integrations">
							Review integrations
						</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if view.displayMode === "list"}
			<div class="bg-card absolute inset-0 overflow-y-auto p-4">
				<section class="flex flex-col gap-2">
					<h2 class="text-foreground text-sm font-medium">
						Subjects ({view.displayEntities.length})
					</h2>
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Name</Table.Head>
								<Table.Head>Type</Table.Head>
								<Table.Head>Connections</Table.Head>
								<Table.Head>Last updated</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each view.displayEntities as entity (entity.id)}
								<Table.Row class={cn(view.selectedId === entity.id && "bg-selection")}>
									<Table.Cell
										><Button variant="link" onclick={() => view.selectEntity(entity)}
											>{makeEntityLabel(entity)}</Button
										></Table.Cell
									>
									<Table.Cell>{entity.attributes.kind.replaceAll("_", " ")}</Table.Cell>
									<Table.Cell>
										{view.connectionCount(entity.id)}
									</Table.Cell>
									<Table.Cell>{formatFreshness(entity.attributes.updatedAt)}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</section>
				<section class="mt-6 flex flex-col gap-2">
					<h2 class="text-foreground text-sm font-medium">
						Relationships ({view.displayRelationships.length})
					</h2>
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Source</Table.Head>
								<Table.Head>Relationship</Table.Head>
								<Table.Head>Target</Table.Head>
								<Table.Head>Last updated</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each view.displayRelationships as relationship (relationship.id)}
								<Table.Row
									class={cn(
										view.inspectedRelationship?.id === relationship.id && "bg-selection"
									)}
								>
									<Table.Cell>
										{view.entityLabel(relationship.attributes.sourceEntityId)}
									</Table.Cell>
									<Table.Cell>
										<Button
											variant="link"
											onclick={() => view.selectRelationship(relationship)}
											>{relationship.attributes.predicate.replaceAll("_", " ")}</Button
										>
									</Table.Cell>
									<Table.Cell>
										{view.entityLabel(relationship.attributes.targetEntityId)}
									</Table.Cell>
									<Table.Cell
										>{formatFreshness(relationship.attributes.updatedAt)}</Table.Cell
									>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</section>
			</div>
		{:else}
			<SvelteFlow
				{...flowSettings}
				{colorMode}
				class="rezible-flow"
				minZoom={0.01}
				bind:nodes={view.nodes}
				bind:edges={view.edges}
				bind:viewport={view.viewport}
				nodesDraggable={true}
				nodesConnectable={false}
				elementsSelectable
				onedgeclick={({ edge }) => {
					const relationship = (edge.data as SystemMapEdgeData | undefined)?.relationship;
					if (relationship) view.selectRelationship(relationship);
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
								Loading neighborhood…
							</div>
						{/if}
						{#if view.truncated}
							<div
								class="bg-background border-border max-w-xs border px-3 py-2 text-sm shadow-sm"
							>
								<Badge variant="outline" class="mb-1">Partial coverage</Badge>
								<p class="text-muted-foreground text-xs">
									This neighborhood is larger than the map loads at once. Expand specific
									subjects to load more of it.
								</p>
							</div>
						{/if}
					</div>
				</Panel>
				<Panel position="bottom-right">
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onclick={() => view.recenter()}>Recenter</Button>
						<Button variant="outline" size="sm" onclick={() => view.reset()}>
							<RiRestartLine data-icon="inline-start" />
							Reset
						</Button>
					</div>
				</Panel>
			</SvelteFlow>
		{/if}
		<SystemMapDetails />
	</div>
</section>
