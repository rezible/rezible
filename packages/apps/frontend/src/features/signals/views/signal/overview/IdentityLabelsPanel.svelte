<script lang="ts">
	import * as Field from "$components/ui/field";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Input } from "$components/ui/input";
	import { Spinner } from "$components/ui/spinner";
	import { useAlertViewController } from "../controller.svelte";

	const view = useAlertViewController();
	const pending = $derived(view.labelsMutation.isPending);
	const labels = $derived(view.identityGroupLabels);
</script>

<section aria-labelledby="identity-labels-title" class="flex flex-col gap-3 rounded-lg border bg-card p-4">
	<h2 id="identity-labels-title" class="region-label">Identity group labels</h2>

	{#if labels.length}
		<ul class="flex flex-wrap gap-1.5" aria-label="Current identity group labels">
			{#each labels as label (label)}
				<li><Badge variant="secondary">{label}</Badge></li>
			{/each}
		</ul>
	{:else}
		<p class="text-sm text-muted-foreground">None set. The alert source's own grouping applies.</p>
	{/if}

	<Field.Field>
		<Field.Label for="identity-labels">Label names</Field.Label>
		<Input
			id="identity-labels"
			placeholder="service, cluster"
			value={view.labelsInput}
			disabled={pending}
			oninput={(event) => view.setLabelsInput(event.currentTarget.value)}
		/>
		<Field.Description>
			Separate names with commas. A change applies when this definition's next episode opens.
		</Field.Description>
	</Field.Field>

	<div class="flex flex-wrap gap-2">
		<Button size="sm" disabled={pending || !view.labelsChanged} onclick={view.saveLabels}>
			{#if pending}
				<Spinner data-icon="inline-start" />
			{/if}
			Save labels
		</Button>
		{#if labels.length}
			<Button size="sm" variant="outline" disabled={pending} onclick={view.clearLabels}>Clear</Button>
		{/if}
	</div>
</section>
