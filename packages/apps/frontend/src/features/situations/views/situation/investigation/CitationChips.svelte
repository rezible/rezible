<script lang="ts">
	import type { CitationView } from "$features/situations/lib/investigation-outputs";
	import { Button } from "$components/ui/button";
	import { useSituationInvestigationController } from "./controller.svelte";
	import ProviderReadCitation from "./ProviderReadCitation.svelte";

	type Props = { citations: CitationView[] };
	const { citations }: Props = $props();

	const controller = useSituationInvestigationController();

	const evidenceCitations = $derived(citations.filter((citation) => !citation.providerRead));
	const providerReadCitations = $derived(citations.filter((citation) => !!citation.providerRead));
</script>

{#if evidenceCitations.length}
	<div class="flex flex-wrap gap-1">
		{#each evidenceCitations as citation (citation.id)}
			<Button
				variant="outline"
				size="sm"
				class="h-6 min-w-7 px-1.5 text-xs tabular-nums"
				aria-label={`Inspect evidence ${citation.number}`}
				onclick={(event) => controller.openCitation(citation, event.currentTarget)}
			>
				[{citation.number}]
			</Button>
		{/each}
	</div>
{/if}
{#if providerReadCitations.length}
	<ul class="flex flex-col gap-1.5">
		{#each providerReadCitations as citation (citation.id)}
			{#if citation.providerRead}
				<li>
					<ProviderReadCitation number={citation.number} read={citation.providerRead} />
				</li>
			{/if}
		{/each}
	</ul>
{/if}
