<script lang="ts">
	import { Button } from "$components/ui/button";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import SituationRow from "$features/situations/components/situation-row/SituationRow.svelte";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const situations = $derived(controller.situations);
	const total = $derived(controller.situationsTotal);
</script>

{#snippet headingActions()}
	{#if total !== undefined && total > situations.length}
		<Button variant="link" size="inline" href="/situations">View all {total}</Button>
	{/if}
{/snippet}

<section aria-labelledby="home-situations" class="flex min-w-0 flex-col gap-3">
	<SectionHeading id="home-situations" title="Open situations" count={total} actions={headingActions} />
	<LoadingQueryWrapper
		query={controller.situationsQuery}
		feedback="quiet"
		isEmpty={(items) => items.length === 0}
	>
		{#snippet loading()}
			<ListSkeleton label="Loading open situations" />
		{/snippet}
		{#snippet error()}
			<span>Could not load open situations.</span>
		{/snippet}
		{#snippet empty()}
			<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No open situations.</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y overflow-hidden rounded-lg border bg-card">
				{#each situations as situation (situation.id)}
					<SituationRow {situation} conclusion={controller.conclusionFor(situation)} />
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
