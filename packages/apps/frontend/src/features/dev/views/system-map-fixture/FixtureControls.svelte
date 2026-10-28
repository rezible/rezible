<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import * as Collapsible from "$components/ui/collapsible";
	import * as ToggleGroup from "$components/ui/toggle-group";
	import { useSystemMapFixtureController } from "./controller.svelte";

	const controller = useSystemMapFixtureController();
</script>

<div class="bg-card border-border flex shrink-0 flex-col gap-2 rounded-md border p-2">
	<div class="flex flex-wrap items-center gap-2">
		<ToggleGroup.Root
			type="single"
			value={controller.scenario}
			onValueChange={controller.setScenario}
			aria-label="Fixture scenario"
			variant="outline"
			size="sm"
			class="max-w-full flex-wrap"
		>
			<ToggleGroup.Item value="terrain">Terrain</ToggleGroup.Item>
			<ToggleGroup.Item value="large">Large</ToggleGroup.Item>
			<ToggleGroup.Item value="empty">Empty graph</ToggleGroup.Item>
		</ToggleGroup.Root>

		<ToggleGroup.Root
			type="single"
			value={controller.mapView}
			onValueChange={controller.setMapView}
			aria-label="Map view"
			variant="outline"
			size="sm"
		>
			<ToggleGroup.Item value="structure">Structure</ToggleGroup.Item>
			<ToggleGroup.Item value="deployment">Deployment</ToggleGroup.Item>
		</ToggleGroup.Root>

		{#if controller.partial}
			<Badge variant="secondary" class="ml-auto">Partial supplied subset</Badge>
		{/if}
	</div>

	<Collapsible.Root class="border-border border-t pt-2 text-xs">
		<Collapsible.Trigger
			class="text-muted-foreground w-full text-left hover:text-foreground focus-visible:outline-2 focus-visible:outline-ring"
		>
			Diagnostics
		</Collapsible.Trigger>
		<Collapsible.Content>
			<div class="flex flex-wrap gap-2 pt-2">
				<Badge variant="outline">{controller.sourceEntityCount} source entities</Badge>
				<Badge variant="outline">{controller.sourceRelationshipCount} source relationships</Badge>
				<Badge variant="outline">
					{controller.unresolvedRelationshipCount} unresolved relationships
				</Badge>
				<Badge variant="outline">enumeration: {controller.graph.enumeration.stopReason}</Badge>
			</div>
		</Collapsible.Content>
	</Collapsible.Root>
</div>
