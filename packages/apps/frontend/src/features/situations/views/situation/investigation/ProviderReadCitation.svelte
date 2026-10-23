<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import type { ProviderReadView } from "$features/situations/lib/investigation-outputs";

	type Props = { number: number; read: ProviderReadView };
	const { number, read }: Props = $props();
</script>

<div class="flex min-w-0 flex-col gap-0.5 text-xs">
	<div class="flex flex-wrap items-center gap-x-1.5 gap-y-0.5">
		<span class="tabular-nums text-muted-foreground">[{number}]</span>
		<span class="font-medium">{read.label}</span>
		<span>{read.serviceName}</span>
		{#if read.integration}
			<span class="text-muted-foreground">{read.integration}</span>
		{/if}
		<span class="tabular-nums text-muted-foreground">{read.range}</span>
		{#if read.failed}
			<Badge variant="outline">Failed</Badge>
		{/if}
		{#if read.sourceUrl}
			<a
				href={read.sourceUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="text-primary underline-offset-2 hover:underline"
			>
				{read.sourceLabel}
			</a>
		{/if}
	</div>
	{#if read.query}
		<code class="truncate text-muted-foreground" title={read.query}>{read.query}</code>
	{/if}
	{#each read.limits as limit (limit)}
		<span class="text-muted-foreground">{limit}</span>
	{/each}
	{#if read.highlight.length > 0}
		<ol class="mt-1 flex flex-col gap-1 rounded-md bg-muted/50 p-2 font-mono">
			{#each read.highlight as item (item.key)}
				<li class="flex flex-col">
					<div class="flex flex-wrap items-baseline gap-x-2">
						<Timestamp value={item.time} format="clock" class="text-muted-foreground" />
						{#if item.repeats}
							<span class="text-muted-foreground">{item.repeats}</span>
						{/if}
						<span class="whitespace-pre-wrap wrap-anywhere">{item.content}</span>
					</div>
					{#if item.labels.length > 0}
						<span class="text-muted-foreground wrap-anywhere">{item.labels.join(" ")}</span>
					{/if}
				</li>
			{/each}
		</ol>
	{/if}
</div>
