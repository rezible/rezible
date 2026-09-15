<script lang="ts">
	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Sheet from "$components/ui/sheet";
	import { Separator } from "$components/ui/separator";
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
		if (relationship) {
			return relationship.attributes.predicate.replaceAll("_", " ");
		}
		return "Knowledge detail";
	});

	const updatedAt = $derived(attributes?.updatedAt);
	const freshness = $derived.by(() => {
		if (!updatedAt) {
			return undefined;
		}
		const minutes = Math.round((Date.now() - new Date(updatedAt).getTime()) / 60_000);
		if (minutes < 1) {
			return "just now";
		}
		if (minutes < 60) {
			return `${minutes}m ago`;
		}
		const hours = Math.round(minutes / 60);
		if (hours < 24) {
			return `${hours}h ago`;
		}
		return `${Math.round(hours / 24)}d ago`;
	});

	const relationshipTargets = $derived.by(() => {
		if (!relationship) {
			return [];
		}
		return [
			{ role: "source", id: relationship.attributes.sourceEntityId },
			{ role: "target", id: relationship.attributes.targetEntityId },
		];
	});

	const situationHref = $derived.by(() => {
		if (entity?.attributes.kind !== "situation") {
			return undefined;
		}
		const name = entity.attributes.latestState?.displayName;
		return resolve("/situations") + (name ? `?search=${encodeURIComponent(name)}` : "");
	});
</script>

<Sheet.Root open={!!inspected} onOpenChange={(open) => !open && view.clearSelection()}>
	<Sheet.Content class="overflow-y-auto sm:max-w-md" onCloseAutoFocus={view.restoreInspectionFocus}>
		<Sheet.Header class="gap-2">
			<div class="flex items-start justify-between gap-3">
				<div class="min-w-0">
					<Sheet.Title>{title}</Sheet.Title>
					<Sheet.Description class="flex items-center gap-2">
						<span>{classification?.replaceAll("_", " ")}</span>
						{#if freshness}
							<span class="text-muted-foreground flex items-center gap-1 text-xs">
								<RiHistoryLine class="size-3" />
								updated {freshness}
							</span>
						{/if}
					</Sheet.Description>
				</div>
			</div>
		</Sheet.Header>
		<div class="flex flex-col gap-4 px-4 pb-4">
			<dl class="flex flex-col gap-1 text-xs">
				<dt class="text-muted-foreground">Identity</dt>
				<dd class="font-mono wrap-anywhere">{entity?.id ?? relationship?.id}</dd>
			</dl>
			{#if state?.description}
				<p class="text-muted-foreground text-sm">{state.description}</p>
			{/if}

			{#if entity && view.relationshipSummary.length > 0}
				<div class="flex flex-col gap-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Relationships</div>
					<ul class="flex flex-col gap-1 text-sm">
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
				<div class="flex flex-col gap-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Between</div>
					<ul class="flex flex-col gap-1 text-sm">
						{#each relationshipTargets as target (target.role)}
							<li class="flex items-center justify-between gap-2">
								<span class="text-muted-foreground capitalize">{target.role}</span>
								<span class="truncate">{view.entityLabel(target.id)}</span>
							</li>
						{/each}
					</ul>
				</div>
			{/if}

			{#if Object.keys(state?.properties ?? {}).length}
				<div class="flex flex-col gap-2">
					<div class="text-muted-foreground text-xs font-medium uppercase">Properties</div>
					<dl class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-x-3 gap-y-1 text-sm">
						{#each Object.entries(state?.properties ?? {}) as [key, value] (key)}
							<dt class="text-muted-foreground">{key.replaceAll("_", " ")}</dt>
							<dd class="whitespace-pre-wrap wrap-anywhere">
								{typeof value === "object" ? JSON.stringify(value, null, 2) : String(value)}
							</dd>
						{/each}
					</dl>
				</div>
			{/if}

			{#if attributes?.aliases?.length}
				<div class="flex flex-col gap-2">
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

			{#if !state?.description && !Object.keys(state?.properties ?? {}).length}
				<Separator />
				<p class="text-muted-foreground text-xs">
					No observed state is recorded for this subject yet. Missing detail here is not evidence
					that nothing is wrong.
				</p>
			{/if}
		</div>
		<Sheet.Footer class="flex-wrap gap-2">
			{#if situationHref}
				<Button variant="outline" size="sm" href={situationHref}>
					<RiFileTextLine data-icon="inline-start" />
					Open in Situations
				</Button>
			{/if}
			{#if entity}
				<Button variant="outline" onclick={() => view.expand(entity)}>
					<RiExpandDiagonalLine data-icon="inline-start" />
					Expand
				</Button>
				<Button variant="outline" onclick={() => view.recenter()}>
					<RiCrosshair2Line data-icon="inline-start" />
					Recenter
				</Button>
			{/if}
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
