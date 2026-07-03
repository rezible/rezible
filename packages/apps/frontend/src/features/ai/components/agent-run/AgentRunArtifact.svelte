<script lang="ts">
	import AgentRunPart from "./AgentRunPart.svelte";
	import type { AgentRunArtifactView } from "./controller.svelte";

	type Props = {
		artifact: AgentRunArtifactView;
	};

	const { artifact }: Props = $props();
</script>

<article class="space-y-3 rounded border border-border bg-background p-3">
	<header class="flex min-w-0 items-center justify-between gap-3">
		<h4 class="truncate text-sm font-medium text-foreground">{artifact.name}</h4>
		<span class="text-xs text-muted-foreground">{artifact.parts.length} parts</span>
	</header>

	{#if artifact.metadata}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Metadata</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-muted/30 p-3 text-xs leading-relaxed text-foreground">{artifact.metadata}</pre>
		</div>
	{/if}

	<div class="space-y-2">
		{#each artifact.parts as part (part)}
			<AgentRunPart {part} />
		{:else}
			<p class="text-sm text-muted-foreground">No artifact parts.</p>
		{/each}
	</div>
</article>
