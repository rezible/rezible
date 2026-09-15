<script lang="ts">
	import { MediaQuery } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import * as Sheet from "$components/ui/sheet";
	import { useIncidentAnalysis } from "../controller.svelte";
	import EntryDetails from "./EntryDetails.svelte";
	import SubjectDetails from "./SubjectDetails.svelte";

	const controller = useIncidentAnalysis();
	const details = $derived(controller.inspectorDetails);
	const desktop = new MediaQuery("(min-width: 1024px)");
</script>

{#snippet selectedDetails()}
	<div class="flex min-w-0 flex-col gap-3 break-words text-sm">
		{#if details?.kind === "entry"}
			<EntryDetails {details} />
		{:else if details?.kind === "subject"}
			<SubjectDetails {details} />
		{:else if controller.selectedRecordLoading}
			<p role="status">Loading selected record…</p>
		{:else}
			<p role="status">The selected record is unavailable.</p>
			<Button variant="outline" size="sm" onclick={controller.systemAnalysis.refreshAll}>
				Retry analysis
			</Button>
		{/if}
	</div>
{/snippet}

{#if desktop.current}
	{#if controller.inspectorOpen}
		<aside
			aria-label="Analysis inspector"
			class="flex min-h-0 w-80 shrink-0 flex-col overflow-hidden rounded-md border border-border bg-card"
		>
			<div class="flex items-center justify-between gap-2 border-b p-3">
				<h2 class="text-sm font-semibold">Analysis details</h2>
				<Button variant="ghost" size="sm" onclick={controller.dismissInspector}>
					Close inspector
				</Button>
			</div>
			<div class="min-h-0 overflow-y-auto p-3">
				{@render selectedDetails()}
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
				<Sheet.Title>Analysis details</Sheet.Title>
				<Sheet.Description>Selected entry or subject and its recorded context.</Sheet.Description>
			</Sheet.Header>
			<div class="min-h-0 overflow-y-auto px-4 pb-4">
				{@render selectedDetails()}
			</div>
		</Sheet.Content>
	</Sheet.Root>
{/if}
