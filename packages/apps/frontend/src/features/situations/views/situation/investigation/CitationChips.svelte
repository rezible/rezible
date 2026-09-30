<script lang="ts">
	import type { CitationView } from "$features/situations/lib/investigation-outputs";
	import { Button } from "$components/ui/button";
	import { useSituationInvestigationController } from "./controller.svelte";

	type Props = { citations: CitationView[] };
	const { citations }: Props = $props();

	const controller = useSituationInvestigationController();
</script>

{#if citations.length}
	<div class="flex flex-wrap gap-1">
		{#each citations as citation (citation.id)}
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
