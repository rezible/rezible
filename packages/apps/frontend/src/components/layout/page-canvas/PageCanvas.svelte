<script lang="ts">
	import type { Snippet } from "svelte";
	import { cn } from "$lib/utils";

	type Props = {
		children: Snippet;
		/** Optional supporting column, ~320px, right of the main column when there is room. */
		context?: Snippet;
		/** Required when `context` is provided; used as the <aside> aria-label. */
		contextLabel?: string;
		/** "wide" = 1280px max; "reading" = 880px max, without a context column. */
		width?: "wide" | "reading";
		class?: string;
	};

	const { children, context, contextLabel, width = "wide", class: className }: Props = $props();

	const hasContext = $derived(!!context && width === "wide");
</script>

<div class={cn("@container/canvas min-w-0 flex-1", className)}>
	<div
		class={cn(
			"mx-auto flex w-full flex-col gap-6 p-4 sm:p-6",
			width === "reading" ? "max-w-[880px]" : "max-w-[1280px]",
			hasContext &&
				"@min-[1000px]/canvas:grid @min-[1000px]/canvas:grid-cols-[minmax(0,1fr)_320px] @min-[1000px]/canvas:items-start"
		)}
	>
		<div class="flex min-w-0 flex-col gap-8">
			{@render children()}
		</div>
		{#if hasContext && context}
			<aside
				aria-label={contextLabel}
				class="flex min-w-0 flex-col gap-6 border-t pt-6 @min-[1000px]/canvas:border-t-0 @min-[1000px]/canvas:border-s @min-[1000px]/canvas:pt-0 @min-[1000px]/canvas:ps-6"
			>
				{@render context()}
			</aside>
		{/if}
	</div>
</div>
