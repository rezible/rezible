<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const rows = $derived(controller.situationRows);
	const total = $derived(controller.situationsTotal);
</script>

{#snippet headingActions()}
	{#if total !== undefined && total > rows.length}
		<Button variant="link" size="inline" href="/situations?status=active">View all {total}</Button>
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
			<p class="text-sm text-muted-foreground">No open situations.</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y rounded-lg border bg-card">
				{#each rows as row (row.id)}
					<li class="flex min-h-[60px] flex-col gap-1 px-4 py-3">
						<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
							<a
								class="rounded-sm text-[15px] leading-[22px] font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring"
								href={row.href}
							>
								{row.title}
							</a>
							<StatusBadge status={row.status} variant="inline" />
						</div>
						{#if row.conclusion.kind === "loading"}
							<Skeleton class="h-4 w-3/4" />
						{:else if row.conclusion.kind === "text"}
							<p class="line-clamp-2 text-sm">
								<span class="text-muted-foreground">Latest conclusion ·</span>
								{row.conclusion.text}
							</p>
						{:else if row.summary}
							<p class="line-clamp-1 text-sm text-muted-foreground">{row.summary}</p>
						{/if}
						<p class="flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground">
							<span>Opened</span>
							<Timestamp value={row.openedAt} />
							{#if row.sourceCount > 0}
								<span class="tabular-nums">
									· {row.sourceCount}
									{row.sourceCount === 1 ? "source" : "sources"}
								</span>
							{/if}
							{#if row.linkedIncidentCount > 0}
								<span class="tabular-nums">
									· Linked to {row.linkedIncidentCount}
									{row.linkedIncidentCount === 1 ? "incident" : "incidents"}
								</span>
							{/if}
						</p>
					</li>
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
