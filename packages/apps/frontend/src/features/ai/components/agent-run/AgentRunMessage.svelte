<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import type { Message } from "$lib/api";

	import AgentRunPart from "./AgentRunPart.svelte";

	type Props = {
		message: Message;
	};

	const { message }: Props = $props();

	const parts = $derived(message.content ?? []);
</script>

<article class="space-y-3 rounded border border-border bg-background p-3">
	<header class="flex min-w-0 items-center justify-between gap-3">
		<Badge variant="outline" class="capitalize">{message.role || "unknown"}</Badge>
		<span class="text-xs text-muted-foreground">{parts.length} parts</span>
	</header>

	{#if message.metadata}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Metadata</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-muted/30 p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					message.metadata,
					null,
					2
				)}</pre>
		</div>
	{/if}

	<div class="space-y-2">
		{#each parts as part (part)}
			<AgentRunPart {part} />
		{:else}
			<p class="text-sm text-muted-foreground">No content parts.</p>
		{/each}
	</div>
</article>
