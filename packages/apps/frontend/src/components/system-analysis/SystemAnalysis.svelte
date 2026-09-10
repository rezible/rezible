<script lang="ts">
	import SystemDiagramWrapper from "./diagram/SystemDiagramWrapper.svelte";
	import { useSystemAnalysisController } from "./controller.svelte";
	import type { SystemAnalysisEntry } from "@rezible/api-client-ts";

	const controller = useSystemAnalysisController();
</script>

{#snippet unattachedEntries(entries: SystemAnalysisEntry[])}
	{#if entries.length}
		<section class="max-h-36 overflow-auto border-t border-border bg-card p-3">
			<h3 class="text-sm font-semibold">Unattached analysis entries</h3>
			{#each entries as entry (entry.id)}
				{@const { kind, title } = entry.attributes}
				<div class="mt-1 text-xs">
					<span class="text-muted-foreground">{kind}</span> · {title}
				</div>
			{/each}
		</section>
	{/if}
{/snippet}

{#if controller.graphError}
	<div class="flex h-full items-center justify-center p-6 text-destructive">
		<div>
			<p>Could not load this analysis graph.</p>
			<button
				class="mt-2 underline"
				onclick={() => {
					controller.refreshAll();
				}}>Retry</button
			>
		</div>
	</div>
{:else if controller.graphLoading}
	<div class="flex h-full items-center justify-center text-sm text-muted-foreground">
		Loading complete graph…
	</div>
{:else}
	<div class="flex h-full min-h-0 flex-col">
		<div class="min-h-0 grow">
			<SystemDiagramWrapper />
		</div>
		{@render unattachedEntries(controller.attachments.unattached)}
	</div>
{/if}
