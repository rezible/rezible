<script lang="ts">
	import { resolve } from "$app/paths";
	import type { Situation } from "$lib/api";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import { situationStatus } from "$features/situations/lib/status";

	type Props = {
		situation: Situation;
	};

	let { situation }: Props = $props();

	const attrs = $derived(situation.attributes);
</script>

<a
	href={resolve("/situations/[id]/[[view=situationView]]", { id: situation.id })}
	class="block min-w-0 space-y-1 border-b border-border px-3 py-3 last:border-b-0 hover:bg-muted/50 focus-visible:outline-primary"
>
	<div class="flex flex-wrap items-center gap-2">
		<span class="min-w-0 break-words text-sm font-medium">{attrs.title}</span>
		<StatusBadge status={situationStatus(attrs)} variant="inline" />
	</div>
	{#if attrs.summary}
		<p class="line-clamp-2 text-sm text-muted-foreground">{attrs.summary}</p>
	{/if}
	<div class="flex flex-wrap justify-between gap-x-3 text-xs text-muted-foreground">
		<span>
			{attrs.signalCount} contributing {attrs.signalCount === 1 ? "signal" : "signals"}
		</span>
		<time datetime={attrs.openedAt}>Opened {new Date(attrs.openedAt).toLocaleString()}</time>
	</div>
</a>
