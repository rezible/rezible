<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Table from "$components/ui/table";
	import { cn } from "$lib/utils";
	import { useSystemMapViewController, makeEntityLabel } from "./controller.svelte";

	const view = useSystemMapViewController();

	function formatUpdatedAt(iso: string) {
		const minutes = Math.round((Date.now() - new Date(iso).getTime()) / 60_000);
		if (minutes < 60) {
			return `${minutes}m ago`;
		}

		const hours = Math.round(minutes / 60);
		if (hours < 24) {
			return `${hours}h ago`;
		}

		return `${Math.round(hours / 24)}d ago`;
	}
</script>

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
						<Table.Cell>
							<Button variant="link" onclick={() => view.selectEntity(entity)}>
								{makeEntityLabel(entity)}
							</Button>
						</Table.Cell>
						<Table.Cell>{entity.attributes.kind.replaceAll("_", " ")}</Table.Cell>
						<Table.Cell>
							{view.connectionCount(entity.id)}
						</Table.Cell>
						<Table.Cell>{formatUpdatedAt(entity.attributes.updatedAt)}</Table.Cell>
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
						class={cn(view.inspectedRelationship?.id === relationship.id && "bg-selection")}
					>
						<Table.Cell>
							{view.entityLabel(relationship.attributes.sourceEntityId)}
						</Table.Cell>
						<Table.Cell>
							<Button variant="link" onclick={() => view.selectRelationship(relationship)}>
								{relationship.attributes.predicate.replaceAll("_", " ")}
							</Button>
						</Table.Cell>
						<Table.Cell>
							{view.entityLabel(relationship.attributes.targetEntityId)}
						</Table.Cell>
						<Table.Cell>{formatUpdatedAt(relationship.attributes.updatedAt)}</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</section>
</div>
