<script lang="ts">
	import type { Event } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { resolve } from "$app/paths";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";
	import EventProjectionDetails from "./EventProjectionDetails.svelte";

	type Props = {
		event: Event;
	};
	const { event }: Props = $props();

	let expanded = $state(false);

	const attrs = $derived(event.attributes);
	const occurredAtLabel = $derived(formatDateTime(attrs.occurredAt));
	const providerLabel = $derived(
		[attrs.resourceRef.provider, attrs.resourceRef.providerNamespace, attrs.providerEventSource]
			.filter(Boolean)
			.join(" / ")
	);
	const projected = $derived(Boolean(attrs.projection));

	function formatDateTime(value: string | undefined) {
		if (!value) return "No timestamp";
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return "No timestamp";
		return new Intl.DateTimeFormat(undefined, {
			month: "short",
			day: "numeric",
			hour: "numeric",
			minute: "2-digit",
		}).format(date);
	}
</script>

<div class="rounded-lg border bg-card text-card-foreground shadow-sm">
	<button
		type="button"
		class="grid w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 text-left transition hover:bg-muted/30"
		aria-expanded={expanded}
		onclick={() => (expanded = !expanded)}
	>
		<span class="flex size-7 shrink-0 items-center justify-center text-muted-foreground">
			{#if expanded}
				<RiArrowDownSLine class="size-4" />
			{:else}
				<RiArrowRightSLine class="size-4" />
			{/if}
		</span>

		<span class="flex min-w-0 flex-col gap-2">
			<span class="flex flex-wrap items-center gap-2">
				<Badge variant="secondary" class="capitalize">{attrs.kind}</Badge>
				{#if projected}
					<Badge variant="outline">Projected</Badge>
				{/if}
			</span>

			<span class="grid min-w-0 gap-1">
				<span class="truncate text-sm font-medium">
					{attrs.resourceRef.resourceRef || event.id}
				</span>
				<span class="truncate text-xs text-muted-foreground">
					{providerLabel || "Unknown provider"}
				</span>
			</span>
		</span>

		<span class="hidden shrink-0 text-xs text-muted-foreground sm:block">{occurredAtLabel}</span>
	</button>

	{#if expanded}
		<div class="border-t px-4 py-3">
			<div class="mb-3 grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
				<div>
					<span class="font-medium text-foreground">Occurred</span>
					<span class="ml-2">{occurredAtLabel}</span>
				</div>
				<div class="min-w-0">
					<span class="font-medium text-foreground">Event ID</span>
					<span class="ml-2 break-all">{event.id}</span>
				</div>
			</div>

			<EventProjectionDetails projection={attrs.projection} />

			<div class="mt-4 rounded-md border border-dashed p-3 text-xs text-muted-foreground">
				Event annotations are not shown yet. Correct annotation display requires the annotations API
				to support filtering by event id.
			</div>

			<div class="mt-3">
				<Button href={resolve("/events/[id]", { id: event.id })} variant="outline" size="sm">
					Open event
				</Button>
			</div>
		</div>
	{/if}
</div>
