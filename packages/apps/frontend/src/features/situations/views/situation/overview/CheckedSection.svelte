<script lang="ts">
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import type { CheckedState } from "$features/situations/lib/brief";
	import ProviderReadCitation from "../investigation/ProviderReadCitation.svelte";

	type Props = { state: CheckedState };
	const { state }: Props = $props();
</script>

<section aria-labelledby="checked-title" class="flex flex-col gap-3">
	<SectionHeading id="checked-title" title="What Rezible checked" />

	{#if state.kind === "no_report"}
		<p class="text-sm text-muted-foreground">No report has been published, so no reads are cited yet.</p>
	{:else if state.kind === "none_cited"}
		<p class="text-sm text-muted-foreground">The latest report cites no provider reads.</p>
	{:else}
		<ul class="flex flex-col gap-3">
			{#each state.reads as checked (checked.id)}
				<li>
					<ProviderReadCitation number={checked.number} read={checked.read} />
				</li>
			{/each}
		</ul>
	{/if}
</section>
