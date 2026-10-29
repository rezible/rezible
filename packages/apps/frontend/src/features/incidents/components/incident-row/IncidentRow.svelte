<script lang="ts">
	import { resolve } from "$app/paths";
	import type { Incident } from "$lib/api";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { incidentListRow } from "$features/incidents/lib/row";

	type Props = {
		incident: Incident;
	};

	let { incident }: Props = $props();

	const row = $derived(incidentListRow(incident));
	const href = $derived(resolve("/incidents/[ref]/[[view=incidentView]]", { ref: row.reference }));
</script>

{#snippet facts()}
	<span>Opened <Timestamp value={row.openedAt} /></span>
	{#if row.services.length}
		<span aria-hidden="true">·</span>
		<span>
			{row.services.join(", ")}{#if row.moreServices > 0}&nbsp;+{row.moreServices}{/if}
		</span>
	{/if}
{/snippet}

{#snippet aside()}
	{#if row.severity}
		<StatusBadge status={row.severity} />
	{/if}
	<span>Updated <Timestamp value={row.updatedAt} /></span>
{/snippet}

<ListRow
	title={row.title}
	{href}
	summary={row.summary}
	status={row.response}
	reference={row.reference}
	{facts}
	{aside}
/>
