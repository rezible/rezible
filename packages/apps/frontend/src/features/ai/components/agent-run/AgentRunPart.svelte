<script lang="ts">
	import type { AgentRunPartView } from "./controller.svelte";

	type Props = {
		part: AgentRunPartView;
	};

	const { part }: Props = $props();
</script>

<div class="space-y-3 rounded border border-border bg-muted/30 p-3">
	{#if part.text}
		<p class="whitespace-pre-wrap text-sm leading-relaxed text-foreground">{part.text}</p>
	{/if}

	{#if part.rows.length > 0}
		<dl class="grid gap-2 text-xs sm:grid-cols-2">
			{#each part.rows as row (row.label)}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">{row.label}</dt>
					<dd class="break-words text-foreground">{row.value}</dd>
				</div>
			{/each}
		</dl>
	{/if}

	{#each part.jsonSections as section (section.label)}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">{section.label}</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{section.value}</pre>
		</div>
	{/each}
</div>
