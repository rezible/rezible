<script lang="ts">
	import type { Snippet } from "svelte";
	import type { StatusPresentation } from "$components/common/status-badge/status";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";

	type Props = {
		title: string;
		/** The whole row links here. */
		href?: string;
		/** One muted line under the title. Replaced by `body` when given. */
		summary?: string;
		body?: Snippet;
		/** Leads the facts line, as an icon and toned label. */
		status?: StatusPresentation;
		/** Human reference such as INC-3; follows the status. */
		reference?: string;
		/** Further facts after the reference: times, counts, owners. */
		facts?: Snippet;
		/** Right column: at most one pill or action, then a time. */
		aside?: Snippet;
	};

	const { title, href, summary, body, status, reference, facts, aside }: Props = $props();
</script>

<!-- The title link stretches over the row; controls inside the row are lifted above it. -->
<li
	class="relative grid min-h-[60px] grid-cols-1 gap-x-6 gap-y-2 px-4 py-3 sm:grid-cols-[minmax(0,1fr)_auto] transition-colors duration-150 has-[a[data-row-link]:hover]:bg-muted/50 has-[a[data-row-link]:focus-visible]:outline-2 has-[a[data-row-link]:focus-visible]:-outline-offset-2 has-[a[data-row-link]:focus-visible]:outline-ring [&_:is(button,[data-slot=button],[data-slot=status-badge][tabindex])]:relative"
>
	<div class="flex min-w-0 flex-col gap-1">
		{#if href}
			<a
				data-row-link
				{href}
				class="truncate text-[15px] leading-[22px] font-medium outline-none after:absolute after:inset-0"
			>
				{title}
			</a>
		{:else}
			<p class="truncate text-[15px] leading-[22px] font-medium">{title}</p>
		{/if}

		{#if body}
			{@render body()}
		{:else if summary}
			<p class="line-clamp-1 text-sm text-muted-foreground">{summary}</p>
		{/if}

		{#if status || reference || facts}
			<div class="flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-muted-foreground">
				{#if status}
					<StatusBadge {status} variant="inline" />
				{/if}
				{#if reference}
					{#if status}
						<span aria-hidden="true">·</span>
					{/if}
					<span class="tabular-nums">{reference}</span>
				{/if}
				{#if facts}
					{#if status || reference}
						<span aria-hidden="true">·</span>
					{/if}
					{@render facts()}
				{/if}
			</div>
		{/if}
	</div>

	{#if aside}
		<div
			class="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-xs text-muted-foreground sm:flex-col sm:flex-nowrap sm:items-end"
		>
			{@render aside()}
		</div>
	{/if}
</li>
