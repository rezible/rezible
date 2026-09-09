<script lang="ts">
	import { useHomeController } from "./controller.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import Filters from "$features/situations/components/situation-filters/SituationFilters.svelte";
	import Row from "$features/situations/components/situation-row/SituationRow.svelte";

	const controller = useHomeController();
	const query = $derived(controller.situationsQuery);
</script>

<section aria-labelledby="home-situations" class="min-w-0 rounded-lg border border-border bg-card">
	<header class="flex items-center justify-between gap-2 p-3">
		<h2 id="home-situations" class="font-semibold">
			Situations
			<span class="ml-2 text-sm font-normal text-muted-foreground">
				{query.data?.pagination.total ?? "—"}
			</span>
		</h2>
		<a class="text-sm text-primary hover:underline" href={controller.situationsHref}>View all</a>
	</header>
	<div class="space-y-2 border-b border-border px-3 pb-3">
		<Filters filters={controller.situations} onchange={controller.setSituations} />
		<Button variant="ghost" size="sm" onclick={controller.resetSituations}>Reset filters</Button>
	</div>
	<LoadingQueryWrapper {query} feedbackOnly />
	{#if query.data}
		<div data-preview="situations">
			{#each query.data.data as item (item.id)}
				<Row situation={item} />
				{:else}
					<p class="p-6 text-center text-sm text-muted-foreground">No matching situations.</p>
			{/each}
		</div>
	{/if}
</section>
