<script lang="ts">
	import type { InvestigationAttributes } from "$lib/api";
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Field from "$components/ui/field";
	import * as Pagination from "$components/ui/pagination";
	import { Spinner } from "$components/ui/spinner";
	import { Textarea } from "$components/ui/textarea";
	import { cn } from "$lib/utils";
	import { timestamp, investigationExecution } from "../model";
	import { useSituationInvestigationController } from "./controller.svelte";

	type Props = {
		investigationAttributes: InvestigationAttributes;
	};

	const { investigationAttributes: attrs }: Props = $props();
	const controller = useSituationInvestigationController();
	const execution = $derived(investigationExecution(attrs));
	const report = $derived(controller.report?.attributes);
	const reportPublishedAt = $derived(timestamp(report?.createdAt));
	const references = $derived(report?.references ?? []);
	const questionPagination = $derived(controller.questionQuery.data?.pagination);
	const questionTotal = $derived(questionPagination?.total ?? 0);
	const lastQuestionPage = $derived(Math.max(1, Math.ceil(questionTotal / 25)));
	const questionFetching = $derived(controller.questionQuery.isFetching);
	const canRetry = $derived(controller.retryAvailable);
</script>

<section aria-labelledby="original-question-title" class="flex flex-col gap-2">
	<h2 id="original-question-title" class="text-lg font-semibold">Original question</h2>
	<p class="max-w-[75ch] whitespace-pre-wrap text-sm leading-relaxed wrap-anywhere">
		{attrs.query || "Original question unavailable."}
	</p>
</section>

<section aria-labelledby="execution-status-title" class="flex flex-col gap-2">
	<div class="flex flex-wrap items-center gap-3">
		<h2 id="execution-status-title" class="text-lg font-semibold">Execution</h2>
		<Badge variant="secondary">{execution.label}</Badge>
	</div>
	{#if attrs.hasPendingWork}
		<p class="text-sm text-muted-foreground">Follow-up work is waiting.</p>
	{/if}
	{#if execution.message}
		<p class="text-sm text-muted-foreground">{execution.message}</p>
	{/if}
	{#if controller.retryMessage}
		<p role="status" class="text-sm text-muted-foreground">{controller.retryMessage}</p>
	{/if}
	{#if controller.retryError}
		<Alert.Root variant="destructive">
			<Alert.Title>Retry status needs attention</Alert.Title>
			<Alert.Description>{controller.retryError}</Alert.Description>
		</Alert.Root>
	{/if}
	{#if canRetry}
		{#if controller.retryNeedsRefresh}
			<Button
				variant="outline"
				size="sm"
				disabled={controller.retryReconciling}
				onclick={controller.refreshRetryStatus}
			>
				{#if controller.retryReconciling}
					<Spinner data-icon="inline-start" /> Refreshing status
				{:else}
					Refresh status
				{/if}
			</Button>
		{:else}
			<Button size="sm" disabled={controller.retryDisabled} onclick={controller.retryLatestFailedTurn}>
				{#if controller.retryMutation.isPending || controller.retryReconciling}
					<Spinner data-icon="inline-start" /> Retrying
				{:else}
					Retry failed turn
				{/if}
			</Button>
		{/if}
	{/if}
</section>

<section aria-labelledby="latest-report-title" class="flex flex-col gap-3">
	<h2 id="latest-report-title" class="text-lg font-semibold">Latest report</h2>
	{#if report}
		{#if controller.reportQuery.isError}
			<div role="status" class="flex items-center justify-between gap-3 text-xs text-muted-foreground">
				<span>Refresh failed. Showing the previously loaded report.</span>
				<Button variant="ghost" size="sm" onclick={() => controller.reportQuery.refetch()}>
					Retry
				</Button>
			</div>
		{/if}
		<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
			<time datetime={reportPublishedAt.iso}>Published {reportPublishedAt.label}</time>
			{#if report.turnStatus === "running"}
				<Badge variant="outline">In progress</Badge>
			{/if}
		</div>
		<p class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
			{report.text || "Report text unavailable."}
		</p>
		<div class="flex flex-col gap-2">
			<h3 class="text-sm font-medium">Evidence references</h3>
			{#if references.length}
				<ul class="flex flex-wrap gap-x-3 gap-y-1">
					{#each references as reference, index (reference.id)}
						<li>
							<Button
								variant="link"
								class="h-auto p-0"
								aria-label={`Inspect Evidence ${index + 1}`}
								onclick={(event) =>
									controller.openEvidence(reference.id, event.currentTarget)}
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
	{:else if controller.reportAccessLost}
		<p class="text-sm text-muted-foreground">The latest report is unavailable.</p>
	{:else if controller.reportQuery.isPending}
		<p role="status" class="flex items-center gap-2 text-sm text-muted-foreground">
			<Spinner /> Loading latest report
		</p>
	{:else if controller.reportQuery.isError && controller.reportQuery.error?.status !== 404}
		<div role="alert" class="flex flex-col items-start gap-3 text-sm">
			<p>The latest report could not be loaded.</p>
			<Button variant="outline" size="sm" onclick={() => controller.reportQuery.refetch()}>
				Retry
			</Button>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">No report was published yet.</p>
	{/if}
</section>

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
					{@const submittedAt = timestamp(row.input.attributes.createdAt)}
					<li
						class={cn(
							"flex flex-col gap-3 rounded-md border p-4",
							controller.submittedQuestion?.id === row.input.id && "border-primary bg-muted/40"
						)}
					>
						<div class="flex flex-wrap items-center justify-between gap-2">
							<Badge variant="secondary">{row.statusLabel}</Badge>
							<time datetime={submittedAt.iso} class="text-xs text-muted-foreground">
								Asked {submittedAt.label}
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
										<p role="status" class="text-sm text-muted-foreground">
											Answer unavailable.
										</p>
									{:else if row.answerPending}
										<p
											role="status"
											class="flex items-center gap-2 text-sm text-muted-foreground"
										>
											<Spinner /> Loading selected answer
										</p>
									{:else if row.answer}
										{@const answer = row.answer.attributes}
										{@const answerPublishedAt = timestamp(answer.createdAt)}
										{#if row.answerStale}
											<div
												role="status"
												class="flex items-center justify-between gap-3 text-xs text-muted-foreground"
											>
												<span>
													Refresh failed. Showing the previously loaded answer.
												</span>
												<Button
													variant="ghost"
													size="sm"
													onclick={() => row.answerQuery?.refetch()}
												>
													Retry
												</Button>
											</div>
										{/if}
										<div
											class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground"
										>
											<time datetime={answerPublishedAt.iso}>
												Published {answerPublishedAt.label}
											</time>
											<Badge
												variant={answer.turnStatus === "running"
													? "outline"
													: "secondary"}
											>
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
											<p class="text-sm text-muted-foreground">
												A later finding disputes this answer.
											</p>
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
																	controller.openEvidence(
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
												<p class="text-sm text-muted-foreground">
													No evidence references.
												</p>
											{/if}
										</div>
									{:else if row.answerLoadFailed}
										<div role="alert" class="flex flex-col items-start gap-2 text-sm">
											<p>The selected answer could not be loaded.</p>
											<Button
												variant="outline"
												size="sm"
												onclick={() => row.answerQuery?.refetch()}
											>
												Retry
											</Button>
										</div>
									{:else}
										<p role="status" class="text-sm text-muted-foreground">
											No current answer is selected.
										</p>
									{/if}
								</div>
							{/if}
						{/if}
					</li>
				{/each}
			</ol>
		{:else}
			<p class="text-sm text-muted-foreground">No questions have been submitted.</p>
		{/if}

		<div
			class="flex flex-wrap items-center justify-between gap-3 border-t pt-3 text-sm text-muted-foreground"
		>
			<span>
				{#if questionTotal}
					Page {controller.questionPage.paginator.page} of {lastQuestionPage} · {questionTotal} questions
				{:else}
					Page 1 of 1 · 0 questions
				{/if}
			</span>
			<Pagination.Root
				class="w-auto"
				count={questionTotal}
				perPage={25}
				page={controller.questionPage.paginator.page}
				onPageChange={(page) => controller.questionPage.paginator.setPage(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton disabled={questionFetching || currentPage <= 1} />
						</Pagination.Item>
						{#each pages as page (page.key)}
							<Pagination.Item>
								{#if page.type === "ellipsis"}
									<Pagination.Ellipsis />
								{:else}
									<Pagination.Link
										{page}
										disabled={questionFetching}
										isActive={page.value === currentPage}
									/>
								{/if}
							</Pagination.Item>
						{/each}
						<Pagination.Item>
							<Pagination.NextButton
								disabled={questionFetching || currentPage >= lastQuestionPage}
							/>
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</section>
