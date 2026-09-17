<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Checkbox } from "$components/ui/checkbox";
	import * as Collapsible from "$components/ui/collapsible";
	import * as Field from "$components/ui/field";
	import * as ToggleGroup from "$components/ui/toggle-group";
	import { useSystemMapFixtureController } from "./controller.svelte";

	const controller = useSystemMapFixtureController();
</script>

<div class="bg-card border-border flex shrink-0 flex-col gap-2 rounded-md border p-2">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<ToggleGroup.Root
			type="single"
			value={controller.scenario}
			onValueChange={controller.setScenario}
			aria-label="Fixture scenario"
			variant="outline"
			size="sm"
			class="max-w-full flex-wrap"
		>
			<ToggleGroup.Item value="hierarchy">Nested + shared</ToggleGroup.Item>
			<ToggleGroup.Item value="connections">Direct + summary</ToggleGroup.Item>
			<ToggleGroup.Item value="context">Actors + annotations</ToggleGroup.Item>
			<ToggleGroup.Item value="stress">Long + dense</ToggleGroup.Item>
		</ToggleGroup.Root>

		<Field.FieldSet class="flex-row items-center gap-3 text-sm">
			<Field.FieldLegend class="sr-only">Map context</Field.FieldLegend>
			<label class="flex items-center gap-2">
				<Checkbox
					checked={controller.showActors}
					onCheckedChange={(checked) => controller.setShowActors(!!checked)}
				/>
				Actors
			</label>
			<label class="flex items-center gap-2">
				<Checkbox
					checked={controller.showAnnotations}
					onCheckedChange={(checked) => controller.setShowAnnotations(!!checked)}
				/>
				Annotations
			</label>
		</Field.FieldSet>
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
				<Badge variant="outline">{controller.parentMembershipCoverage} membership coverage</Badge>
				<Badge variant="outline">{controller.relationshipCoverage} relationship coverage</Badge>
			</div>
		</Collapsible.Content>
	</Collapsible.Root>
</div>
