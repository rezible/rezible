<script lang="ts">
	import { useHomeController } from "./controller.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import Filters from "$features/events/components/event-filters/EventFilters.svelte";

	const controller = useHomeController();
	const query = $derived(controller.eventsQuery);
</script>

<section aria-labelledby="home-events" class="min-w-0 rounded-lg border border-border bg-card lg:col-span-2">
	<header class="flex items-center justify-between gap-2 p-3">
		<h2 id="home-events" class="font-semibold">
			Operational events
			<span class="ml-2 text-sm font-normal text-muted-foreground">
				{query.data?.pagination.total ?? "—"}
			</span>
		</h2>
		<a class="text-sm text-primary hover:underline" href={controller.eventsHref}>View all</a>
	</header>
	<div class="space-y-2 border-b border-border px-3 pb-3">
		<Filters filters={controller.events} onchange={controller.setEvents} />
		<Button variant="ghost" size="sm" onclick={controller.resetEvents}>Reset filters</Button>
	</div>
	<LoadingQueryWrapper {query} feedbackOnly />
	{#if query.data}
		<div data-preview="events">
			{#each query.data.data as item (item.id)}
				<div
					class="grid min-w-0 gap-1 border-b border-border px-3 py-3 text-sm last:border-b-0 sm:grid-cols-[minmax(8rem,1fr)_minmax(0,2fr)_minmax(0,1fr)_auto] sm:gap-3"
				>
					<span class="break-words font-medium">{item.attributes.kind}</span>
					<span class="break-all">{item.attributes.resourceRef.resourceRef}</span>
					<span class="break-words text-xs text-muted-foreground">
						{[item.attributes.resourceRef.provider, item.attributes.providerEventSource]
							.filter(Boolean)
							.join(" / ")}
					</span>
					<time class="text-xs text-muted-foreground" datetime={item.attributes.occurredAt}>
						{new Date(item.attributes.occurredAt).toLocaleString()}
					</time>
				</div>
			{:else}
				<p class="p-6 text-center text-sm text-muted-foreground">No matching events.</p>
			{/each}
		</div>
	{/if}
</section>
