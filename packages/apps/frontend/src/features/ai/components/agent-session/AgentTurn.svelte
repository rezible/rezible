<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import type { AgentTurn } from "$lib/api";

	type Props = {
		turn: AgentTurn;
	};

	const { turn }: Props = $props();
	const attrs = $derived(turn.attributes);
	const badgeVariant = $derived(attrs.status === "failed" ? "destructive" : "outline");
</script>

<article class="space-y-3 rounded-md border border-border bg-background p-3">
	<header class="flex flex-wrap items-center justify-between gap-2">
		<div class="flex items-center gap-2">
			<Badge variant={badgeVariant} class="capitalize">{attrs.status}</Badge>
			{#if attrs.finishReason}
				<span class="text-xs text-muted-foreground">{attrs.finishReason}</span>
			{/if}
		</div>
		<time class="text-xs text-muted-foreground" datetime={attrs.updatedAt}>{attrs.updatedAt}</time>
	</header>

	{#if attrs.error}
		<pre class="overflow-auto rounded bg-destructive/10 p-3 text-xs text-destructive">{JSON.stringify(
				attrs.error,
				null,
				2
			)}</pre>
	{/if}

	{#if attrs.knowledgeCitations?.length}
		<div class="space-y-2 rounded bg-muted p-3 text-xs text-foreground">
			{#each attrs.knowledgeCitations as citation (citation.id)}
				<div>
					<div class="font-medium">{citation.attributes.summary}</div>
					<div class="text-muted-foreground">{citation.attributes.knowledgeEvidenceId}</div>
				</div>
			{/each}
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">No knowledge citations.</p>
	{/if}
</article>
