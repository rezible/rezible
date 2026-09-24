<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Separator } from "$components/ui/separator";
	import type { MapSelection } from "$features/system/lib/system-map/presentation";
	import { buildInspectionItems, resolveInspection, type MapInspectorItem } from "./inspection";
	import { buildInspectionViewData, isInspectorItemSelected, loadedCountsText } from "./presentation";
	import type { MapInspectorProps } from "./types";

	let { graph, selection, onSelectionChange }: MapInspectorProps = $props();

	const inspection = $derived(resolveInspection(graph, selection));
	const inspectionView = $derived(buildInspectionViewData(inspection));
	const items = $derived(buildInspectionItems(graph));
	const relationshipCount = $derived(graph.relationships.length + graph.unresolvedRelationships.length);
	const countsText = $derived(loadedCountsText(graph.entities.length, relationshipCount));
	const isPartial = $derived(
		graph.enumeration.stopReason !== "exhausted" || graph.unresolvedRelationships.length > 0
	);

	function select(target: MapSelection, event: MouseEvent) {
		onSelectionChange(target, event.currentTarget as HTMLElement);
	}

	function clearSelection(event: MouseEvent) {
		onSelectionChange(undefined, event.currentTarget as HTMLElement);
	}

	function selectRelationship(relationshipId: string, event: MouseEvent) {
		select({ kind: "relationship", relationshipId }, event);
	}

	function selectEntity(entityId: string, event: MouseEvent) {
		select({ kind: "entity", entityId }, event);
	}

	function isSelected(item: MapInspectorItem): boolean {
		return isInspectorItemSelected(item, selection);
	}
</script>

<aside
	class="bg-card border-border flex w-80 min-h-0 min-w-0 shrink-0 flex-col border-l"
	aria-label="Map inspector"
>
	<div class="flex shrink-0 flex-col gap-2 p-3">
		<div class="flex items-start justify-between gap-2">
			<div class="min-w-0">
				<h2 class="text-sm font-semibold">Map inspector</h2>
				<p class="text-muted-foreground text-xs">{countsText}</p>
			</div>
			{#if selection}
				<Button variant="ghost" size="sm" onclick={clearSelection}>Clear</Button>
			{/if}
		</div>
		{#if isPartial}
			<Alert.Root>
				<Alert.Description>
					Partial supplied subset. Counts and relationships describe only the records available
					here.
				</Alert.Description>
			</Alert.Root>
		{/if}
	</div>

	{#if selection && inspectionView}
		<div class="min-h-0 max-h-[55%] shrink-0 overflow-y-auto">
			<section class="flex flex-col gap-3 p-3" aria-label="Selected source inspection">
				<div class="flex items-start justify-between gap-2">
					<h3 class="break-words text-sm font-semibold">{inspectionView.title}</h3>
					{#if inspectionView.selectionAvailable}
						<Badge variant="secondary">Supplied</Badge>
					{:else}
						<Badge variant="outline">Unavailable</Badge>
					{/if}
				</div>

				{#if inspectionView.entity}
					<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1 text-xs">
						<dt class="text-muted-foreground">ID</dt>
						<dd class="wrap-anywhere font-mono">{inspectionView.entity.id}</dd>
						<dt class="text-muted-foreground">Category</dt>
						<dd class="wrap-anywhere">{inspectionView.entity.category}</dd>
						<dt class="text-muted-foreground">Kind</dt>
						<dd class="wrap-anywhere">{inspectionView.entity.kind}</dd>
					</dl>
					{#if inspectionView.incidentRelationships.length > 0}
						<div class="flex flex-col gap-1">
							<h4 class="text-xs font-medium uppercase">Incident relationships</h4>
							<ul class="flex flex-col gap-1 text-xs">
								{#each inspectionView.incidentRelationships as relationship (relationship.id)}
									<li>
										<Button
											variant="link"
											size="xs"
											class="h-auto max-w-full justify-start whitespace-normal px-0 text-left"
											onclick={(event) => selectRelationship(relationship.id, event)}
										>
											{relationship.text}
										</Button>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				{:else if inspectionView.relationship}
					<div class="flex flex-col gap-2 text-xs">
						<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1">
							<dt class="text-muted-foreground">ID</dt>
							<dd class="wrap-anywhere font-mono">{inspectionView.relationship.id}</dd>
							<dt class="text-muted-foreground">Predicate</dt>
							<dd class="wrap-anywhere">{inspectionView.relationship.predicate}</dd>
						</dl>
						<div class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1">
							<span class="text-muted-foreground">Source</span>
							{#if inspectionView.relationship.source.entity}
								<Button
									variant="link"
									size="xs"
									class="h-auto justify-start whitespace-normal px-0 text-left"
									onclick={(event) =>
										selectEntity(inspectionView.relationship!.source.id, event)}
								>
									{inspectionView.relationship.source.text}
								</Button>
							{:else}
								<span class="wrap-anywhere">{inspectionView.relationship.source.text}</span>
							{/if}
							<span class="text-muted-foreground">Target</span>
							{#if inspectionView.relationship.target.entity}
								<Button
									variant="link"
									size="xs"
									class="h-auto justify-start whitespace-normal px-0 text-left"
									onclick={(event) =>
										selectEntity(inspectionView.relationship!.target.id, event)}
								>
									{inspectionView.relationship.target.text}
								</Button>
							{:else}
								<span class="wrap-anywhere">{inspectionView.relationship.target.text}</span>
							{/if}
						</div>
					</div>
				{:else if inspectionView.summary}
					<div class="flex flex-col gap-2">
						<div class="flex items-center justify-between gap-2">
							<h4 class="text-xs font-medium uppercase">Summary contributors</h4>
							<Badge variant="secondary">
								{inspectionView.summary.distinctContributorCount} distinct
							</Badge>
						</div>
						<ul class="flex flex-col gap-1 text-xs">
							{#each inspectionView.summary.relationships as contributor (contributor.id)}
								<li>
									<Button
										variant="link"
										size="xs"
										class="h-auto max-w-full justify-start whitespace-normal px-0 text-left"
										onclick={(event) => selectRelationship(contributor.id, event)}
									>
										{contributor.text}
									</Button>
								</li>
							{:else}
								<li class="text-muted-foreground text-xs">
									No contributors are available in this subset.
								</li>
							{/each}
						</ul>
						{#if inspectionView.summary.missingRelationshipIds.length > 0}
							<Alert.Root>
								<Alert.Description>
									Contributors missing from this subset:
									{inspectionView.summary.missingRelationshipIds.join(", ")}
								</Alert.Description>
							</Alert.Root>
						{/if}
					</div>
				{:else}
					<Alert.Root>
						<Alert.Description>
							The selected subject is not supplied in this graph.
						</Alert.Description>
					</Alert.Root>
				{/if}
			</section>
		</div>
		<Separator />
	{/if}

	<section class="min-h-0 flex-1 overflow-y-auto" aria-label="Supplied graph subjects">
		<div class="flex flex-col gap-1 p-2">
			{#each items as item (item.id)}
				<Button
					variant={isSelected(item) ? "secondary" : "ghost"}
					size="sm"
					class="h-auto min-w-0 justify-start whitespace-normal py-2 text-left"
					aria-current={isSelected(item) ? "true" : undefined}
					onclick={(event) => select(item.target, event)}
				>
					<span class="flex min-w-0 flex-col gap-0.5">
						<span class="break-words">{item.label}</span>
						<span class="text-muted-foreground break-words text-xs">{item.detail}</span>
					</span>
				</Button>
			{:else}
				<p class="text-muted-foreground p-3 text-xs">No source records are supplied.</p>
			{/each}
		</div>
	</section>
</aside>
