<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Separator } from "$components/ui/separator";
	import { useIncidentAnalysis, type InspectorEntry } from "../controller.svelte";

	type Props = {
		details: InspectorEntry;
	};
	let { details }: Props = $props();

	const controller = useIncidentAnalysis();
	const attrs = $derived(details.attrs);
</script>

<div class="text-xs text-muted-foreground">{attrs.kind}</div>

<h3 class="font-semibold">{attrs.title}</h3>

{#if attrs.occurredAt}
	<time datetime={attrs.occurredAt}>{details.occurredAtLabel}</time>
{/if}

{#if attrs.body}
	<p class="whitespace-pre-wrap">{attrs.body}</p>
{/if}

<Button variant="outline" size="sm" onclick={controller.editSelectedEntry}>Edit entry</Button>

{#if attrs.reference}
	<Separator />
	<section class="flex flex-col gap-3" aria-label="Source">
		<h4 class="font-medium">Source</h4>
		{#if controller.sourceUrl}
			<a class="underline" href={controller.sourceUrl} target="_blank" rel="noreferrer">
				{attrs.reference}
			</a>
		{:else}
			<p>{attrs.reference}</p>
		{/if}
	</section>
{/if}

{#if controller.subjects.length}
	<Separator />
	<section class="flex flex-col gap-3" aria-label="Subjects and evidence">
		<h4 class="font-medium">Subjects and evidence</h4>
		{#each controller.subjects as subject (subject.id)}
			{@const selection = subject.selection}
			<div class="flex flex-col gap-1">
				<span class="text-xs text-muted-foreground">{subject.role}</span>
				{#if selection}
					<a
						class="underline"
						href={controller.selectionUrl(selection)}
						onclick={(event) => controller.selectInspectorLink(event, selection)}
					>
						{subject.label}
					</a>
				{:else}
					<p class="text-muted-foreground">{subject.unavailable}</p>
					<span class="text-xs">{subject.reference}</span>
				{/if}
			</div>
		{/each}
	</section>
{/if}
