<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { MapInspector } from "$features/system/components/map-inspector";
	import { SystemMap, type SystemMapHandle } from "$features/system/components/system-map";
	import { initSystemMapViewController } from "./controller.svelte";
	import MapToolbar from "./components/map-toolbar/MapToolbar.svelte";

	const view = initSystemMapViewController();
	let mapHandle = $state<SystemMapHandle>();
</script>

<section class="flex h-full min-h-0 min-w-0 flex-1 flex-col overflow-hidden" aria-label="System map">
	<MapToolbar map={mapHandle} />

	{#if view.showingPriorQuery}
		<div
			role="status"
			aria-live="polite"
			class="bg-muted/50 border-border shrink-0 border-b px-3 py-2 text-sm"
		>
			Showing previous results for <strong>{view.priorQueryLabel}</strong>
			. Requested filters:
			<strong>{view.requestedQueryLabel}</strong>
			.
		</div>
	{/if}

	{#if view.loaded}
		{#if view.graphQuery.status === "loading"}
			<p role="status" aria-live="polite" class="text-muted-foreground shrink-0 px-3 py-2 text-sm">
				Refreshing supplied records…
			</p>
		{/if}

		{#if view.graphQuery.error}
			<div class="flex shrink-0 items-start gap-2 px-3 pt-2">
				<div class="min-w-0 flex-1">
					<ErrorAlert error={view.graphQuery.error} dismissable={false} />
				</div>
				<Button variant="outline" size="sm" onclick={view.refresh}>Retry</Button>
			</div>
		{/if}

		<div
			class="text-muted-foreground flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 px-3 py-2 text-xs"
		>
			<span>{view.graphQuery.entityCount} loaded entities</span>
			<span>{view.graphQuery.relationshipCount} resolved relationships</span>
			<span>{view.graphQuery.unresolvedRelationshipCount} unresolved relationships</span>
			{#if view.graphQuery.status === "success"}
				<Badge variant="outline">Loaded</Badge>
			{/if}
		</div>
		{#if view.graph.entities.length > 0 && !view.hasArchitecturalEntities}
			<p
				role="status"
				class="bg-muted/30 border-border text-muted-foreground shrink-0 border-y px-3 py-2 text-sm"
			>
				No architectural entities are available for the canvas. Use the Inspector to browse the
				supplied source records.
			</p>
		{/if}

		{#if view.graphQuery.partial && view.partialDescription}
			<div
				role="status"
				class="bg-muted/30 border-border flex shrink-0 items-start gap-2 border-y px-3 py-2 text-sm"
			>
				<Badge variant="outline">Partial result</Badge>
				<p class="text-muted-foreground min-w-0 flex-1">{view.partialDescription}</p>
			</div>
		{/if}

		<div class="flex min-h-0 min-w-0 flex-1">
			<div class="min-h-0 min-w-0 flex-1">
				<SystemMap
					bind:this={mapHandle}
					graph={view.graph}
					positions={view.positions}
					selection={view.selection}
					onSelectionChange={view.setSelection}
					onNodeMove={view.moveNode}
				/>
			</div>
			{#if view.inspectorOpen}
				<MapInspector
					graph={view.graph}
					selection={view.selection}
					onSelectionChange={view.setSelection}
				/>
			{/if}
		</div>
	{:else if view.graphQuery.status === "error"}
		<div class="grid min-h-0 flex-1 place-items-center p-4">
			<div class="flex w-full max-w-xl flex-col gap-2">
				<ErrorAlert error={view.graphQuery.error} dismissable={false} />
				<Button variant="outline" size="sm" class="self-start" onclick={view.refresh}>Retry</Button>
			</div>
		</div>
	{:else}
		<div
			role="status"
			aria-live="polite"
			class="text-muted-foreground grid min-h-0 flex-1 place-items-center"
		>
			<div class="flex items-center gap-2">
				<Spinner />
				<span>Loading graph records…</span>
			</div>
		</div>
	{/if}
</section>
