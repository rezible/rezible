<script lang="ts" module>
	export type TimelineEntry = {
		key: string;
		/** ISO time of the entry. */
		at: string;
		label: string;
		description?: string;
		meta?: string;
	};
</script>

<script lang="ts">
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { cn } from "$lib/utils";

	type Props = {
		entries: TimelineEntry[];
		/** Accessible name for the list. */
		label: string;
		timeFormat?: "relative" | "absolute";
		class?: string;
	};

	const { entries, label, timeFormat = "relative", class: className }: Props = $props();
</script>

<ol aria-label={label} class={cn("flex flex-col", className)}>
	{#each entries as entry (entry.key)}
		<li
			class="relative grid grid-cols-[12px_minmax(0,1fr)] gap-x-3 pb-5 before:absolute before:top-4 before:bottom-0 before:left-[5.5px] before:w-px before:bg-border last:pb-0 last:before:hidden"
		>
			<span class="mt-1.5 size-2 justify-self-center rounded-full bg-border" aria-hidden="true"></span>
			<div class="flex min-w-0 flex-col gap-0.5">
				<span class="text-sm font-medium">{entry.label}</span>
				<Timestamp
					value={entry.at}
					format={timeFormat}
					class="self-start text-xs text-muted-foreground"
				/>
				{#if entry.description}
					<p class="text-sm whitespace-pre-wrap text-muted-foreground">{entry.description}</p>
				{/if}
				{#if entry.meta}
					<span class="text-xs text-muted-foreground">{entry.meta}</span>
				{/if}
			</div>
		</li>
	{/each}
</ol>
