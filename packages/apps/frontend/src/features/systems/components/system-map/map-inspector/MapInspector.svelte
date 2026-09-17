<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Separator } from "$components/ui/separator";
	import { useSystemMapController } from "../controller.svelte";
	import { buildInspectionViewData } from "./presentation";

	type Props = {
		onHide: () => void;
	};
	let { onHide }: Props = $props();
	
	const controller = useSystemMapController();
	const inspectionView = $derived(buildInspectionViewData(controller.inspection));
</script>

<aside
	class="bg-card border-border flex w-80 min-h-0 min-w-0 shrink-0 flex-col border-l"
	aria-label="Map inspection"
>
	<div class="flex shrink-0 items-center justify-between gap-2 p-3">
		<div class="min-w-0">
			<h2 class="text-sm font-semibold">Map contents</h2>
			<p class="text-muted-foreground text-xs">Keyboard-accessible source list</p>
		</div>
		<div class="flex items-center gap-1">
			{#if controller.selectedTarget}
				<Button variant="ghost" size="sm" onclick={controller.clearSelection}>Clear</Button>
			{/if}
			<Button variant="ghost" size="sm" aria-label="Hide map inspector" onclick={onHide}>Hide</Button>
		</div>
	</div>

	{#if controller.inspection && inspectionView}
		<div class="min-h-0 max-h-[50%] shrink-0 overflow-y-auto">
			<section class="flex flex-col gap-3 p-3" aria-label="Selected source inspection">
				<div class="flex items-start justify-between gap-2">
					<div class="min-w-0">
						<h3 class="break-words text-sm font-semibold">{inspectionView.title}</h3>
						<p class="text-muted-foreground text-xs">{inspectionView.visibilityText}</p>
					</div>
					{#if inspectionView.representativeId}
						<Badge variant="outline" class="shrink-0">{inspectionView.representativeId}</Badge>
					{/if}
				</div>

				{#if controller.canReveal}
					<Button variant="outline" size="sm" onclick={controller.revealSelected}>
						Reveal selected subject
					</Button>
				{/if}

				{#if controller.inspection.entity}
					<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1 text-xs">
						<dt class="text-muted-foreground">ID</dt>
						<dd class="wrap-anywhere font-mono">{controller.inspection.entity.id}</dd>
						<dt class="text-muted-foreground">Category</dt>
						<dd class="wrap-anywhere">{controller.inspection.entity.category}</dd>
						<dt class="text-muted-foreground">Kind</dt>
						<dd class="wrap-anywhere">{controller.inspection.entity.kind}</dd>
					</dl>
				{/if}

				{#if inspectionView.summary}
					<div class="flex flex-col gap-2">
						<div class="flex items-center justify-between gap-2">
							<h4 class="text-xs font-medium uppercase">Summary contributors</h4>
							<Badge variant="secondary">{inspectionView.summary.count}</Badge>
						</div>
						<ul class="flex flex-col gap-1 text-xs">
							{#each inspectionView.summary.relationships as contributor (contributor.id)}
								<li>
									<Button
										variant="link"
										size="xs"
										class="h-auto max-w-full justify-start whitespace-normal px-0 text-left"
										onclick={() =>
											controller.selectItem({
												kind: "relationship",
												relationshipId: contributor.id,
											})}
									>
										{contributor.text}
									</Button>
								</li>
							{/each}
						</ul>
						{#if inspectionView.summary.missingRelationshipIds.length > 0}
							<Alert.Root>
								<Alert.Description>
									Unavailable contributor IDs:
									{inspectionView.summary.missingRelationshipIds.join(", ")}
								</Alert.Description>
							</Alert.Root>
						{/if}
					</div>
				{:else if inspectionView.relationship}
					<div class="flex flex-col gap-2">
						<h4 class="text-xs font-medium uppercase">Source relationship</h4>
						<p class="break-words text-xs">{inspectionView.relationship.text}</p>
						<p class="text-muted-foreground wrap-anywhere font-mono text-xs">
							{inspectionView.relationship.id}
						</p>
					</div>
				{/if}

				{#if controller.inspection.annotation}
					<div class="flex flex-col gap-1 text-xs">
						<h4 class="font-medium uppercase">Annotation target</h4>
						<p class="break-words">{inspectionView.annotationTargetLabel}</p>
					</div>
				{/if}

				{#if inspectionView.memberships.length > 0}
					<div class="flex flex-col gap-1 text-xs">
						<h4 class="font-medium uppercase">Membership facts</h4>
						{#each inspectionView.memberships as membership (membership.id)}
							<Button
								variant="link"
								size="xs"
								class="h-auto justify-start whitespace-normal px-0 text-left"
								onclick={() =>
									controller.selectItem({
										kind: "relationship",
										relationshipId: membership.id,
									})}
							>
								{membership.text}
							</Button>
						{/each}
					</div>
				{/if}

				<p class="text-muted-foreground text-xs">{inspectionView.relationshipCountText}</p>
			</section>
		</div>
		<Separator />
	{/if}

	<section class="min-h-0 flex-1 overflow-y-auto" aria-label="Supplied graph subjects">
		<div class="flex flex-col gap-1 p-2">
			{#each controller.inspectionItems as item (item.id)}
				<Button
					variant={controller.selectedItemId === item.id ? "secondary" : "ghost"}
					size="sm"
						class="h-auto min-w-0 justify-start whitespace-normal py-2 text-left"
					aria-current={controller.selectedItemId === item.id ? "true" : undefined}
					onclick={() => controller.selectItem(item.target)}
				>
					<span class="flex min-w-0 flex-col gap-0.5">
						<span class="break-words">{item.label}</span>
						<span class="text-muted-foreground break-words text-xs">{item.detail}</span>
					</span>
				</Button>
			{/each}
		</div>
	</section>
</aside>
