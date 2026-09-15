<script lang="ts">
	import * as Empty from "$components/ui/empty";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { cn } from "$lib/utils";
	import { useSituationController } from "./controller.svelte";
	import { timestamp } from "./model";
	import type { SituationInvestigation } from "@rezible/api-client-ts";

	const controller = useSituationController();

	const query = $derived(controller.investigationsQuery.query);
</script>

<aside class="flex min-h-0 min-w-0 flex-col gap-3" aria-label="Existing investigations">
	<h2 class="text-lg font-semibold">Existing investigations</h2>
	<LoadingQueryWrapper {query}>
		{#snippet view(investigations: SituationInvestigation[])}
			<PaginatedQueryListBox {...controller.investigationsQuery}>
				{#each investigations as inv (inv.id)}
					{@const updatedAt = timestamp(inv.attributes.updatedAt)}
					{@const isSelected = controller.selectedInvestigationId = inv.id}
					<a
						href={controller.investigationSelectionHref(inv.id)}
						data-sveltekit-noscroll
						aria-current={isSelected ? "true" : undefined}
						class={cn(
							"flex min-w-0 flex-col gap-2 rounded-md border p-3 text-sm hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring",
							isSelected ? "border-l-2 border-l-brand bg-selection text-selection-foreground" : "bg-card"
						)}
					>
						<span class="font-medium wrap-anywhere"
							>{inv.attributes.query || "Investigation"}</span
						>
						<time
							class="text-xs text-muted-foreground tabular-nums"
							datetime={updatedAt.iso}
						>
							Updated {updatedAt.label}
						</time>
					</a>
				{:else}
					<Empty.Root>
						<Empty.Header><Empty.Title>No investigations</Empty.Title></Empty.Header>
					</Empty.Root>
				{/each}
			</PaginatedQueryListBox>
		{/snippet}
	</LoadingQueryWrapper>
</aside>
