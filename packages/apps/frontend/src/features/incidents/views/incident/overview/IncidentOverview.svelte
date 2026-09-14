<script lang="ts">
	import { initIncidentOverviewController } from "./controller.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { Button } from "$components/ui/button";

	const controller = initIncidentOverviewController();
	const updatesQuery = $derived(controller.paginatedUpdatesQuery.query);
	const paginator = controller.paginatedUpdatesQuery.paginator;
	const updates = $derived(controller.updates);
	const pagination = $derived(controller.updatesQuery.data?.pagination);
</script>

<section class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4" aria-labelledby="incident-updates">
	<h2 id="incident-updates" class="text-lg font-semibold">Incident updates</h2>
	<LoadingQueryWrapper query={updatesQuery} feedbackOnly />

	{#if updates}
		{#each updates as update (update.id)}
			<article
				id={`update-${update.id}`}
				class="space-y-2 rounded-md border border-border bg-card p-4 target:ring-2 target:ring-ring"
			>
				<p class="whitespace-pre-wrap text-sm">{update.attributes.body}</p>
				<time class="text-xs text-muted-foreground" datetime={update.attributes.createdAt}
					>{new Date(update.attributes.createdAt).toLocaleString()}</time
				>
			</article>
		{:else}
			<p class="text-sm text-muted-foreground">No incident updates.</p>
		{/each}

		<nav aria-label="Incident update pages" class="flex items-center justify-between gap-3 text-sm">
			{#if pagination}
				<span>Page {paginator.page} · {pagination.total} updates</span>
				<div class="flex gap-2">
					<Button
						variant="outline"
						size="sm"
						disabled={paginator.page <= 1 || updatesQuery.isFetching}
						onclick={() => paginator.setPage(paginator.page - 1)}>Previous</Button
					>
					<Button
						variant="outline"
						size="sm"
						disabled={paginator.page * paginator.pageSize >= pagination.total ||
							updatesQuery.isFetching}
						onclick={() => paginator.setPage(paginator.page + 1)}>Next</Button
					>
				</div>
			{/if}
		</nav>
	{/if}
</section>
