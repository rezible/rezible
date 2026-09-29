<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import * as Field from "$components/ui/field";
	import * as Pagination from "$components/ui/pagination";
	import { Spinner } from "$components/ui/spinner";
	import { Textarea } from "$components/ui/textarea";
	import SituationInvestigationQuestionRow from "./SituationInvestigationQuestionRow.svelte";
	import { useSituationInvestigationController, QUESTION_PAGE_SIZE } from "./controller.svelte";

	const controller = useSituationInvestigationController();
</script>

<section aria-labelledby="follow-up-questions-title" class="flex flex-col gap-4">
	<h2 id="follow-up-questions-title" class="text-lg font-semibold">Questions and answers</h2>

	<form
		class="flex flex-col items-start gap-3"
		onsubmit={(event) => {
			event.preventDefault();
			void controller.submitQuestion();
		}}
	>
		<Field.FieldGroup>
			<Field.Field data-invalid={controller.questionValidationError}>
				<Field.FieldLabel for="investigation-follow-up-question">Ask a question</Field.FieldLabel>
				<Textarea
					id="investigation-follow-up-question"
					bind:value={controller.questionDraft}
					rows={4}
					disabled={controller.questionSubmitting}
					aria-invalid={controller.questionValidationError}
					oninput={(event) => controller.updateQuestionDraft(event.currentTarget.value)}
				/>
				<Field.FieldDescription>
					Ask about the investigation's current report. Whitespace-only questions are not accepted.
				</Field.FieldDescription>
				{#if controller.questionValidationError}
					<Field.FieldError>Enter a question before submitting.</Field.FieldError>
				{/if}
			</Field.Field>
		</Field.FieldGroup>

		<Button type="submit" disabled={controller.questionSubmitting}>
			{#if controller.questionSubmitting}
				<Spinner data-icon="inline-start" /> Submitting
			{:else}
				Ask
			{/if}
		</Button>
	</form>

	{#if controller.questionError}
		<Alert.Root variant="destructive">
			<Alert.Title>Question submission failed</Alert.Title>
			<Alert.Description>
				<p>{controller.questionError}</p>
				<p>
					Your draft is retained. Submit it again without changing the text to reuse the same
					submission key; changing the text starts a new submission.
				</p>
			</Alert.Description>
		</Alert.Root>
	{/if}

	{#if controller.questionSuccessMessage}
		<p role="status" class="rounded-md border bg-muted/40 px-3 py-2 text-sm">
			{controller.questionSuccessMessage}
		</p>
	{/if}

	{#if controller.questionInputsUnavailable}
		<p role="status" class="text-sm text-muted-foreground">Question history is unavailable.</p>
	{:else if controller.questionQuery.isPending && !controller.questionQuery.data}
		<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
			<Spinner /> Loading questions
		</p>
	{:else if controller.questionQuery.isError && !controller.questionQuery.data}
		<div role="alert" class="flex flex-col items-start gap-3 text-sm">
			<p>Question history could not be loaded.</p>
			<Button variant="outline" size="sm" onclick={() => controller.questionQuery.refetch()}>
				Retry
			</Button>
		</div>
	{:else}
		{#if controller.questionQuery.isError}
			<div role="status" class="flex items-center justify-between gap-3 text-xs text-muted-foreground">
				<span>Refresh failed. Showing previously loaded questions.</span>
				<Button variant="ghost" size="sm" onclick={() => controller.questionQuery.refetch()}>
					Retry
				</Button>
			</div>
		{/if}

		{#if controller.answerRows.length}
			<ol class="flex flex-col gap-3">
				{#each controller.answerRows as row (row.input.id)}
					<SituationInvestigationQuestionRow {row} />
				{/each}
			</ol>
		{:else}
			<p class="text-sm text-muted-foreground">No questions have been submitted.</p>
		{/if}

		<div
			class="flex flex-wrap items-center justify-between gap-3 border-t pt-3 text-sm text-muted-foreground"
		>
			<span>
				{#if controller.questionTotal}
					Page {controller.questionPage.paginator.page} of {controller.lastQuestionPage} · {controller.questionTotal}
					questions
				{:else}
					Page 1 of 1 · 0 questions
				{/if}
			</span>
			<Pagination.Root
				class="w-auto"
				count={controller.questionTotal}
				perPage={QUESTION_PAGE_SIZE}
				page={controller.questionPage.paginator.page}
				onPageChange={(page) => controller.questionPage.paginator.setPage(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton
								disabled={controller.questionFetching || currentPage <= 1}
							/>
						</Pagination.Item>
						{#each pages as page (page.key)}
							<Pagination.Item>
								{#if page.type === "ellipsis"}
									<Pagination.Ellipsis />
								{:else}
									<Pagination.Link
										{page}
										disabled={controller.questionFetching}
										isActive={page.value === currentPage}
									/>
								{/if}
							</Pagination.Item>
						{/each}
						<Pagination.Item>
							<Pagination.NextButton
								disabled={controller.questionFetching ||
									currentPage >= controller.lastQuestionPage}
							/>
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</section>
