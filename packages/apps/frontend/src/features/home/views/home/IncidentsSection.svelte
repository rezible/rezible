<script lang="ts">
	import { Button } from "$components/ui/button";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { cn } from "$lib/utils";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const rows = $derived(controller.incidentRows);
	const total = $derived(controller.incidentsTotal);

	const accentClasses = {
		danger: "inset-shadow-[3px_0_var(--color-status-danger-foreground)]",
		warning: "inset-shadow-[3px_0_var(--color-status-warning-foreground)]",
		none: "",
	};
</script>

{#snippet headingActions()}
	{#if total !== undefined && total > rows.length}
		<Button variant="link" size="inline" href="/incidents?status=active">View all {total}</Button>
	{/if}
{/snippet}

<section aria-labelledby="home-incidents" class="flex min-w-0 flex-col gap-3">
	<SectionHeading id="home-incidents" title="Active incidents" count={total} actions={headingActions} />
	<LoadingQueryWrapper
		query={controller.incidentsQuery}
		feedback="quiet"
		isEmpty={(items) => items.length === 0}
	>
		{#snippet loading()}
			<ListSkeleton label="Loading active incidents" />
		{/snippet}
		{#snippet error()}
			<span>Could not load active incidents.</span>
		{/snippet}
		{#snippet empty()}
			<p class="text-sm text-muted-foreground">No active incidents.</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y overflow-hidden rounded-lg border bg-card">
				{#each rows as row (row.id)}
					<li class={cn("flex min-h-[60px] gap-4 px-4 py-3", accentClasses[row.accent])}>
						<div class="flex min-w-0 flex-1 flex-col gap-1">
							<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
								<a
									class="rounded-sm text-[15px] leading-[22px] font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring"
									href={row.href}
								>
									{row.title}
								</a>
								<StatusBadge status={row.severity} />
								<StatusBadge status={row.response} variant="inline" />
							</div>
							{#if row.summary}
								<p class="line-clamp-1 text-sm text-muted-foreground">{row.summary}</p>
							{/if}
							<p class="flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground">
								<span>Opened</span>
								<Timestamp value={row.openedAt} />
								{#if row.openFor}
									<span>· {row.openFor}</span>
								{/if}
								{#if row.services.length}
									<span>
										· {row.services.join(
											", "
										)}{#if row.moreServices > 0}&nbsp;+{row.moreServices}{/if}
									</span>
								{/if}
							</p>
						</div>
						<p class="shrink-0 text-xs text-muted-foreground max-sm:hidden">
							Updated
							<Timestamp value={row.updatedAt} />
						</p>
					</li>
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
