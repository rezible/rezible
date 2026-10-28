<script lang="ts">
	import RiArrowRightUpLine from "remixicon-svelte/icons/arrow-right-up-line";
	import { Button } from "$components/ui/button";
	import * as Popover from "$components/ui/popover";
	import { Separator } from "$components/ui/separator";
	import { categoryIcon } from "$features/system/lib/system-map/category-icon";
	import { useSystemMapController } from "../controller.svelte";
	import type { CardNodeData, CardPlacement } from "../projection";

	type Props = {
		cardId: string;
		cardSelected: boolean;
		placement: CardNodeData["placement"];
	};
	const { cardId, cardSelected, placement }: Props = $props();

	const controller = useSystemMapController();
	/** The chip whose popover is open: a target id, or "collapsed" for the "+N" chip. */
	let openChip = $state<string>();

	function setChipOpen(chip: string, open: boolean) {
		if (open) {
			openChip = chip;
		} else if (openChip === chip) {
			openChip = undefined;
		}
	}

	/** The chip belongs to the card, so it selects the card rather than letting the node click do so too. */
	function chipClicked(event: MouseEvent) {
		event.stopPropagation();
		if (!cardSelected) {
			controller.selectNode(cardId, event);
		}
	}

	function selectTarget(target: CardPlacement) {
		openChip = undefined;
		controller.selectEntity(target.entity.id);
	}

	function exploreDeployment(target: CardPlacement) {
		openChip = undefined;
		controller.exploreDeployment(target.entity.id);
	}
</script>

{#snippet targetSummary(target: CardPlacement)}
	{@const TargetIcon = categoryIcon(target.entity.category)}
	<div class="flex flex-col gap-2">
		<div class="flex min-w-0 items-start gap-2">
			<span class="bg-muted flex size-8 shrink-0 items-center justify-center rounded-md">
				<TargetIcon class="size-4" aria-hidden="true" />
			</span>
			<div class="flex min-w-0 flex-col">
				<span class="truncate font-medium" title={target.label}>{target.label}</span>
				<span class="text-muted-foreground truncate text-xs">{target.entity.kind}</span>
			</div>
		</div>
		<p class="text-muted-foreground text-xs">{target.countLine}</p>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" size="sm" onclick={() => selectTarget(target)}>Select</Button>
			{#if controller.canExploreDeployment}
				<Button size="sm" onclick={() => exploreDeployment(target)}>
					Explore deployment
					<RiArrowRightUpLine data-icon="inline-end" aria-hidden="true" />
				</Button>
			{/if}
		</div>
	</div>
{/snippet}

<!-- A target chip shows its icon and shrinks to fit; the "+N" chip keeps its width. -->
{#snippet chip(id: string, label: string, title: string, targets: readonly CardPlacement[])}
	{@const ChipIcon = id === "collapsed" ? undefined : categoryIcon(targets[0].entity.category)}
	<Popover.Root open={openChip === id} onOpenChange={(open) => setChipOpen(id, open)}>
		<Popover.Trigger onclick={chipClicked}>
			{#snippet child({ props })}
				<button
					{...props}
					type="button"
					class="placement-chip nodrag nopan bg-muted text-muted-foreground hover:text-foreground flex h-5 max-w-full min-w-0 items-center gap-1 rounded-full px-2 text-[11px] leading-4 transition-colors duration-150"
					class:shrink-0={id === "collapsed"}
					{title}
				>
					{#if ChipIcon}
						<ChipIcon class="size-3 shrink-0" aria-hidden="true" />
					{/if}
					<span class="truncate">{label}</span>
				</button>
			{/snippet}
		</Popover.Trigger>
		<Popover.Content side="right" align="start" class="w-72">
			{#each targets as target, index (target.entity.id)}
				{#if index > 0}
					<Separator />
				{/if}
				{@render targetSummary(target)}
			{/each}
		</Popover.Content>
	</Popover.Root>
{/snippet}

<!-- One chip per row; the "+N" chip shares the last row, so the card's height stays fixed. -->
<div class="flex max-w-full min-w-0 flex-col items-center gap-1" aria-label="Runs on" role="group">
	{#each placement.shown as target, index (target.entity.id)}
		<div class="flex max-w-full min-w-0 items-center gap-1">
			{@render chip(target.entity.id, target.label, `Runs on ${target.label}`, [target])}
			{#if index === placement.shown.length - 1 && placement.collapsed.length > 0}
				{@const count = placement.collapsed.length}
				{@render chip("collapsed", `+${count}`, `Runs on ${count} more`, placement.targets)}
			{/if}
		</div>
	{/each}
</div>

<style>
	.placement-chip:focus-visible {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}
</style>
