<script lang="ts">
	import { Badge } from "$components/ui/badge";

	import AgentRunPart from "./AgentRunPart.svelte";
	import type { AgentRunMessageView } from "./controller.svelte";

	type Props = {
		message: AgentRunMessageView;
	};

	const { message }: Props = $props();
</script>

<article class="space-y-3 rounded border border-border bg-background p-3">
	<header class="flex min-w-0 items-center justify-between gap-3">
		<Badge variant="outline" class="capitalize">{message.role}</Badge>
		<span class="text-xs text-muted-foreground">{message.parts.length} parts</span>
	</header>

	{#if message.metadata}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Metadata</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-muted/30 p-3 text-xs leading-relaxed text-foreground">{message.metadata}</pre>
		</div>
	{/if}

	<div class="space-y-2">
		{#each message.parts as part (part)}
			<AgentRunPart {part} />
		{:else}
			<p class="text-sm text-muted-foreground">No content parts.</p>
		{/each}
	</div>
</article>
