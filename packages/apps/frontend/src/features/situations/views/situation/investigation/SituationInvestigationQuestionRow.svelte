<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { cn } from "$lib/utils";
	import { useSituationInvestigationController, type InvestigationQuestionRow } from "./controller.svelte";

	type Props = { row: InvestigationQuestionRow };
	const { row }: Props = $props();
	const controller = useSituationInvestigationController();
</script>

<li
	class={cn(
		"flex flex-col gap-3 rounded-md border p-4",
		controller.submittedQuestion?.id === row.input.id && "border-primary bg-muted/40"
	)}
>
	<div class="flex flex-wrap items-center justify-between gap-2">
		<Badge variant="secondary">{row.statusLabel}</Badge>
		<time datetime={row.submittedAt.iso} class="text-xs text-muted-foreground">
			Asked {row.submittedAt.label}
		</time>
	</div>
	<p class="whitespace-pre-wrap text-sm leading-relaxed wrap-anywhere">
		{row.input.attributes.text}
	</p>
	{#if row.noAnswerPublished}
		<p class="text-sm text-muted-foreground">No answer was published.</p>
	{/if}

	{#if row.hasSelectedAnswer || row.expanded}
		<Button
			variant="outline"
			size="sm"
			aria-expanded={row.expanded}
			aria-controls={`answer-${row.input.id}`}
			onclick={() => controller.toggleAnswer(row.input.id)}
		>
			{row.expanded ? "Hide answer" : "Show answer"}
		</Button>

		{#if row.expanded}
			<div id={`answer-${row.input.id}`} class="flex flex-col gap-3 border-t pt-3">
				{#if row.answerNeedsUpdate}
					<p role="status" class="text-sm text-muted-foreground">
						Updating answer selection. No current answer is shown.
					</p>
				{:else if row.answerUnavailable}
					<p role="status" class="text-sm text-muted-foreground">Answer unavailable.</p>
				{:else if row.answerPending}
					<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
						<Spinner /> Loading selected answer
					</p>
				{:else if row.answer}
					{@const answer = row.answer.attributes}

					{#if row.answerStale}
						<div
							role="status"
							class="flex items-center justify-between gap-3 text-xs text-muted-foreground"
						>
							<span>Refresh failed. Showing the previously loaded answer.</span>
							<Button variant="ghost" size="sm" onclick={() => row.answerQuery?.refetch()}>
								Retry
							</Button>
						</div>
					{/if}
					<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
						<time datetime={row.answerPublishedAt.iso}>
							Published {row.answerPublishedAt.label}
						</time>
						<Badge variant={answer.turnStatus === "running" ? "outline" : "secondary"}>
							{answer.turnStatus === "running" ? "Running" : "Completed"}
						</Badge>
						{#if answer.provisional}
							<Badge variant="outline">Provisional</Badge>
						{/if}
					</div>
					<h3 class="text-base font-semibold">{answer.title || "Finding"}</h3>
					<p class="whitespace-pre-wrap text-sm leading-relaxed wrap-anywhere">
						{answer.body || "Answer text unavailable."}
					</p>
					{#if answer.invalidatedByVersionIds.length}
						<p class="text-sm text-muted-foreground">A later finding disputes this answer.</p>
					{/if}
					<div class="flex flex-col gap-2">
						<h4 class="text-sm font-medium">Evidence references</h4>
						{#if answer.references.length}
							<ul class="flex flex-wrap gap-x-3 gap-y-1">
								{#each answer.references as reference, index (reference.id)}
									<li>
										<Button
											variant="link"
											class="h-auto p-0"
											aria-label={`Inspect Evidence ${index + 1}`}
											onclick={(event) =>
												controller.inspection.openEvidence(
													reference.id,
													event.currentTarget
												)}
										>
											Evidence {index + 1}
										</Button>
									</li>
								{/each}
							</ul>
						{:else}
							<p class="text-sm text-muted-foreground">No evidence references.</p>
						{/if}
					</div>
				{:else if row.answerLoadFailed}
					<div role="alert" class="flex flex-col items-start gap-2 text-sm">
						<p>The selected answer could not be loaded.</p>
						<Button variant="outline" size="sm" onclick={() => row.answerQuery?.refetch()}>
							Retry
						</Button>
					</div>
				{:else}
					<p role="status" class="text-sm text-muted-foreground">No current answer is selected.</p>
				{/if}
			</div>
		{/if}
	{/if}
</li>
