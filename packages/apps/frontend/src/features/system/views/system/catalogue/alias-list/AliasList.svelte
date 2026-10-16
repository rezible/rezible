<script lang="ts">
	import type { AliasView } from "$features/system/lib/catalogue/proposals";

	type Props = { aliases: AliasView[] };
	const { aliases }: Props = $props();
</script>

<ul class="space-y-1 text-xs text-muted-foreground">
	{#each aliases as alias (alias.key)}
		<li>
			<span class="font-medium text-foreground">{alias.integration}</span>
			<span>names it</span>
			<code class="text-foreground">{alias.name}</code>
			<span title={alias.lastRecordedTitle}>· {alias.lastRecorded}</span>
			{#if alias.mergedNote}
				<div>{alias.mergedNote}</div>
			{/if}
		</li>
	{:else}
		<li>No sources</li>
	{/each}
</ul>
