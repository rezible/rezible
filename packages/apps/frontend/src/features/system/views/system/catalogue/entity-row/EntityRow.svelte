<script lang="ts">
	import type { EntityRowView } from "$features/system/lib/catalogue/proposals";
	import { useCatalogueController } from "../controller.svelte";
	import AliasList from "../alias-list/AliasList.svelte";

	type Props = { entity: EntityRowView };
	const { entity }: Props = $props();

	const controller = useCatalogueController();

	const possibleMatchPath = $derived.by(() => {
		if (!entity.possibleMatchProposalId) {
			return undefined;
		}
		return controller.proposalHref(entity.possibleMatchProposalId);
	});
</script>

<li class="grid gap-1 px-4 py-3 sm:grid-cols-[minmax(0,14rem)_minmax(0,1fr)] sm:gap-4">
	<div class="min-w-0">
		<div class="flex flex-wrap items-baseline gap-x-2">
			<span class="font-medium wrap-anywhere">{entity.displayName}</span>
			{#if possibleMatchPath}
				<a
					href={possibleMatchPath}
					class="text-xs text-muted-foreground underline underline-offset-2"
				>
					Possible match
					<span class="sr-only">for {entity.displayName}</span>
				</a>
			{/if}
		</div>
		<div class="flex flex-wrap gap-1 text-xs text-muted-foreground">
			{#each entity.values as value (value)}
				<code>{value}</code>
			{/each}
		</div>
	</div>
	<AliasList aliases={entity.aliases} />
</li>
