<script lang="ts">
	import { MediaQuery } from "svelte/reactivity";
	import { Button } from "$components/ui/button";
	import * as Sheet from "$components/ui/sheet";
	import { Separator } from "$components/ui/separator";
	import { useIncidentAnalysis } from "../controller.svelte";
	const controller = useIncidentAnalysis();
	const desktop = new MediaQuery("(min-width: 1024px)");
</script>

{#if desktop.current}
	{#if controller.inspectorOpen}
		<aside
			aria-label="Analysis inspector"
			class="flex min-h-0 w-80 shrink-0 flex-col overflow-hidden rounded-md border border-border bg-card"
		>
			<div class="flex items-center justify-between gap-2 border-b p-3">
				<h2 class="text-sm font-semibold">Analysis details</h2>
				<Button
					variant="ghost"
					size="sm"
					onclick={() => {
						controller.closeInspector();
						controller.restoreFocus();
					}}>Close inspector</Button
				>
			</div>
			<div class="min-h-0 overflow-y-auto p-3">{@render details()}</div>
		</aside>
	{/if}
{:else}
	<Sheet.Root
		open={controller.inspectorOpen}
		onOpenChange={(open) => {
			if (!open) controller.closeInspector();
		}}
	>
		<Sheet.Content
			side="right"
			class="w-[calc(100vw-1rem)] sm:max-w-sm"
			onCloseAutoFocus={(event) => {
				event.preventDefault();
				controller.restoreFocus();
			}}
		>
			<Sheet.Header>
				<Sheet.Title>Analysis details</Sheet.Title>
				<Sheet.Description>Selected entry or subject and its recorded context.</Sheet.Description>
			</Sheet.Header>
			<div class="min-h-0 overflow-y-auto px-4 pb-4">{@render details()}</div>
		</Sheet.Content>
	</Sheet.Root>
{/if}

{#snippet details()}
	<div class="flex min-w-0 flex-col gap-3 break-words text-sm">
		{#if controller.selectedEntry}
			{@const entry = controller.selectedEntry}
			<div class="text-xs text-muted-foreground">{entry.attributes.kind}</div>
			<h3 class="font-semibold">{entry.attributes.title}</h3>
			{#if entry.attributes.occurredAt}<time datetime={entry.attributes.occurredAt}
					>{new Date(entry.attributes.occurredAt).toLocaleString()}</time
				>{/if}
			{#if entry.attributes.body}<p class="whitespace-pre-wrap">{entry.attributes.body}</p>{/if}
			<Button variant="outline" size="sm" onclick={() => controller.entryEditor.setEditing(entry)}
				>Edit entry</Button
			>
			{#if entry.attributes.reference}
				<Separator />
				<h4 class="font-medium">Source</h4>
				{#if controller.sourceUrl}<a
						class="underline"
						href={controller.sourceUrl}
						target="_blank"
						rel="noreferrer">{entry.attributes.reference}</a
					>
				{:else}<p>{entry.attributes.reference}</p>{/if}
			{/if}
			{#if controller.subjects.length}
				<Separator />
				<h4 class="font-medium">Subjects and evidence</h4>
				{#each controller.subjects as { subject, node, edge } (subject.id)}
					<div class="flex flex-col gap-1">
						<span class="text-xs text-muted-foreground">{subject.attributes.role}</span>
						{#if node}
							<a
								href={controller.selectionUrl({ nodeId: node.id })}
								onclick={(event) => {
									if (!event.ctrlKey && !event.metaKey && !event.shiftKey) {
										event.preventDefault();
										controller.select({ nodeId: node.id }, event.currentTarget);
									}
								}}
								class="underline"
								>{node.attributes.labelOverride ??
									node.attributes.knowledgeEntity.attributes.latestState?.displayName ??
									"Subject"}</a
							>
						{:else if edge}
							<a
								href={controller.selectionUrl({ edgeId: edge.id })}
								onclick={(event) => {
									if (!event.ctrlKey && !event.metaKey && !event.shiftKey) {
										event.preventDefault();
										controller.select({ edgeId: edge.id }, event.currentTarget);
									}
								}}
								class="underline"
								>{edge.attributes.knowledgeRelationship.attributes.predicate}</a
							>
						{:else}
							<p class="text-muted-foreground">
								{subject.attributes.knowledgeEvidenceId
									? "Evidence details unavailable"
									: "Subject details unavailable"}
							</p>
							<span class="text-xs"
								>{subject.attributes.knowledgeEvidenceId ??
									subject.attributes.normalizedEventId ??
									subject.attributes.knowledgeEntityId ??
									subject.attributes.knowledgeRelationshipId}</span
							>
						{/if}
					</div>
				{/each}
			{/if}
		{:else if controller.selectedNode || controller.selectedEdge}
			{#if controller.selectedNode}
				{@const node = controller.selectedNode}
				{@const entity = node.attributes.knowledgeEntity.attributes}
				<h3 class="font-semibold">
					{node.attributes.labelOverride ?? entity.latestState?.displayName ?? "Subject"}
				</h3>
				<p class="text-xs text-muted-foreground">{entity.category} · {entity.kind}</p>
				{#if node.attributes.descriptionOverride ?? entity.latestState?.description}<p
						class="whitespace-pre-wrap"
					>
						{node.attributes.descriptionOverride ?? entity.latestState?.description}
					</p>{/if}
			{:else if controller.selectedEdge}
				{@const relationship = controller.selectedEdge.attributes.knowledgeRelationship.attributes}
				<h3 class="font-semibold">
					{relationship.latestState?.displayName ?? relationship.predicate}
				</h3>
				<p>{relationship.predicate}</p>
				{#if relationship.latestState?.description}<p>{relationship.latestState.description}</p>{/if}
			{/if}
			{#if controller.relationships.length}
				<Separator />
				<h4 class="font-medium">Relationships</h4>
				{#each controller.relationships as edge (edge.id)}
					<a
						class="underline"
						href={controller.selectionUrl({ edgeId: edge.id })}
						onclick={(event) => {
							if (!event.ctrlKey && !event.metaKey && !event.shiftKey) {
								event.preventDefault();
								controller.select({ edgeId: edge.id }, event.currentTarget);
							}
						}}>{edge.attributes.knowledgeRelationship.attributes.predicate}</a
					>
				{/each}
			{/if}
			<Separator />
			<h4 class="font-medium">Analysis entries</h4>
			{#if controller.analysis.entriesQuery.isPending}<p role="status">Loading entries…</p>
			{:else if controller.analysis.entriesQuery.error}<p role="alert">
					Entries could not be refreshed.
				</p>
				<Button variant="outline" size="sm" onclick={controller.analysis.refreshEntries}
					>Retry entries</Button
				>{/if}
			{#each controller.attachedEntries as entry (entry.id)}
				<a
					class="underline"
					href={controller.selectionUrl({ entryId: entry.id })}
					onclick={(event) => {
						if (!event.ctrlKey && !event.metaKey && !event.shiftKey) {
							event.preventDefault();
							controller.selectEntry(entry.id, event.currentTarget);
						}
					}}>{entry.attributes.title}</a
				>
			{:else}
				{#if !controller.analysis.entriesQuery.isPending && !controller.analysis.entriesQuery.error}<p
						class="text-muted-foreground"
					>
						No entries attached.
					</p>{/if}
			{/each}
		{:else if controller.analysis.entriesQuery.isPending || controller.analysis.graphLoading}
			<p role="status">Loading selected record…</p>
		{:else}
			<p role="status">The selected record is unavailable.</p>
			<Button variant="outline" size="sm" onclick={controller.analysis.refreshAll}
				>Retry analysis</Button
			>
		{/if}
	</div>
{/snippet}
