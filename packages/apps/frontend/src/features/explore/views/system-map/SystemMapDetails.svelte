<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import * as Button from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import { useSystemMapViewController } from "./controller.svelte";

	const view = useSystemMapViewController();

	const entity = $derived(view.selected?.kind === "entity" ? view.selected.entity : undefined);
	const relationship = $derived(
		view.selected?.kind === "relationship" ? view.selected.relationship : undefined
	);
	const attributes = $derived(entity?.attributes ?? relationship?.attributes);
	const state = $derived(attributes?.latestState);
	const title = $derived(
		state?.displayName ||
			(entity?.attributes.aliases[0]?.attributes.providerSubjectRef ?? relationship?.attributes.kind) ||
			"Knowledge detail"
	);
</script>

{#if view.selected}
	<Card.Root class="absolute top-3 right-3 z-20 w-96 max-w-[calc(100%-1.5rem)] shadow-lg">
		<Card.Header class="gap-2">
			<div class="flex items-start justify-between gap-3">
				<div class="min-w-0">
					<Card.Title class="truncate text-base">{title}</Card.Title>
					<Card.Description>
						{entity ? "Entity" : "Relationship"} · {attributes?.kind.replaceAll("_", " ")}
					</Card.Description>
				</div>
				<Button.Root
					variant="ghost"
					size="icon-sm"
					aria-label="Close details"
					onclick={() => view.clearSelection()}
				>
					<RiCloseLine />
				</Button.Root>
			</div>
		</Card.Header>
		<Card.Content class="space-y-4">
			{#if state?.description}
				<p class="text-muted-foreground text-sm">{state}</p>
			{/if}

			{#if state && Object.keys(state).length}
				<div class="space-y-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Properties</div>
					<dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-sm">
						{#each Object.entries(state) as [key, value] (key)}
							<dt class="text-muted-foreground">{key.replaceAll("_", " ")}</dt>
							<dd class="truncate text-right">{String(value)}</dd>
						{/each}
					</dl>
				</div>
			{/if}

			<div class="space-y-2">
				<div class="text-muted-foreground text-xs font-medium uppercase">Sources</div>
				<div class="flex flex-wrap gap-1">
					{#each attributes?.aliases ?? [] as alias (alias.id)}
						<Badge variant="outline">
							{alias.attributes.provider} · {alias.attributes.providerSource}
						</Badge>
					{/each}
				</div>
			</div>
		</Card.Content>
		<Card.Footer class="gap-2">
			<Button.Root variant="outline" disabled>View</Button.Root>
			<Button.Root variant="outline" disabled>Edit</Button.Root>
			{#if entity}
				<Button.Root variant="secondary" onclick={() => view.expand(entity)}>Expand</Button.Root>
				<Button.Root onclick={() => view.startHere(entity)}>Start here</Button.Root>
			{/if}
		</Card.Footer>
	</Card.Root>
{/if}
