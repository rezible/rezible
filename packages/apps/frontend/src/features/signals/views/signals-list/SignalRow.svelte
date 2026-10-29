<script lang="ts">
	import { resolve } from "$app/paths";
	import type { AlertDefinition } from "$lib/api";
	import ListRow from "$components/common/list-row/ListRow.svelte";
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

{#snippet facts()}
	{#if rosterName}
		<span>Roster: {rosterName}</span>
	{/if}
	{#if rosterName && groupLabels}
		<span aria-hidden="true">·</span>
	{/if}
	{#if groupLabels}
		<span>Grouped by {groupLabels}</span>
	{/if}
{/snippet}

<ListRow
	title={attrs.title}
	href={resolve("/signals/[id]/[[view=signalView]]", { id: definition.id })}
	summary={attrs.description}
	status={attention}
	facts={rosterName || groupLabels ? facts : undefined}
/>
