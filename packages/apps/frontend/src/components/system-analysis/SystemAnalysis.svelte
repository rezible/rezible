<script lang="ts">
	import type { Snippet } from "svelte";
	import { Button } from "$components/ui/button";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { SystemDiagram } from "$components/system-diagram";
	import RelationshipEdge from "./relationship-edge/RelationshipEdge.svelte";
	import { Panel } from "@xyflow/svelte";
	import ContextMenu from "$components/common/context-menu/ContextMenu.svelte";
	import { useSystemAnalysisController } from "./controller.svelte";

	type Props = {
		inspector?: Snippet;
	};
	let { inspector: selectionInspector = defaultInspector }: Props = $props();

	const controller = useSystemAnalysisController();

	const graphError = $derived(controller.graphError ?? controller.entriesQuery.error);
	const graphErrorText = $derived(
		controller.hasGraph ? "Refresh failed. Showing available analysis." : "Could not load analysis."
	);
</script>

{#snippet removeSelectionButton()}
	{@const subject = controller.selection.nodeId ? "node" : "relationship"}
	<Button
		variant="outline"
		disabled={controller.deleting}
		onclick={() => controller.remove(controller.selection)}
	>
		{controller.deleting ? "Removing…" : `Remove ${subject}`}
	</Button>
{/snippet}

{#snippet defaultInspector()}
	{#if !!controller.selectionInspector}
		<aside
			class="w-72 max-h-80 overflow-auto border border-border bg-card p-3 text-card-foreground shadow"
		>
			<h3 class="mb-2 text-sm font-semibold">{controller.selectionInspector.title}</h3>
			{#if controller.selectionInspector.entries.length === 0}
				<p class="text-xs text-muted-foreground">No entries attached.</p>
			{:else}
				<div class="flex flex-col gap-2">
					{#each controller.selectionInspector.entries as entry (entry.id)}
						<article class="border-l-2 border-primary pl-2 text-xs">
							<div class="text-muted-foreground">
								{entry.kind}
								{#if entry.occurredAt}
									· {entry.occurredAt}
								{/if}
							</div>
							<div class="font-medium">{entry.title}</div>
							{#if entry.body}
								<p class="mt-1 whitespace-pre-wrap">
									{entry.body}
								</p>
							{/if}
						</article>
					{/each}
				</div>
			{/if}
		</aside>
	{/if}
{/snippet}

<div class="flex h-full min-h-0 flex-col gap-2">
	{#if !controller.analysisId}
		<p class="p-6 text-sm text-muted-foreground">No analysis is available.</p>
	{:else}
		{#if graphError}
			<div role="alert" class="flex shrink-0 flex-wrap items-center gap-2 p-2">
				<span class="text-sm">{graphErrorText}</span>
				<ErrorAlert error={graphError} />
				<Button variant="outline" size="sm" onclick={controller.refreshAll}>Retry analysis</Button>
			</div>
		{/if}

		{#if controller.hasGraph}
			{#if controller.mutationError}
				<ErrorAlert error={controller.mutationError} />
			{/if}

			<div class="flex min-h-0 flex-1 gap-3 overflow-hidden">
				<div class="relative min-w-0 flex-1" {@attach controller.setContainer}>
					<SystemDiagram
						controller={controller.diagram}
						edgeTypes={{ relationship: RelationshipEdge }}
					>
						{#if !controller.readOnly}
							{#if controller.selection.nodeId || controller.selection.edgeId}
								<Panel position="bottom-right">
									{@render removeSelectionButton()}
								</Panel>
							{/if}

							{#if controller.ctxMenu && controller.container}
								<ContextMenu
									title="Analysis actions"
									containerRect={controller.container.getBoundingClientRect()}
									clickPos={controller.ctxMenu.position}
								>
									<Button
										variant="ghost"
										disabled={controller.deleting}
										onclick={() =>
											controller.ctxMenu &&
											controller.remove(controller.ctxMenu.selection)}
									>
										{controller.ctxMenu.selection.nodeId
											? "Remove node"
											: "Remove relationship"}
									</Button>
								</ContextMenu>
							{/if}
						{/if}
					</SystemDiagram>
				</div>

				{@render selectionInspector()}
			</div>
		{:else if !graphError}
			<p role="status" class="p-6 text-sm text-muted-foreground">Loading analysis…</p>
		{/if}
	{/if}
</div>
