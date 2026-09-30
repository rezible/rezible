<script lang="ts">
	import { Button } from "$components/ui/button";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
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
			<p class="text-sm text-muted-foreground">Nothing needs your attention.</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y rounded-lg border bg-card">
				{#each rows as row (row.id)}
					<li
						class="grid min-h-[60px] grid-cols-[20px_minmax(0,1fr)] items-start gap-x-3 gap-y-2 px-4 py-3 sm:grid-cols-[20px_minmax(0,1fr)_auto]"
					>
						<row.icon class="mt-0.5 size-5 text-muted-foreground" aria-hidden="true" />
						<div class="flex min-w-0 flex-col gap-0.5 wrap-anywhere">
							{#if row.href}
								<a
									class="rounded-sm text-[15px] leading-[22px] font-medium hover:underline focus-visible:outline-2 focus-visible:outline-ring"
									href={row.href}
								>
									{row.title}
								</a>
							{:else}
								<p class="text-[15px] leading-[22px] font-medium">{row.title}</p>
							{/if}
							<p class="text-sm text-muted-foreground">{row.reason}</p>
							{#if row.context}
								<p class="line-clamp-1 text-xs text-muted-foreground">{row.context}</p>
							{/if}
						</div>
						<div class="col-start-2 flex items-center gap-3 sm:col-start-auto">
							{#if row.due}
								<StatusBadge status={row.due} variant="inline" />
							{/if}
							<Timestamp value={row.occurredAt} class="text-xs text-muted-foreground" />
							{#if row.href && row.actionLabel}
								<Button variant="outline" size="sm" href={row.href}>{row.actionLabel}</Button>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
