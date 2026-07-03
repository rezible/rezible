<script lang="ts">
	import type { Part } from "$lib/api";

	type Props = {
		part: Part;
	};

	const { part }: Props = $props();

	const resource = $derived(part.resource);
	const toolRequest = $derived(part.toolRequest);
	const toolResponse = $derived(part.toolResponse);
	const hasSummaryRows = $derived(
		!!part.kind || !!part.contentType || !!resource?.uri || !!toolRequest || !!toolResponse
	);
</script>

<div class="space-y-3 rounded border border-border bg-muted/30 p-3">
	{#if part.text}
		<p class="whitespace-pre-wrap text-sm leading-relaxed text-foreground">{part.text}</p>
	{/if}

	{#if hasSummaryRows}
		<dl class="grid gap-2 text-xs sm:grid-cols-2">
			{#if part.kind}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Kind</dt>
					<dd class="break-words text-foreground">{part.kind}</dd>
				</div>
			{/if}
			{#if part.contentType}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Content type</dt>
					<dd class="break-words text-foreground">{part.contentType}</dd>
				</div>
			{/if}
			{#if resource?.uri}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Resource URI</dt>
					<dd class="break-words text-foreground">{resource.uri}</dd>
				</div>
			{/if}
			{#if toolRequest?.name}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Tool request</dt>
					<dd class="break-words text-foreground">{toolRequest.name}</dd>
				</div>
			{/if}
			{#if toolRequest?.ref}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Tool request ref</dt>
					<dd class="break-words text-foreground">{toolRequest.ref}</dd>
				</div>
			{/if}
			{#if toolResponse?.name}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Tool response</dt>
					<dd class="break-words text-foreground">{toolResponse.name}</dd>
				</div>
			{/if}
			{#if toolResponse?.ref}
				<div class="min-w-0">
					<dt class="font-medium text-muted-foreground">Tool response ref</dt>
					<dd class="break-words text-foreground">{toolResponse.ref}</dd>
				</div>
			{/if}
		</dl>
	{/if}

	{#if toolRequest?.input}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Tool request input</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					toolRequest.input,
					null,
					2
				)}</pre>
		</div>
	{/if}

	{#if toolResponse?.output}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Tool response output</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					toolResponse.output,
					null,
					2
				)}</pre>
		</div>
	{/if}

	{#if toolResponse?.content}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Tool response content</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					toolResponse.content,
					null,
					2
				)}</pre>
		</div>
	{/if}

	{#if part.metadata}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Metadata</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					part.metadata,
					null,
					2
				)}</pre>
		</div>
	{/if}

	{#if part.custom}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Custom</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					part.custom,
					null,
					2
				)}</pre>
		</div>
	{/if}
</div>
