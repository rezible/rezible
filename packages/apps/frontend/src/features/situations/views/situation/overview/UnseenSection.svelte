<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import type { UnseenState } from "$features/situations/lib/brief";
	import { useSituationOverviewController } from "./controller.svelte";

	type Props = { state: UnseenState };
	const { state }: Props = $props();

	const controller = useSituationOverviewController();
</script>

<section aria-labelledby="unseen-title" class="flex flex-col gap-3">
	<SectionHeading id="unseen-title" title="What Rezible cannot see" />

	{#if state.kind === "loading"}
		<div class="flex flex-col gap-2" aria-label="Loading what Rezible cannot see">
			<Skeleton class="h-5 w-3/5" />
		</div>
	{:else if state.kind === "error"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
			<span>What Rezible cannot see is unavailable.</span>
			<Button variant="outline" size="sm" onclick={controller.retryBrief}>Retry</Button>
		</div>
	{:else}
		<ul class="flex list-disc flex-col gap-1 pl-5 text-sm">
			{#each state.entries as entry (entry.key)}
				<li class="wrap-anywhere">{entry.text}</li>
			{/each}
		</ul>
	{/if}
</section>
