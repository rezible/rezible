<script lang="ts">
	import { MediaQuery } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import * as Sheet from "$components/ui/sheet";
	import { MapInspector } from "$features/system/components/map-inspector";
	import { useIncidentAnalysis } from "../controller.svelte";
	import EntryDetails from "./EntryDetails.svelte";
	import SubjectDetails from "./SubjectDetails.svelte";

	const controller = useIncidentAnalysis();
	const details = $derived(controller.inspectorDetails);
	const desktop = new MediaQuery("(min-width: 1024px)");
</script>

{#snippet selectedDetails()}
	{#if controller.hasSelection}
		<div class="flex min-w-0 flex-col gap-3 break-words text-sm">
			{#if details?.kind === "entry"}
				<EntryDetails {details} />
			{:else if details?.kind === "milestone"}
				<h3 class="font-semibold capitalize">{details.attrs.kind}</h3>
				<time datetime={details.attrs.timestamp}>{details.occurredAtLabel}</time>
				<p class="whitespace-pre-wrap">{details.attrs.description}</p>
				{#if details.attrs.source}
					<p class="text-muted-foreground">Source: {details.attrs.source}</p>
				{/if}
				{#if details.attrs.user?.attributes?.name}
					<p>{details.attrs.user.attributes.name}</p>
				{/if}
			{:else if details?.kind === "subject"}
				<SubjectDetails {details} />
			{:else if controller.mapSelection?.kind === "summary"}
				<p role="status">The selected summary is shown in the map inspector.</p>
			{:else if controller.selectedRecordLoading}
				<p role="status">Loading selected record…</p>
			{:else}
				<p role="status">The selected record is unavailable.</p>
				<Button variant="outline" size="sm" onclick={controller.retrySelection}>
					Retry analysis
				</Button>
			{/if}
		</div>
	{/if}
{/snippet}

{#if desktop.current}
	{#if controller.inspectorOpen}
		<aside
			aria-label="Analysis inspector"
			class="flex min-h-0 w-80 shrink-0 flex-col overflow-hidden rounded-md border border-border bg-card"
		>
			<div class="flex items-center justify-between gap-2 border-b p-3">
				<h2 class="text-sm font-semibold">Analysis inventory</h2>
				<Button variant="ghost" size="sm" onclick={controller.dismissInspector}>
					Close inspector
				</Button>
			</div>
			<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden p-3">
				{#if controller.systemAnalysis.graph}
					<div class="min-h-0 flex-1">
						<MapInspector
							graph={controller.systemAnalysis.graph}
							selection={controller.mapSelection}
							onSelectionChange={controller.systemAnalysis.onSelected}
						/>
					</div>
				{/if}
				<div class="max-h-64 shrink-0 overflow-y-auto">
					{@render selectedDetails()}
				</div>
			</div>
		</aside>
	{/if}
{:else}
	<Sheet.Root open={controller.inspectorOpen} onOpenChange={controller.setInspectorOpen}>
		<Sheet.Content
			side="right"
			class="w-[calc(100vw-1rem)] sm:max-w-sm"
			onCloseAutoFocus={controller.restoreInspectorFocus}
		>
			<Sheet.Header>
				<Sheet.Title>Analysis inventory</Sheet.Title>
				<Sheet.Description>
					Browse supplied analysis sources and inspect selected entries or subjects.
				</Sheet.Description>
			</Sheet.Header>
			<div class="min-h-0 flex-1 overflow-y-auto px-4 pb-4">
				<div class="flex min-h-0 flex-col gap-3">
					{#if controller.systemAnalysis.graph}
						<MapInspector
							graph={controller.systemAnalysis.graph}
							selection={controller.mapSelection}
							onSelectionChange={controller.systemAnalysis.onSelected}
						/>
					{/if}
					{@render selectedDetails()}
				</div>
			</div>
		</Sheet.Content>
	</Sheet.Root>
{/if}
