<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Badge } from "$components/ui/badge";
	import * as Table from "$components/ui/table";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { listSystemTopologyRelationshipsOptions, type SystemTopologyRelationship } from "$lib/api";
	import { useSystemTopologyEntityViewController } from "../controller.svelte";

	type Props = {
		id?: string;
	};

	const { id }: Props = $props();
	const controller = useSystemTopologyEntityViewController();
	const entityId = $derived(id ?? controller.entityId);
	const relationshipsQuery = createQuery(() =>
		listSystemTopologyRelationshipsOptions({ query: { entityId } }),
	);

	const entity = $derived(controller.entity);
	const description = $derived(entity?.attributes.description?.trim());
	const kindLabel = $derived(entity?.attributes.kind || "unclassified");
	const aliases = $derived(entity?.attributes.aliases ?? []);

	function endpointName(relationship: SystemTopologyRelationship, side: "source" | "target") {
		const endpoint = side === "source" ? relationship.attributes.source : relationship.attributes.target;
		return endpoint?.attributes.displayName || relationship.attributes[side === "source" ? "sourceEntityId" : "targetEntityId"];
	}
</script>

{#if entity}
	<div class="flex min-w-0 flex-col gap-6 p-4">
		<section class="space-y-3">
			<div class="flex flex-wrap items-center gap-2">
				<Badge variant="secondary">{kindLabel}</Badge>
				<Badge variant="outline">{entity.attributes.relationships?.length ?? 0} links</Badge>
			</div>
			{#if description}
				<p class="max-w-3xl text-sm leading-6 text-foreground">{description}</p>
			{:else}
				<p class="text-sm text-muted-foreground">No description provided.</p>
			{/if}
			{#if aliases.length > 0}
				<div class="grid gap-1 text-sm">
					<span class="font-medium text-foreground">Provider aliases</span>
					<div class="flex flex-wrap gap-2">
						{#each aliases as alias (alias.id)}
							<Badge variant="outline">{alias.provider}:{alias.subjectKind}</Badge>
						{/each}
					</div>
				</div>
			{/if}
		</section>

		<section class="space-y-3">
			<h2 class="text-sm font-medium text-foreground">Relationships</h2>
			<LoadingQueryWrapper query={relationshipsQuery}>
				{#snippet view(relationships)}
					{#if relationships.length > 0}
						<div class="overflow-hidden rounded border border-border">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Source</Table.Head>
										<Table.Head>Target</Table.Head>
										<Table.Head>Description</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each relationships as relationship (relationship.id)}
										<Table.Row>
											<Table.Cell>{endpointName(relationship, "source")}</Table.Cell>
											<Table.Cell>{endpointName(relationship, "target")}</Table.Cell>
											<Table.Cell class="text-muted-foreground">
												{relationship.attributes.description || "No description provided."}
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					{:else}
						<div class="rounded border border-dashed border-border p-6 text-sm text-muted-foreground">
							No relationships linked to this entity.
						</div>
					{/if}
				{/snippet}
			</LoadingQueryWrapper>
		</section>
	</div>
{:else}
	<div class="p-4 text-sm text-muted-foreground">Loading entity...</div>
{/if}
