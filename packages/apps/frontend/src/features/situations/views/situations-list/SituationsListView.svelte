<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initSituationsListController } from "./controller.svelte";
	import Filters from "$features/situations/components/situation-filters/SituationFilters.svelte";
	import Row from "$features/situations/components/situation-row/SituationRow.svelte";
	import { situationTabs } from "$features/situations/lib/filters";
	import * as Tabs from "$components/ui/tabs";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const controller = initSituationsListController();

	registerPageDescriptor(() => ({ title: "Situations" }));
</script>

<Tabs.Root
	value={controller.filters.tab}
	onValueChange={controller.setTab}
	class="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4"
>
	<div class="flex flex-wrap items-end justify-between gap-2">
		<Tabs.List aria-label="Situation status">
			{#each situationTabs as tab (tab.value)}
				<Tabs.Trigger value={tab.value}>
					{tab.label}
					{#if tab.value === "watching" && controller.watchingCount !== undefined}
						<span class="tabular-nums">({controller.watchingCount})</span>
					{/if}
				</Tabs.Trigger>
			{/each}
		</Tabs.List>
		<Filters filters={controller.filters} onchange={controller.setFilters} />
	</div>

	<Tabs.Content value={controller.filters.tab} class="flex min-h-0 flex-1 flex-col gap-3">
		<PaginatedQueryListBox {...controller.paginatedSituationsQuery}>
			<LoadingQueryWrapper query={controller.query} isEmpty={(situations) => situations.length === 0}>
				{#snippet loading()}
					<ListSkeleton label="Loading situations" />
				{/snippet}
				{#snippet empty()}
					<p class="p-6 text-sm text-muted-foreground">{controller.activeTab.emptyMessage}</p>
				{/snippet}
				{#snippet view(situations)}
					<ul class="divide-y rounded-lg border bg-card">
						{#each situations as situation (situation.id)}
							<Row {situation} now={controller.now} />
						{/each}
					</ul>
				{/snippet}
			</LoadingQueryWrapper>
		</PaginatedQueryListBox>
	</Tabs.Content>
</Tabs.Root>
