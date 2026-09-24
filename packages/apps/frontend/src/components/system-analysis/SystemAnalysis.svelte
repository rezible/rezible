<script lang="ts">
	import type { Snippet } from "svelte";
	import { Button } from "$components/ui/button";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { MapInspector } from "$features/system/components/map-inspector";
	import { SystemMap } from "$features/system/components/system-map";
	import { useSystemAnalysisController } from "./controller.svelte";

	type Props = {
		inspector?: Snippet;
	};
	let { inspector: selectionInspector = defaultInspector }: Props = $props();

	const controller = useSystemAnalysisController();
	const queryError = $derived(controller.graphError ?? controller.entriesQuery.error);
	const queryErrorText = $derived(
		controller.hasGraph ? "Refresh failed. Showing available analysis." : "Could not load analysis."
	);
</script>

{#snippet defaultInspector()}
	<aside
		class="flex min-h-0 w-80 shrink-0 flex-col overflow-hidden rounded-md border border-border bg-card"
	>
		<div class="min-h-0 flex-1">
			<MapInspector
				graph={controller.graph!}
				selection={controller.selection}
				onSelectionChange={controller.onSelected}
			/>
		</div>

		{#if controller.selectionInspector}
			<section class="max-h-64 shrink-0 overflow-y-auto border-t p-3" aria-label="Attached entries">
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
									<p class="mt-1 whitespace-pre-wrap">{entry.body}</p>
								{/if}
							</article>
						{/each}
					</div>
				{/if}
			</section>
		{/if}
	</aside>
{/snippet}

<div class="flex h-full min-h-0 flex-col gap-2">
	{#if !controller.analysisId}
		<p class="p-6 text-sm text-muted-foreground">No analysis is available.</p>
	{:else}
		{#if queryError}
			<div role="alert" class="flex shrink-0 flex-wrap items-center gap-2 p-2">
				<span class="text-sm">{queryErrorText}</span>
				<ErrorAlert error={queryError} />
				<Button variant="outline" size="sm" onclick={controller.refreshAll}>Retry analysis</Button>
			</div>
		{/if}

		{#if controller.hasGraph && controller.graph}
			<div class="flex min-h-0 flex-1 gap-3 overflow-hidden">
				<div class="relative min-w-0 flex-1">
					<!-- Graph removal and canvas context-menu controls remain unmounted for future reuse; timeline controls stay separate. -->
					{#key controller.analysisId}
						<SystemMap
							graph={controller.graph}
							positions={controller.positions}
							selection={controller.selection}
							highlights={controller.highlights}
							onSelectionChange={controller.onSelected}
						/>
					{/key}
				</div>

				{@render selectionInspector()}
			</div>
		{:else if !queryError}
			<p role="status" class="p-6 text-sm text-muted-foreground">
				{controller.graphLoading ? "Loading analysis…" : "Preparing analysis…"}
			</p>
		{/if}
	{/if}
</div>
