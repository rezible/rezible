<script lang="ts">
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import type { SituationTime } from "$features/situations/lib/status";

	type Props = { times: SituationTime[] };
	const { times }: Props = $props();
</script>

{#each times as time, index (time.key)}
	{#if index > 0}
		<span aria-hidden="true">·</span>
	{/if}
	<span>{time.label}</span>
	<Timestamp value={time.at} />
	{#if time.reason}
		<span>· {time.reason}</span>
	{/if}
{/each}
