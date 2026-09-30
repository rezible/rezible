<script lang="ts">
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const query = $derived(controller.activityQuery);
</script>

<section aria-labelledby="home-activity" class="flex min-w-0 flex-col gap-4">
	<SectionHeading id="home-activity" title="Recent activity" level={3} />

	<LoadingQueryWrapper {query} feedback="quiet" isEmpty={(items) => items.length === 0}>
		{#snippet loading()}
			<div class="flex flex-col gap-4" aria-label="Loading activity">
				{#each [1, 2, 3] as row (row)}
					<div class="flex gap-3">
						<Skeleton class="h-3 w-10" />
						<div class="flex flex-1 flex-col gap-2">
							<Skeleton class="h-4 w-3/4" />
							<Skeleton class="h-3 w-full" />
						</div>
					</div>
				{/each}
			</div>
		{/snippet}
		{#snippet error()}
			<span>Could not load recent activity.</span>
		{/snippet}
		{#snippet empty()}
			<p class="text-sm text-muted-foreground">No recent activity.</p>
		{/snippet}
		{#snippet view()}
			{#each controller.activityGroups as group (group.label)}
				<section aria-label={group.label} class="flex flex-col gap-3">
					<h4 class="region-label">{group.label}</h4>
					<ol class="flex flex-col gap-4">
						{#each group.rows as row (row.id)}
							<li class="grid grid-cols-[3.5rem_20px_minmax(0,1fr)] gap-3">
								<Timestamp
									value={row.occurredAt}
									format="clock"
									class="text-xs text-muted-foreground"
								/>
								<row.icon class="size-5 text-muted-foreground" aria-hidden="true" />
								<div class="flex min-w-0 flex-col gap-0.5 wrap-anywhere">
									{#if row.href}
										<a
											class="rounded-sm text-sm font-medium hover:underline focus-visible:outline-2 focus-visible:outline-ring"
											href={row.href}
										>
											{row.title}
										</a>
									{:else}
										<p class="text-sm font-medium">{row.title}</p>
									{/if}
									<p class="line-clamp-2 text-sm text-muted-foreground">
										{row.explanation}
									</p>
								</div>
							</li>
						{/each}
					</ol>
				</section>
			{/each}
		{/snippet}
	</LoadingQueryWrapper>
</section>
