<script lang="ts">
	import type { Situation } from "$lib/api";
	import { Skeleton } from "$components/ui/skeleton";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import SituationTimes from "$features/situations/components/situation-times/SituationTimes.svelte";
	import type { SituationConclusion } from "$features/situations/lib/model";
	import { situationHref } from "$features/situations/lib/routes";
	import {
		longRunningStatus,
		situationStatus,
		situationSummaryTimes,
	} from "$features/situations/lib/status";

	type Props = {
		situation: Situation;
		/** Render time, used only to decide the long-running marker. */
		now?: number;
		/** Latest investigation conclusion; shown in place of the summary when published. */
		conclusion?: SituationConclusion;
	};

	let { situation, now = Date.now(), conclusion }: Props = $props();

	const attributes = $derived(situation.attributes);
	const status = $derived(situationStatus(attributes));
	const longRunning = $derived(longRunningStatus(attributes, now));
	const observationCount = $derived(attributes.observations.length);
	const linkedIncidentCount = $derived(attributes.linkedIncidentIds.length);
	const times = $derived(situationSummaryTimes(attributes));
	const summary = $derived(attributes.summary.trim());
</script>

{#snippet body()}
	{#if conclusion?.kind === "loading"}
		<Skeleton class="h-5 w-3/4" />
	{:else if conclusion?.kind === "text"}
		<p class="line-clamp-1 text-sm">
			<span class="text-muted-foreground">Latest conclusion ·</span>
			{conclusion.text}
		</p>
	{:else if summary}
		<p class="line-clamp-1 text-sm text-muted-foreground wrap-anywhere">{summary}</p>
	{/if}
{/snippet}

{#snippet facts()}
	<SituationTimes {times} />
	<span aria-hidden="true">·</span>
	<span class="tabular-nums">
		{observationCount}
		{observationCount === 1 ? "observation" : "observations"}
	</span>
	{#if linkedIncidentCount > 0}
		<span aria-hidden="true">·</span>
		<span class="tabular-nums">
			Linked to {linkedIncidentCount}
			{linkedIncidentCount === 1 ? "incident" : "incidents"}
		</span>
	{/if}
{/snippet}

{#snippet aside()}
	{#if longRunning}
		<StatusBadge status={longRunning} variant="inline" />
	{/if}
{/snippet}

<ListRow
	title={attributes.title}
	href={situationHref(attributes.reference)}
	{body}
	{status}
	reference={attributes.reference}
	{facts}
	aside={longRunning ? aside : undefined}
/>
