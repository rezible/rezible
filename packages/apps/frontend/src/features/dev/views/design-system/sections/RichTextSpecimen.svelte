<script lang="ts">
	import * as Card from "$components/ui/card";
	import RichTextView from "$components/rich-text-view/RichTextView.svelte";
	import { markdownSummary } from "$components/rich-text-view/markdown-summary";

	// Illustrative document; not product data.
	const reportMarkdown = [
		"# Checkout search degradation",
		"",
		"Search requests from checkout are **timing out** for a subset of customers. Redis connection pressure rose shortly after `search-api v2.18` deployed; see the [runbook](https://example.com/runbook).",
		"",
		"The cause is *unconfirmed*: temporal proximity does not establish causality.",
		"",
		"## Findings",
		"",
		"- Timeouts increased after the deployment",
		"- Redis pool pressure rose in the same window",
		"  - Pool wait time exceeded 800ms",
		"",
		"1. Compare pool metrics before and after the deployment",
		"2. Check whether the rollback reduced pressure",
		"",
		"### Supporting quote",
		"",
		"> Customers report that search keeps timing out during checkout.",
		"",
		"```",
		'redis_pool_wait_ms{service="search-api"} > 800',
		"```",
		"",
		"---",
		"",
		"| Signal | Source | Time |",
		"| --- | --- | --- |",
		"| Timeout rate | Application logs | 14:09 |",
		"| Pool pressure | Metrics | 14:07 |",
	].join("\n");

	const compactMarkdown = [
		"Redis connection pressure is the **strongest signal**.",
		"",
		"The deployment may have contributed; more evidence is needed.",
	].join("\n");

	const summary = markdownSummary(reportMarkdown);
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Rich text</Card.Title>
		<Card.Description>
			Read-only Markdown rendering through the shared Tiptap schema. The document is illustrative.
		</Card.Description>
	</Card.Header>
	<Card.Content class="flex flex-col gap-6">
		<RichTextView markdown={reportMarkdown} label="Illustrative report" />
		<div class="flex flex-col gap-2 border-t pt-4">
			<p class="region-label">Plain-text summary</p>
			<p class="text-sm">{summary}</p>
		</div>
		<div class="flex flex-col gap-2 border-t pt-4">
			<p class="region-label">Compact</p>
			<RichTextView markdown={compactMarkdown} size="compact" label="Illustrative finding details" />
		</div>
	</Card.Content>
</Card.Root>
