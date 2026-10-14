<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import type { ProviderReadView } from "$features/situations/lib/investigation-outputs";

	type Props = { number: number; read: ProviderReadView };
	const { number, read }: Props = $props();
</script>

<div class="flex min-w-0 flex-col gap-0.5 text-xs">
	<div class="flex flex-wrap items-center gap-x-1.5 gap-y-0.5">
		<span class="tabular-nums text-muted-foreground">[{number}]</span>
		<span class="font-medium">{read.label}</span>
		<span>{read.serviceName}</span>
		<span class="tabular-nums text-muted-foreground">{read.range}</span>
		{#if read.failed}
			<Badge variant="outline">Failed</Badge>
		{/if}
		{#if read.exploreUrl}
			<a
				href={read.exploreUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="text-primary underline-offset-2 hover:underline"
			>
				{read.exploreLabel}
			</a>
		{/if}
	</div>
	{#if read.query}
		<code class="truncate text-muted-foreground" title={read.query}>{read.query}</code>
	{/if}
	{#each read.limits as limit (limit)}
		<span class="text-muted-foreground">{limit}</span>
	{/each}
</div>
