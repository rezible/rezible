<script lang="ts">
	import { Button } from "$components/ui/button";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import type { HomeInboxRow } from "./model";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const rows = $derived(controller.inboxRows);
	const total = $derived(controller.inboxTotal);
</script>

{#snippet headingActions()}
	{#if total !== undefined && total > rows.length}
		<span class="text-xs text-muted-foreground">Showing {rows.length} of {total}</span>
	{/if}
{/snippet}

{#snippet facts(row: HomeInboxRow)}
	{#if row.due}
		<StatusBadge status={row.due} variant="inline" />
	{/if}
	{#if row.due && row.context}
		<span aria-hidden="true">·</span>
	{/if}
	{#if row.context}
		<span class="line-clamp-1">{row.context}</span>
	{/if}
{/snippet}

{#snippet aside(row: HomeInboxRow)}
	{#if row.href && row.actionLabel}
		<Button variant="outline" size="sm" href={row.href}>{row.actionLabel}</Button>
	{/if}
	<Timestamp value={row.occurredAt} />
{/snippet}

<section aria-labelledby="home-needs-you" class="flex min-w-0 flex-col gap-3">
	<SectionHeading id="home-needs-you" title="Needs you" count={total} actions={headingActions} />
	<LoadingQueryWrapper
		query={controller.inboxQuery}
		feedback="quiet"
		isEmpty={(items) => items.length === 0}
	>
		{#snippet loading()}
			<ListSkeleton label="Loading items that need you" />
		{/snippet}
		{#snippet error()}
			<span>Could not load items that need you.</span>
		{/snippet}
		{#snippet empty()}
			<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
				Nothing needs your attention.
			</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y overflow-hidden rounded-lg border bg-card">
				{#each rows as row (row.id)}
					{#snippet rowFacts()}
						{@render facts(row)}
					{/snippet}
					{#snippet rowAside()}
						{@render aside(row)}
					{/snippet}
					<ListRow
						title={row.title}
						href={row.href}
						summary={row.reason}
						status={row.kind}
						facts={row.due || row.context ? rowFacts : undefined}
						aside={rowAside}
					/>
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
