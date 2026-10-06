<script lang="ts">
	import type { Situation } from "$lib/api";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { situationSourceCount } from "$features/situations/lib/model";
	import { situationHref } from "$features/situations/lib/routes";
	import { longRunningStatus, situationStateTime, situationStatus } from "$features/situations/lib/status";

	type Props = {
		situation: Situation;
		/** Render time, used only to decide the long-running marker. */
		now: number;
	};

	let { situation, now }: Props = $props();

	const attributes = $derived(situation.attributes);
	const status = $derived(situationStatus(attributes));
	const longRunning = $derived(longRunningStatus(attributes, now));
	const sourceCount = $derived(situationSourceCount(attributes));
	const sourceLabel = $derived(sourceCount === 1 ? "source" : "sources");
	const stateTime = $derived(situationStateTime(attributes));
</script>

<li class="flex min-w-0 flex-col gap-1 px-4 py-3">
	<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
		<a
			href={situationHref(situation.id)}
			class="min-w-0 rounded-sm text-[15px] leading-[22px] font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring"
		>
			{attributes.title}
		</a>
		<StatusBadge {status} variant="inline" />
		{#if longRunning}
			<StatusBadge status={longRunning} variant="inline" />
		{/if}
	</div>
	{#if attributes.summary}
		<p class="line-clamp-2 text-sm text-muted-foreground wrap-anywhere">{attributes.summary}</p>
	{/if}
	<p class="flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground">
		<span class="tabular-nums">{sourceCount} {sourceLabel}</span>
		<span aria-hidden="true">·</span>
		<span>{stateTime.label}</span>
		<Timestamp value={stateTime.at} />
		{#if stateTime.reason}
			<span>· {stateTime.reason}</span>
		{/if}
	</p>
</li>
