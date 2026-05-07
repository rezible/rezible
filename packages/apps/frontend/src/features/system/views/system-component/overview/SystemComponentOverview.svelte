<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Badge } from "$components/ui/badge";
	import * as Table from "$components/ui/table";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { listSystemComponentRelationshipsOptions, type SystemComponentRelationship } from "$lib/api";
	import { useSystemComponentViewController } from "../controller.svelte";

	type Props = {
		id?: string;
	};

	const { id }: Props = $props();
	const controller = useSystemComponentViewController();
	const componentId = $derived(id ?? controller.componentId);
	const relationshipsQuery = createQuery(() =>
		listSystemComponentRelationshipsOptions({ query: { componentId } }),
	);

	const component = $derived(controller.component);
	const description = $derived(component?.attributes.description?.trim());
	const kindLabel = $derived(component?.attributes.kind?.attributes.label || "Unclassified");
	const linkedRepositoryRef = $derived(component?.attributes.linkedRepositoryRef);

	function endpointName(relationship: SystemComponentRelationship, side: "source" | "target") {
		const endpoint = side === "source" ? relationship.attributes.source : relationship.attributes.target;
		return endpoint?.attributes.name || relationship.attributes[side === "source" ? "sourceId" : "targetId"];
	}
</script>

{#if component}
	<div class="flex min-w-0 flex-col gap-6 p-4">
		<section class="space-y-3">
			<div class="flex flex-wrap items-center gap-2">
				<Badge variant="secondary">{kindLabel}</Badge>
				<Badge variant="outline">{component.attributes.relationshipCount} links</Badge>
			</div>
			{#if description}
				<p class="max-w-3xl text-sm leading-6 text-foreground">{description}</p>
			{:else}
				<p class="text-sm text-muted-foreground">No description provided.</p>
			{/if}
			{#if linkedRepositoryRef}
				<div class="grid gap-1 text-sm">
					<span class="font-medium text-foreground">Linked repository</span>
					<span class="text-muted-foreground">{linkedRepositoryRef}</span>
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
							No relationships linked to this component.
						</div>
					{/if}
				{/snippet}
			</LoadingQueryWrapper>
		</section>
	</div>
{:else}
	<div class="p-4 text-sm text-muted-foreground">Loading component...</div>
{/if}
