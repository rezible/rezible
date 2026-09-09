<script lang="ts">
	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import * as Button from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Separator } from "$components/ui/separator";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import RiCrosshair2Line from "remixicon-svelte/icons/crosshair-2-line";
	import RiExpandDiagonalLine from "remixicon-svelte/icons/expand-diagonal-line";
	import RiHistoryLine from "remixicon-svelte/icons/history-line";
	import RiFileTextLine from "remixicon-svelte/icons/file-text-line";
	import { useSystemMapViewController } from "./controller.svelte";

	const view = useSystemMapViewController();

	const inspected = $derived(view.inspector);
	const entity = $derived(inspected?.kind === "entity" ? inspected.entity : undefined);
	const relationship = $derived(inspected?.kind === "relationship" ? inspected.relationship : undefined);
	const attributes = $derived(entity?.attributes ?? relationship?.attributes);
	const state = $derived(attributes?.latestState);
	const classification = $derived(entity?.attributes.kind ?? relationship?.attributes.predicate);
	const title = $derived.by(() => {
		if (entity) {
			return (
				entity.attributes.latestState?.displayName ||
				entity.attributes.aliases[0]?.attributes.resourceRef.resourceRef ||
				"Unknown subject"
			);
		}
		if (relationship) return relationship.attributes.predicate.replaceAll("_", " ");
		return "Knowledge detail";
	});

	const updatedAt = $derived(attributes?.updatedAt);
	const freshness = $derived.by(() => {
		if (!updatedAt) return undefined;
		const minutes = Math.round((Date.now() - new Date(updatedAt).getTime()) / 60_000);
		if (minutes < 1) return "just now";
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.round(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		return `${Math.round(hours / 24)}d ago`;
	});

	const relationshipTargets = $derived.by(() => {
		if (!relationship) return [];
		return [
			{ role: "source", id: relationship.attributes.sourceEntityId },
			{ role: "target", id: relationship.attributes.targetEntityId },
		];
	});

	const situationHref = $derived.by(() => {
		if (entity?.attributes.kind !== "situation") return undefined;
		const name = entity.attributes.latestState?.displayName;
		return resolve("/situations") + (name ? `?search=${encodeURIComponent(name)}` : "");
	});
</script>

{#if inspected}
	<Card.Root class="absolute top-3 right-3 z-20 w-96 max-w-[calc(100%-1.5rem)] shadow-lg">
		<Card.Header class="gap-2">
			<div class="flex items-start justify-between gap-3">
				<div class="min-w-0">
					<Card.Title class="truncate text-base capitalize">{title}</Card.Title>
					<Card.Description class="flex items-center gap-2">
						<span>{classification?.replaceAll("_", " ")}</span>
						{#if freshness}
							<span class="text-muted-foreground flex items-center gap-1 text-xs">
								<RiHistoryLine class="size-3" />
								updated {freshness}
							</span>
						{/if}
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
		<Card.Content class="max-h-[60vh] space-y-4 overflow-y-auto">
			{#if state?.description}
				<p class="text-muted-foreground text-sm">{state.description}</p>
			{/if}

			{#if entity && view.relationshipSummary.length > 0}
				<div class="space-y-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Relationships</div>
					<ul class="space-y-1 text-sm">
						{#each view.relationshipSummary as [predicate, rels] (predicate)}
							<li class="flex items-center justify-between gap-2">
								<span class="capitalize">{predicate.replaceAll("_", " ")}</span>
								<Badge variant="outline">{rels.length}</Badge>
							</li>
						{/each}
					</ul>
				</div>
			{/if}

			{#if relationship}
				<div class="space-y-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Between</div>
					<ul class="space-y-1 text-sm">
						{#each relationshipTargets as target (target.role)}
							<li class="flex items-center justify-between gap-2">
								<span class="text-muted-foreground capitalize">{target.role}</span>
								<span class="truncate">{view.entityLabel(target.id)}</span>
							</li>
						{/each}
					</ul>
				</div>
			{/if}

			{#if state && Object.keys(state).length}
				<div class="space-y-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Properties</div>
					<dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-sm">
						{#each Object.entries(state.properties ?? {}) as [key, value] (key)}
							<dt class="text-muted-foreground">{key.replaceAll("_", " ")}</dt>
							<dd class="truncate text-right">{String(value)}</dd>
						{/each}
					</dl>
				</div>
			{/if}

			{#if attributes?.aliases?.length}
				<div class="space-y-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">
						Source identities ({attributes.aliases.length})
					</div>
					<div class="flex flex-wrap gap-1">
						{#each attributes.aliases as alias (alias.id)}
							<Badge variant="outline" class="max-w-full">
								<span class="truncate">
									{alias.attributes.resourceRef.provider}
									· {alias.attributes.resourceRef.resourceRef}
								</span>
							</Badge>
						{/each}
					</div>
				</div>
			{/if}

			{#if state && Object.keys(state).length === 0 && !state.description}
				<Separator />
				<p class="text-muted-foreground text-xs">
					No observed state is recorded for this subject yet. Missing detail here is not evidence
					that nothing is wrong.
				</p>
			{/if}
		</Card.Content>
		<Card.Footer class="flex-wrap gap-2">
			{#if situationHref}
				<Button.Root variant="outline" size="sm" href={situationHref}>
					<RiFileTextLine />
					Open in Situations
				</Button.Root>
			{/if}
			{#if entity}
				<Button.Root variant="outline" onclick={() => view.expand(entity)}>
					<RiExpandDiagonalLine />
					Expand
				</Button.Root>
				<Button.Root variant="outline" onclick={() => view.recenter()}>
					<RiCrosshair2Line />
					Recenter
				</Button.Root>
			{/if}
		</Card.Footer>
	</Card.Root>
{/if}
