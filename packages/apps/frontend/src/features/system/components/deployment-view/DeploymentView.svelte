<script lang="ts">
	import { onDestroy, tick } from "svelte";
	import { watch } from "runed";
	import RiArrowRightUpLine from "remixicon-svelte/icons/arrow-right-up-line";
	import { Button } from "$components/ui/button";
	import * as Empty from "$components/ui/empty";
	import { categoryIcon } from "$features/system/lib/system-map/category-icon";
	import {
		deploymentGroups,
		type DeploymentEntry,
		type DeploymentRow,
	} from "$features/system/lib/system-map/deployment";
	import { cn } from "$lib/utils";
	import type { DeploymentViewProps } from "./types";

	const HIGHLIGHT_MS = 1600;

	const { graph, selection, onSelectionChange, focusTargetId, onShowInStructure }: DeploymentViewProps =
		$props();

	const groups = $derived(deploymentGroups(graph));
	const isEmpty = $derived(groups.sections.length === 0 && groups.placementUnknown.length === 0);
	const selectedEntityId = $derived(selection?.kind === "entity" ? selection.entityId : undefined);

	let container = $state<HTMLElement>();
	let highlightedSectionId = $state<string>();
	let highlightTimer: ReturnType<typeof setTimeout> | undefined;

	watch(
		() => focusTargetId,
		(targetId) => {
			if (targetId) {
				void tick().then(() => focusSection(targetId));
			}
		}
	);

	onDestroy(() => clearTimeout(highlightTimer));

	/** Scrolls to a target's section and highlights it briefly. */
	function focusSection(targetId: string) {
		const section = container?.querySelector<HTMLElement>(`[data-section-id="${CSS.escape(targetId)}"]`);
		if (!section) {
			return;
		}

		const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
		section.scrollIntoView({ block: "nearest", behavior: reducedMotion ? "auto" : "smooth" });
		highlightedSectionId = targetId;
		clearTimeout(highlightTimer);
		highlightTimer = setTimeout(() => {
			highlightedSectionId = undefined;
		}, HIGHLIGHT_MS);
	}

	function selectEntity(entityId: string, event: MouseEvent) {
		onSelectionChange({ kind: "entity", entityId }, event.currentTarget as HTMLElement);
	}

	function showSelectionInStructure() {
		if (selectedEntityId) {
			onShowInStructure(selectedEntityId);
		}
	}
</script>

{#snippet entityText(entry: DeploymentEntry, nameClass: string)}
	{@const EntryIcon = categoryIcon(entry.entity.category)}
	{@const ancestry = entry.breadcrumb.join(" / ")}
	<span class="bg-muted flex size-8 shrink-0 items-center justify-center rounded-md">
		<EntryIcon class="tier-icon size-4" aria-hidden="true" />
	</span>
	<span class="flex min-w-0 flex-col">
		<span class={cn("truncate", nameClass)} title={entry.label}>{entry.label}</span>
		<span class="text-muted-foreground truncate text-xs" title={ancestry || entry.entity.kind}>
			{entry.entity.kind}
			{#if ancestry}
				<span aria-hidden="true">·</span>
				{ancestry}
			{/if}
		</span>
	</span>
{/snippet}

{#snippet row(entry: DeploymentEntry, alsoOn: DeploymentRow["alsoOn"])}
	{@const selected = entry.entity.id === selectedEntityId}
	<li class={cn("deployment-row border-border flex items-center gap-2 border-t", selected && "selected")}>
		<button
			type="button"
			class="flex min-w-0 flex-1 items-center gap-3 px-3 py-2 text-left"
			aria-pressed={selected}
			onclick={(event) => selectEntity(entry.entity.id, event)}
		>
			{@render entityText(entry, "text-sm font-medium")}
		</button>
		{#if alsoOn.length > 0}
			<div class="flex shrink-0 flex-col items-end gap-1 pr-3">
				{#each alsoOn as other (other.entity.id)}
					<Button
						variant="secondary"
						size="xs"
						class="h-6 rounded-full"
						onclick={() => focusSection(other.entity.id)}
					>
						Also on {other.label}
						<RiArrowRightUpLine data-icon="inline-end" aria-hidden="true" />
					</Button>
				{/each}
			</div>
		{/if}
	</li>
{/snippet}

<div bind:this={container} class="h-full min-h-0 overflow-y-auto" aria-label="Deployment">
	<div class="flex flex-col gap-6 p-6">
		<header class="flex flex-wrap items-start justify-between gap-3">
			<div>
				<h2 class="text-2xl leading-9 font-semibold">Deployment</h2>
				<p class="text-muted-foreground text-sm">Grouped by placement target</p>
			</div>
			<Button
				variant="outline"
				size="sm"
				disabled={!selectedEntityId}
				onclick={showSelectionInStructure}
			>
				Show in structure
			</Button>
		</header>

		{#if isEmpty}
			<Empty.Root>
				<Empty.Header>
					<Empty.Description>No placement is known for the loaded entities.</Empty.Description>
				</Empty.Header>
			</Empty.Root>
		{:else}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(22rem,1fr))] items-start gap-4">
				{#each groups.sections as section (section.entity.id)}
					{@const targetSelected = section.entity.id === selectedEntityId}
					<section
						class={cn(
							"deployment-section bg-card border-border overflow-hidden rounded-md border",
							highlightedSectionId === section.entity.id && "highlighted"
						)}
						data-section-id={section.entity.id}
						aria-label={section.label}
					>
						<button
							type="button"
							class={cn(
								"section-heading flex w-full items-center gap-3 p-3 text-left",
								targetSelected && "selected"
							)}
							aria-pressed={targetSelected}
							onclick={(event) => selectEntity(section.entity.id, event)}
						>
							{@render entityText(section, "text-[15px] leading-[22px] font-semibold")}
							<span class="text-muted-foreground ml-auto shrink-0 text-xs tabular-nums">
								{section.countLine}
							</span>
						</button>
						<ul>
							{#each section.rows as entry (entry.entity.id)}
								{@render row(entry, entry.alsoOn)}
							{/each}
						</ul>
					</section>
				{/each}

				{#if groups.placementUnknown.length > 0}
					<section
						class="deployment-section bg-card border-border overflow-hidden rounded-md border"
						aria-label="Placement unknown"
					>
						<h3 class="p-3 text-[15px] leading-[22px] font-semibold">Placement unknown</h3>
						<ul>
							{#each groups.placementUnknown as entry (entry.entity.id)}
								{@render row(entry, [])}
							{/each}
						</ul>
					</section>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.deployment-section {
		transition:
			border-color 160ms ease,
			box-shadow 160ms ease;
	}

	/* Briefly marks the section an Explore deployment or "Also on" link moved to. */
	.deployment-section.highlighted {
		border-color: var(--primary);
		box-shadow: 0 0 0 1px var(--primary);
	}

	.deployment-section :global(.tier-icon) {
		color: var(--chart-2);
	}

	.section-heading:hover,
	.deployment-row:hover {
		background: var(--accent);
	}

	.section-heading.selected,
	.deployment-row.selected {
		background: var(--selection);
		color: var(--selection-foreground);
		box-shadow: inset 2px 0 0 var(--primary);
	}

	.section-heading:focus-visible,
	.deployment-row button:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: -2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.deployment-section {
			transition: none;
		}
	}
</style>
