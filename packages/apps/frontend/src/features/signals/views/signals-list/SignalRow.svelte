<script lang="ts">
	import { resolve } from "$app/paths";
	import type { AlertDefinition } from "$lib/api";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import { attentionStatus } from "$features/signals/lib/attention";

	type Props = {
		definition: AlertDefinition;
	};

	const { definition }: Props = $props();

	const attrs = $derived(definition.attributes);
	const attention = $derived(attentionStatus(attrs.situationSignalAttention));
	const rosterName = $derived(attrs.roster?.attributes?.name);
	const groupLabels = $derived(attrs.identityGroupLabels.join(", "));
</script>

<a
	href={resolve("/signals/[id]/[[view=signalView]]", { id: definition.id })}
	class="block min-w-0 space-y-1 px-3 py-3 hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-ring"
>
	<div class="flex flex-wrap items-center gap-2">
		<span class="min-w-0 text-sm font-medium wrap-anywhere">{attrs.title}</span>
		{#if attention}
			<StatusBadge status={attention} variant="inline" />
		{/if}
	</div>
	{#if attrs.description}
		<p class="line-clamp-1 text-sm text-muted-foreground">{attrs.description}</p>
	{/if}
	{#if rosterName || groupLabels}
		<div class="flex flex-wrap gap-x-3 text-xs text-muted-foreground">
			{#if rosterName}
				<span>Roster: {rosterName}</span>
			{/if}
			{#if groupLabels}
				<span>Grouped by {groupLabels}</span>
			{/if}
		</div>
	{/if}
</a>
