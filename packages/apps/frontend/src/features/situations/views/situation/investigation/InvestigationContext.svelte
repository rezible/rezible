<script lang="ts">
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();
	const attributes = $derived(controller.investigationAttributes);
	const question = $derived(attributes?.query.trim() ?? "");
</script>

<section aria-labelledby="investigation-question-title" class="flex flex-col gap-2">
	<SectionHeading id="investigation-question-title" title="Question" level={3} />
	{#if question}
		<p class="text-sm whitespace-pre-wrap wrap-anywhere">{question}</p>
	{:else}
		<p class="text-sm text-muted-foreground">
			No question was provided. The investigation examines the situation's evidence.
		</p>
	{/if}
</section>

<section aria-labelledby="investigation-run-title" class="flex flex-col gap-2">
	<SectionHeading id="investigation-run-title" title="Run" level={3} />
	<dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
		<dt class="text-muted-foreground">Status</dt>
		<dd><StatusBadge status={controller.run} variant="inline" /></dd>
		<dt class="text-muted-foreground">Started</dt>
		<dd><Timestamp value={attributes?.createdAt} /></dd>
		<dt class="text-muted-foreground">Last updated</dt>
		<dd><Timestamp value={attributes?.updatedAt} /></dd>
		<dt class="text-muted-foreground">Evidence cited</dt>
		<dd class="tabular-nums">{controller.citedEvidenceCount}</dd>
	</dl>
</section>
