import {
	createMutation,
	createQueries,
	useQueryClient,
	type CreateQueryResult,
} from "@tanstack/svelte-query";
import { SvelteSet } from "svelte/reactivity";
import {
	getInvestigationFindingOptions,
	getInvestigationOptions,
	getInvestigationReportOptions,
	listInvestigationFindingsOptions,
	listInvestigationHypothesesOptions,
	listInvestigationUserInputsOptions,
	listInvestigationUserInputsQueryKey,
	submitInvestigationUserInputMutation,
	type InvestigationFinding,
	type ListInvestigationUserInputsResponse,
	type ApiError,
	type InvestigationUserInput,
	type AgentTurnStatusOverview,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, watch } from "runed";
import { SITUATION_POLL_INTERVAL_MS, isDefinitiveUnavailableError } from "$features/situations/lib/model";
import { formatTime } from "$lib/time";
import {
	buildCitationIndex,
	buildFindingViews,
	buildHypothesisViews,
	citationViews,
	referenceNumbers,
	findingRow,
	hypothesisRow,
	type CitationView,
	type OutputRow,
} from "$features/situations/lib/investigation-outputs";

import { useSituationController } from "../controller.svelte";
import { SourceInspection } from "$features/situations/lib/sourceInspection.svelte";

export const QUESTION_PAGE_SIZE = 25;
export const OUTPUT_PAGE_SIZE = 50;

const idPath = (id?: string) => ({ id: id ?? "" });

function userInputStatusLabel(turn: AgentTurnStatusOverview | null) {
	switch (turn?.status) {
		case "queued":
			return "Queued";
		case "running":
			return "Running";
		case "completed":
			return "Completed";
		case "failed":
			return "Failed";
		case "aborted":
			return "Aborted";
		default:
			return "Waiting";
	}
}

type OutputList = {
	name: string;
	label: string;
	query: CreateQueryResult<{ data: unknown[] }, ApiError>;
	rows: OutputRow[];
	total: number;
};

export class SituationInvestigationController {
	private pageController = useSituationController();
	private queryClient = useQueryClient();

	inspection = new SourceInspection();
	questionDraft = $state("");
	questionValidationError = $state(false);
	questionError = $state("");
	submittedQuestion = $state<InvestigationUserInput>();
	expandedQuestionIds = new SvelteSet<string>();
	private submissionAttempt = $state<{ text: string; key: string }>();
	private answerSelectionRefreshes = new SvelteSet<string>();

	questionSubmitMutation = createMutation(() => submitInvestigationUserInputMutation());

	questionPage = createPaginatedQuery<
		ListInvestigationUserInputsResponse,
		ListInvestigationUserInputsResponse,
		ReturnType<typeof listInvestigationUserInputsQueryKey>
	>({
		queryOptions: (pagination) => ({
			...listInvestigationUserInputsOptions({
				path: idPath(this.investigationId),
				query: { ...pagination, pageSize: QUESTION_PAGE_SIZE },
			}),
			enabled: !!this.investigationId,
			refetchInterval: (query) => {
				if (isDefinitiveUnavailableError(query.state.error) || this.investigationUnavailable) {
					return false;
				}
				return SITUATION_POLL_INTERVAL_MS;
			},
			refetchIntervalInBackground: false,
			retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
		}),
		resetWhen: () => this.investigationId,
		keepPreviousQueryData: false,
	});
	questionQuery = $derived(this.questionPage.query);

	private findingsPage = createPaginatedQuery({
		source: "local",
		queryOptions: () => ({
			...listInvestigationFindingsOptions({
				path: idPath(this.investigationId),
				query: { page: 1, pageSize: OUTPUT_PAGE_SIZE },
			}),
			enabled: !!this.investigationId,
			refetchInterval: (query: { state: { error: ApiError | null } }) =>
				this.outputRefetchInterval(query.state.error),
			refetchIntervalInBackground: false,
			retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
		}),
	});
	findingsQuery = $derived(this.findingsPage.query);

	private hypothesesPage = createPaginatedQuery({
		source: "local",
		queryOptions: () => ({
			...listInvestigationHypothesesOptions({
				path: idPath(this.investigationId),
				query: { page: 1, pageSize: OUTPUT_PAGE_SIZE },
			}),
			enabled: !!this.investigationId,
			refetchInterval: (query: { state: { error: ApiError | null } }) =>
				this.outputRefetchInterval(query.state.error),
			refetchIntervalInBackground: false,
			retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
		}),
	});
	hypothesesQuery = $derived(this.hypothesesPage.query);

	private findingItems = $derived(this.findingsQuery.data?.data ?? []);
	private hypothesisItems = $derived(this.hypothesesQuery.data?.data ?? []);
	private citationIndex = $derived(
		buildCitationIndex(this.reportAttributes, this.findingItems, this.hypothesisItems)
	);

	findings = $derived(buildFindingViews(this.findingItems, this.citationIndex));
	hypotheses = $derived(buildHypothesisViews(this.hypothesisItems, this.citationIndex));
	findingRows = $derived(this.findings.map(findingRow));
	hypothesisRows = $derived(this.hypotheses.map(hypothesisRow));
	private findingsTotal = $derived(this.findingsQuery.data?.pagination.total ?? 0);
	private hypothesesTotal = $derived(this.hypothesesQuery.data?.pagination.total ?? 0);
	reportCitations = $derived(citationViews(this.reportAttributes?.references ?? [], this.citationIndex));
	citedEvidenceCount = $derived(this.citationIndex.size);
	conclusion = $derived(this.getConclusion());
	expandedOutputIds = new SvelteSet<string>();
	questionTotal = $derived(this.questionQuery.data?.pagination.total ?? 0);
	lastQuestionPage = $derived(Math.max(1, Math.ceil(this.questionTotal / QUESTION_PAGE_SIZE)));
	questionFetching = $derived(this.questionQuery.isFetching);
	questionInputsUnavailable = $derived(
		isDefinitiveUnavailableError(this.questionQuery.error) || this.investigationUnavailable
	);
	questionInputs = $derived(this.questionInputsUnavailable ? [] : (this.questionQuery.data?.data ?? []));
	questionSuccessMessage = $derived(this.getQuestionSuccessMessage());

	expandedAnswerInputs = $derived(this.getExpandedAnswerInputs());

	expandedAnswerQueries = createQueries(() => ({
		queries: this.expandedAnswerInputs.map((input) => {
			const versionId = input.attributes.answerVersionId ?? "";
			return {
				...getInvestigationFindingOptions({ path: { id: this.investigationId ?? "", versionId } }),
				enabled: !!this.investigationId && !!versionId,
				refetchInterval: (query: { state: { error: ApiError | null } }) => {
					if (isDefinitiveUnavailableError(query.state.error) || this.investigationUnavailable) {
						return false;
					}
					return SITUATION_POLL_INTERVAL_MS;
				},
				refetchIntervalInBackground: false,
				retry: (failureCount: number, error: ApiError) =>
					!isDefinitiveUnavailableError(error) && failureCount < 2,
			};
		}),
	}));

	answerRows = $derived(this.buildAnswerRows());

	questionSubmitting = $derived(this.questionSubmitMutation.isPending);

	get investigationId() {
		return this.pageController.investigationId;
	}

	get investigationQuery() {
		return this.pageController.investigationQuery;
	}

	get investigationAttributes() {
		return this.pageController.investigationAttributes;
	}

	get investigationUnavailable() {
		return this.pageController.investigationUnavailable;
	}

	get run() {
		return this.pageController.runStatus;
	}

	get startPending() {
		return this.pageController.raiseMutation.isPending;
	}

	get actionPending() {
		return this.pageController.actionPending;
	}

	get investigationOffer() {
		return this.pageController.investigationOffer;
	}

	get reportAttributes() {
		return this.pageController.reportAttributes;
	}

	get reportState() {
		return this.pageController.reportState;
	}

	retryReport = () => {
		this.pageController.retryReport();
	};

	startInvestigation = () => {
		this.pageController.raise();
	};

	toggleOutput = (id: string) => {
		if (this.expandedOutputIds.has(id)) {
			this.expandedOutputIds.delete(id);
			return;
		}
		this.expandedOutputIds.add(id);
	};

	/** Everything a findings or hypotheses list needs to render. */
	outputList(kind: "findings" | "hypotheses"): OutputList {
		if (kind === "findings") {
			return {
				name: "findings",
				label: "Findings",
				query: this.findingsQuery,
				rows: this.findingRows,
				total: this.findingsTotal,
			};
		}
		return {
			name: "hypotheses",
			label: "Hypotheses",
			query: this.hypothesesQuery,
			rows: this.hypothesisRows,
			total: this.hypothesesTotal,
		};
	}

	openCitation = (citation: CitationView, trigger: HTMLElement) => {
		this.inspection.openEvidence(citation.id, trigger);
	};

	constructor() {
		watch(
			() => this.investigationId,
			() => this.resetInteractionState(),
			{ lazy: true }
		);
		watch(
			() => this.answerRows,
			() => {
				void this.refetchStaleAnswerSelection();
			}
		);
	}

	updateQuestionDraft = (text: string) => {
		if (this.submissionAttempt && this.submissionAttempt.text !== text) {
			this.submissionAttempt = undefined;
			this.questionError = "";
		}
		this.questionDraft = text;
		if (text.trim()) {
			this.questionValidationError = false;
		}
	};

	submitQuestion = async () => {
		if (this.questionSubmitting) {
			return;
		}

		const text = this.questionDraft;
		if (!text.trim()) {
			this.questionValidationError = true;
			return;
		}

		if (!this.investigationId) {
			return;
		}

		this.questionValidationError = false;
		this.questionError = "";
		this.submittedQuestion = undefined;
		if (!this.submissionAttempt || this.submissionAttempt.text !== text) {
			this.submissionAttempt = { text, key: crypto.randomUUID() };
		}

		let submitted: InvestigationUserInput;
		try {
			const response = await this.questionSubmitMutation.mutateAsync({
				path: { id: this.investigationId },
				body: { text, submissionKey: this.submissionAttempt.key },
			});
			submitted = response.data;
		} catch (error) {
			this.questionError = (error as ApiError).detail ?? "The question could not be submitted.";
			return;
		}

		this.questionDraft = "";
		this.submissionAttempt = undefined;
		this.submittedQuestion = submitted;

		const refreshed = await this.refreshVisibleInvestigationData();
		const pagination = refreshed.data?.pagination;
		if (!refreshed.isError && pagination) {
			const lastPage = Math.max(1, Math.ceil(pagination.total / QUESTION_PAGE_SIZE));
			this.questionPage.paginator.setPage(lastPage);
		}
	};

	toggleAnswer = (inputId: string) => {
		if (this.expandedQuestionIds.has(inputId)) {
			this.expandedQuestionIds.delete(inputId);
			return;
		}
		this.expandedQuestionIds.add(inputId);
	};

	private async refreshVisibleInvestigationData() {
		if (!this.investigationId) {
			return this.questionQuery.refetch();
		}

		const path = idPath(this.investigationId);
		const invalidations = [
			this.queryClient.invalidateQueries({
				queryKey: getInvestigationOptions({ path }).queryKey,
			}),
			this.queryClient.invalidateQueries({
				queryKey: getInvestigationReportOptions({ path }).queryKey,
			}),
			this.queryClient.invalidateQueries({
				queryKey: listInvestigationUserInputsOptions({ path, query: {} }).queryKey,
				refetchType: "none",
			}),
		];

		for (const input of this.expandedAnswerInputs) {
			const versionId = input.attributes.answerVersionId;
			if (versionId) {
				invalidations.push(
					this.queryClient.invalidateQueries({
						queryKey: getInvestigationFindingOptions({
							path: { id: this.investigationId, versionId },
						}).queryKey,
					})
				);
			}
		}

		await Promise.all(invalidations);
		return this.questionQuery.refetch();
	}

	private async refetchStaleAnswerSelection() {
		let shouldRefetch = false;

		for (const row of this.answerRows) {
			if (!row.answerNeedsUpdate) {
				continue;
			}
			const key = `${row.input.id}:${row.input.attributes.answerVersionId}`;
			if (!this.answerSelectionRefreshes.has(key)) {
				this.answerSelectionRefreshes.add(key);
				shouldRefetch = true;
			}
		}

		if (shouldRefetch) {
			await this.questionQuery.refetch();
		}
	}

	private getConclusion() {
		const report = this.pageController.reportState;
		if (report.kind === "published") {
			return report.summary;
		}
		return undefined;
	}

	private outputRefetchInterval(error: ApiError | null) {
		if (isDefinitiveUnavailableError(error) || this.investigationUnavailable) {
			return false;
		}
		return SITUATION_POLL_INTERVAL_MS;
	}

	private resetInteractionState() {
		this.expandedOutputIds.clear();
		this.inspection.reset();
		this.questionDraft = "";
		this.questionValidationError = false;
		this.questionError = "";
		this.submissionAttempt = undefined;
		this.submittedQuestion = undefined;
		this.expandedQuestionIds.clear();
		this.answerSelectionRefreshes.clear();
	}

	private getQuestionSuccessMessage() {
		const input = this.submittedQuestion;
		if (!input) {
			return "";
		}

		const turn = input.attributes.agentTurn;
		const status = userInputStatusLabel(turn);
		if (turn?.status === "completed" && !input.attributes.answerVersionId) {
			return `Question submitted. Status: ${status}. No answer was published.`;
		}
		if (input.attributes.answerVersionId) {
			return `Question submitted. Status: ${status}. An answer is available.`;
		}
		return `Question submitted. Status: ${status}.`;
	}

	private getExpandedAnswerInputs() {
		const inputs: InvestigationUserInput[] = [];
		for (const input of this.questionInputs) {
			if (this.expandedQuestionIds.has(input.id) && input.attributes.answerVersionId) {
				inputs.push(input);
			}
		}
		return inputs;
	}

	private buildAnswerRows() {
		const rows = [];
		let answerQueryIndex = 0;

		for (const input of this.questionInputs) {
			const expanded = this.expandedQuestionIds.has(input.id);
			const versionId = input.attributes.answerVersionId;
			let answerQuery: (typeof this.expandedAnswerQueries)[number] | undefined;
			if (expanded && versionId) {
				answerQuery = this.expandedAnswerQueries[answerQueryIndex];
				answerQueryIndex += 1;
			}
			const queriedAnswer = answerQuery?.data?.data;
			const answerUnavailable = isDefinitiveUnavailableError(answerQuery?.error);
			const answerNeedsUpdate =
				!!queriedAnswer && !["running", "completed"].includes(queriedAnswer.attributes.turnStatus);
			let answer: InvestigationFinding | undefined;
			if (!answerUnavailable && !answerNeedsUpdate) {
				answer = queriedAnswer;
			}

			const answerReferences = answer?.attributes.references ?? [];
			const answerNumbers = referenceNumbers(answerReferences);

			const turn = input.attributes.agentTurn;
			rows.push({
				input,
				submittedAt: formatTime(input.attributes.createdAt),
				answerPublishedAt: formatTime(answer?.attributes.createdAt),
				statusLabel: userInputStatusLabel(turn),
				noAnswerPublished: turn?.status === "completed" && !input.attributes.answerVersionId,
				hasSelectedAnswer: !!versionId,
				expanded,
				answerQuery,
				answer,
				answerCitations: citationViews(answerReferences, answerNumbers),
				answerUnavailable,
				answerNeedsUpdate,
				answerStale: !!answer && !!answerQuery?.isError,
				answerPending: !!answerQuery?.isPending && !answerQuery.data,
				answerLoadFailed: !!answerQuery?.isError && !answerQuery.data && !answerUnavailable,
			});
		}
		return rows;
	}
}

const ctx = new Context<SituationInvestigationController>("SituationInvestigationController");
export const initSituationInvestigationController = () => ctx.set(new SituationInvestigationController());
export const useSituationInvestigationController = () => ctx.get();

export type InvestigationQuestionRow = SituationInvestigationController["answerRows"][number];
