<script lang="ts">
	import type { Snippet } from "svelte";
	import { Button } from "$components/ui/button";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import SystemDiagram from "./diagram/SystemDiagram.svelte";
	import type { GraphInteraction } from "./diagram/diagramController.svelte";
	import { useSystemAnalysisController } from "./controller.svelte";

	type Props = { 
		interaction?: GraphInteraction;
		inspector?: Snippet,
	};
	let { interaction, inspector }: Props = $props();

	const controller = useSystemAnalysisController();

	const graphError = $derived(controller.graphError ?? controller.entriesQuery.error);
	const graphErrorText = $derived(controller.hasGraph
		? "Refresh failed. Showing available analysis."
		: "Could not load analysis.");
</script>

<div class="flex h-full min-h-0 flex-col gap-2">
	{#if !controller.analysisId}
		<p class="p-6 text-sm text-muted-foreground">No analysis is associated with this incident.</p>
	{:else}
		{#if graphError}
			<div role="alert" class="flex shrink-0 flex-wrap items-center gap-2 p-2">
				<span class="text-sm">{graphErrorText}</span>
				<ErrorAlert error={graphError} />
				<Button variant="outline" size="sm" onclick={controller.refreshAll}>Retry analysis</Button>
			</div>
		{/if}
		
		{#if controller.hasGraph}
			<div class="min-h-0 flex-1">
				<SystemDiagram {interaction} {inspector} />
			</div>
		{:else if !graphError}
			<p role="status" class="p-6 text-sm text-muted-foreground">Loading analysis…</p>
		{/if}
	{/if}
</div>
