<script lang="ts">
	import type { AiAgentRunSnapshotStateArtifact } from "$lib/api";

	import AgentRunPart from "./AgentRunPart.svelte";

	type Props = {
		artifact: AiAgentRunSnapshotStateArtifact;
		index?: number;
	};

	const { artifact, index = 0 }: Props = $props();

	const name = $derived(artifact.name || `Artifact ${index + 1}`);
	const parts = $derived(artifact.parts);
</script>

<article class="space-y-3 rounded border border-border bg-background p-3">
	<header class="flex min-w-0 items-center justify-between gap-3">
		<h4 class="truncate text-sm font-medium text-foreground">{name}</h4>
		<span class="text-xs text-muted-foreground">{parts.length} parts</span>
	</header>

	{#if artifact.metadata}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Metadata</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-muted/30 p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					artifact.metadata,
					null,
					2
				)}</pre>
		</div>
	{/if}

	<div class="space-y-2">
		{#each parts as part (part)}
			<AgentRunPart {part} />
		{:else}
			<p class="text-sm text-muted-foreground">No artifact parts.</p>
		{/each}
	</div>
</article>
