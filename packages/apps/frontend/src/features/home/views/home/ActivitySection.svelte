<script lang="ts">
	import { useHomeController } from "./controller.svelte";
	import DisplayTime from "./DisplayTime.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { Skeleton } from "$components/ui/skeleton";
	import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import RiInboxLine from "remixicon-svelte/icons/inbox-line";
	import { activityHref } from "./model";

	const controller = useHomeController();
	const query = $derived(controller.activityQuery);

	const icons = {
		"incident-update": RiAlarmWarningLine,
		"situation-investigation": RiSearchLine,
		"inbox-item": RiInboxLine,
	};
</script>

<aside
	aria-labelledby="home-activity"
	class="@container/activity min-w-0 border-t border-border pt-6 @min-[1310px]/dashboard:border-t-0 @min-[1310px]/dashboard:border-l @min-[1310px]/dashboard:pt-0 @min-[1310px]/dashboard:pl-6"
>
	<header class="mb-6 flex items-center justify-between gap-3">
		<h2 id="home-activity" class="text-2xl font-semibold tracking-tight">Recent activity</h2>
		<a 
			class="shrink-0 text-sm text-primary hover:underline focus-visible:outline-ring" 
			href={"/"}
		>
			View all
		</a>
	</header>
	<LoadingQueryWrapper {query} feedbackOnly>
		{#snippet loading()}
			<div class="space-y-6" aria-label="Loading activity">
				{#each [1, 2, 3, 4] as row (row)}
					<div class="flex gap-4">
						<Skeleton class="h-3 w-10" />
						<Skeleton class="size-5" />
						<div class="flex-1 space-y-2">
							<Skeleton class="h-4 w-3/4" />
							<Skeleton class="h-3 w-full" />
						</div>
					</div>
				{/each}
			</div>
		{/snippet}
	</LoadingQueryWrapper>

	{#if query.data}
		<div class="space-y-8">
			{#each controller.activityGroups as group (group.label)}
				<section aria-label={group.label}>
					<h3 class="mb-4 text-sm font-semibold">{group.label}</h3>
					<ol>
						{#each group.items as item (item.id)}
							{@const Icon = icons[item.attributes.recordKind]}
							<li
								class="relative grid grid-cols-[40px_8px_20px_minmax(0,1fr)] gap-3 pb-6 last:pb-0 before:absolute before:top-2 before:-bottom-2 before:left-[55.5px] before:w-px before:bg-border before:content-[''] last:before:hidden"
							>
								<DisplayTime value={item.attributes.occurredAt} clock />
								<span
									class="mt-1.5 size-2 rounded-full bg-muted-foreground opacity-45"
									aria-hidden="true"
								>
									
								</span>
								<Icon class="size-5 shrink-0" aria-hidden="true" />
								<div class="relative min-w-0 wrap-anywhere @min-[440px]/activity:pr-22">
									<a
										class="text-sm font-medium hover:underline focus-visible:outline-ring"
										href={activityHref(item)}
									>
										{item.attributes.title}
									</a>
									<p class="mt-1 text-sm text-muted-foreground">
										{item.attributes.explanation}
									</p>
									<div
										class="@min-[440px]/activity:absolute @min-[440px]/activity:top-0 @min-[440px]/activity:right-0"
									>
										<DisplayTime value={item.attributes.occurredAt} />
									</div>
								</div>
							</li>
						{/each}
					</ol>
				</section>
			{:else}
				<p class="py-6 text-sm text-muted-foreground">No recent activity.</p>
			{/each}
		</div>
	{/if}
</aside>
