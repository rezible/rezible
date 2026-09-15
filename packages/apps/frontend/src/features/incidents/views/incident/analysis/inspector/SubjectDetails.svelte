<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Separator } from "$components/ui/separator";
	import { useIncidentAnalysis, type InspectorSubject } from "../controller.svelte";

	type Props = {
		details: InspectorSubject;
	};
	let { details }: Props = $props();

	const controller = useIncidentAnalysis();
	const entriesQuery = $derived(controller.systemAnalysis.entriesQuery);
</script>

<h3 class="font-semibold">{details.title}</h3>

<p class="text-xs text-muted-foreground">{details.context}</p>

{#if details.description}systemAnalysis
	<p class="whitespace-pre-wrap">{details.description}</p>
{/if}

{#if controller.relationships.length}
	<Separator />
	<section class="flex flex-col gap-3" aria-label="Relationships">
		<h4 class="font-medium">Relationships</h4>
		{#each controller.relationships as relationship (relationship.id)}
			<a
				class="underline"
				href={controller.selectionUrl({ edgeId: relationship.id })}
				onclick={(event) => controller.selectInspectorLink(event, { edgeId: relationship.id })}
			>
				{relationship.label}
			</a>
		{/each}
	</section>
{/if}

<Separator />

<section class="flex flex-col gap-3" aria-label="Analysis entries">
	<h4 class="font-medium">Analysis entries</h4>
	{#if entriesQuery.isPending}
		<p role="status">Loading entries…</p>
	{:else}
		{#if entriesQuery.error}
			<p role="alert">Entries could not be refreshed.</p>
			<Button variant="outline" size="sm" onclick={controller.systemAnalysis.refreshEntries}>
				Retry entries
			</Button>
		{/if}
		{#each controller.attachedEntries as entry (entry.id)}
			<a
				class="underline"
				href={controller.selectionUrl({ entryId: entry.id })}
				onclick={(event) => controller.selectInspectorLink(event, { entryId: entry.id })}
			>
				{entry.attributes.title}
			</a>
		{:else}
			{#if !entriesQuery.error}
				<p class="text-muted-foreground">No entries attached.</p>
			{/if}
		{/each}
	{/if}
</section>
