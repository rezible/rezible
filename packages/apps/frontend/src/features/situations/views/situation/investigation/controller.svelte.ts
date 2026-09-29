import { createMutation, createQueries, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { tick } from "svelte";
import { SvelteSet } from "svelte/reactivity";
import {
	getInvestigationFindingOptions,
	getInvestigationOptions,
	getInvestigationReportOptions,
	listInvestigationUserInputsOptions,
	listInvestigationUserInputsQueryKey,
	retryAgentTurnMutation,
	submitInvestigationUserInputMutation,
	type InvestigationFinding,
	type ListInvestigationUserInputsResponse,
	type ErrorModel,
	type InvestigationUserInput,
	type AgentTurnStatusOverview,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, watch, type Getter } from "runed";
import { investigationRefreshInterval, isDefinitiveUnavailableError, type SourceTarget } from "../model";

const idPath = (id?: string) => ({ id: id ?? "" });

function userInputStatusLabel(ts: AgentTurnStatusOverview) {
	switch (ts.status) {
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

function errorStatus(error: unknown) {
	if (error && typeof error === "object" && "status" in error && typeof error.status === "number") {
		return error.status;
	}
	return undefined;
}

function errorMessage(error: unknown, fallback: string) {
	if (error && typeof error === "object" && "detail" in error && typeof error.detail === "string") {
		return error.detail;
	}
	if (error instanceof Error && error.message) {
		return error.message;
	}
	return fallback;
}

export class SituationInvestigationController {
	investigationId = $state<string>();
	sourceTarget = $state<SourceTarget>();
	private sourceTrigger = $state<HTMLElement>();
	questionDraft = $state("");
	questionValidationError = $state(false);
	questionError = $state("");
	submittedQuestion = $state<InvestigationUserInput>();
	expandedQuestionIds = new SvelteSet<string>();
	private submissionAttempt = $state<{ text: string; key: string }>();
	private answerSelectionRefreshes = new SvelteSet<string>();

	constructor(idFn: Getter<string | undefined>) {
		watch(idFn, (id) => {
			if (this.investigationId !== id) {
				this.sourceTarget = undefined;
				this.sourceTrigger = undefined;
				this.questionDraft = "";
				this.questionValidationError = false;
				this.questionError = "";
				this.submissionAttempt = undefined;
				this.submittedQuestion = undefined;
				this.expandedQuestionIds.clear();
				this.answerSelectionRefreshes.clear();
				this.retryMessage = "";
				this.retryError = "";
				this.retryNeedsRefresh = false;
			}
			this.investigationId = id;
		});

		watch(
			() =>
				this.answerRows
					.filter((row) => row.answerNeedsUpdate)
					.map((row) => `${row.input.id}:${row.input.attributes.answerVersionId}`)
					.sort()
					.join("|"),
			() => {
				void this.refetchStaleAnswerSelection();
			}
		);
	}

	private id = $derived(this.investigationId ?? "");
	private queryClient = useQueryClient();

	investigationQuery = createQuery(() => ({
		...getInvestigationOptions({ path: idPath(this.id) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			if (isDefinitiveUnavailableError(query.state.error)) {
				return false;
			}
			if (query.state.error && !query.state.data) {
				return 30000;
			}
			return investigationRefreshInterval(query.state.data?.data.attributes);
		},
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	investigation = $derived(
		isDefinitiveUnavailableError(this.investigationQuery.error)
			? undefined
			: this.investigationQuery.data?.data
	);
	refreshInterval = $derived.by(() => {
		if (!this.investigationId || this.investigationUnavailable) {
			return 30000;
		}
		if (this.investigationQuery.isError && !this.investigation) {
			return 30000;
		}
		return investigationRefreshInterval(this.investigation?.attributes);
	});
	investigationUnavailable = $derived(isDefinitiveUnavailableError(this.investigationQuery.error));
	investigationAttributes = $derived(this.investigation?.attributes);

	reportQuery = createQuery(() => ({
		...getInvestigationReportOptions({ path: idPath(this.id) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			if ([401, 403].includes(query.state.error?.status ?? 0) || this.investigationUnavailable) {
				return false;
			}
			return this.refreshInterval;
		},
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	report = $derived(
		this.investigationUnavailable || isDefinitiveUnavailableError(this.reportQuery.error)
			? undefined
			: this.reportQuery.data?.data
	);
	reportAccessLost = $derived([401, 403].includes(this.reportQuery.error?.status ?? 0));

	questionPage = createPaginatedQuery<
		ListInvestigationUserInputsResponse,
		ListInvestigationUserInputsResponse,
		ReturnType<typeof listInvestigationUserInputsQueryKey>
	>({
		queryOptions: (pagination) => ({
			...listInvestigationUserInputsOptions({
				path: idPath(this.id),
				query: { ...pagination, pageSize: 25 },
			}),
			enabled: !!this.investigationId,
			refetchInterval: (query) => {
				if (isDefinitiveUnavailableError(query.state.error) || this.investigationUnavailable) {
					return false;
				}
				return this.refreshInterval;
			},
			refetchIntervalInBackground: false,
			retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
		}),
		resetWhen: () => this.investigationId,
		keepPreviousQueryData: false,
	});
	questionQuery = $derived(this.questionPage.query);
	questionInputsUnavailable = $derived(
		isDefinitiveUnavailableError(this.questionQuery.error) || this.investigationUnavailable
	);
	questionInputs = $derived(this.questionInputsUnavailable ? [] : (this.questionQuery.data?.data ?? []));
	questionSuccessMessage = $derived.by(() => {
		const input = this.submittedQuestion;
		if (!input) {
			return "";
		}

		const ts = input.attributes.agentTurn;
		const status = userInputStatusLabel(ts);
		if (ts.status === "completed" && !input.attributes.answerVersionId) {
			return `Question submitted. Status: ${status}. No answer was published.`;
		}
		if (input.attributes.answerVersionId) {
			return `Question submitted. Status: ${status}. An answer is available.`;
		}
		return `Question submitted. Status: ${status}.`;
	});
	expandedAnswerInputs = $derived(
		this.questionInputs.filter(
			(input) => this.expandedQuestionIds.has(input.id) && !!input.attributes.answerVersionId
		)
	);

	expandedAnswerQueries = createQueries(() => ({
		queries: this.expandedAnswerInputs.map((input) => {
			const versionId = input.attributes.answerVersionId ?? "";
			return {
				...getInvestigationFindingOptions({ path: { id: this.id, versionId } }),
				enabled: !!this.investigationId && !!versionId,
				refetchInterval: (query: { state: { error: ErrorModel | null } }) => {
					if (isDefinitiveUnavailableError(query.state.error) || this.investigationUnavailable) {
						return false;
					}
					return this.refreshInterval;
				},
				refetchIntervalInBackground: false,
				retry: (failureCount: number, error: ErrorModel) =>
					!isDefinitiveUnavailableError(error) && failureCount < 2,
			};
		}),
	}));

	answerRows = $derived.by(() => {
		let answerQueryIndex = 0;

		return this.questionInputs.map((input) => {
			const expanded = this.expandedQuestionIds.has(input.id);
			const versionId = input.attributes.answerVersionId;
			const answerQuery =
				expanded && versionId ? this.expandedAnswerQueries[answerQueryIndex++] : undefined;
			const queriedAnswer = answerQuery?.data?.data;
			const answerUnavailable = isDefinitiveUnavailableError(answerQuery?.error);
			const answerNeedsUpdate =
				!!queriedAnswer && !["running", "completed"].includes(queriedAnswer.attributes.turnStatus);
			const answer: InvestigationFinding | undefined =
				answerUnavailable || answerNeedsUpdate ? undefined : queriedAnswer;

			const ts = input.attributes.agentTurn;
			return {
				input,
				statusLabel: userInputStatusLabel(ts),
				noAnswerPublished: ts.status === "completed" && !input.attributes.answerVersionId,
				hasSelectedAnswer: !!versionId,
				expanded,
				answerQuery,
				answer,
				answerUnavailable,
				answerNeedsUpdate,
				answerStale: !!answer && !!answerQuery?.isError,
				answerPending: !!answerQuery?.isPending && !answerQuery.data,
				answerLoadFailed: !!answerQuery?.isError && !answerQuery.data && !answerUnavailable,
			};
		});
	});

	questionSubmitMutation = createMutation(() => submitInvestigationUserInputMutation());
	questionSubmitting = $derived(this.questionSubmitMutation.isPending);

	retryMutation = createMutation(() => retryAgentTurnMutation());
	retryReconciling = $state(false);
	retryNeedsRefresh = $state(false);
	retryMessage = $state("");
	retryError = $state("");
	retryAvailable = $derived(
		this.investigationAttributes?.latestTurn.status === "failed" &&
			!this.investigationAttributes.activeTurn &&
			!!this.investigationAttributes.latestTurn
	);
	retryDisabled = $derived(this.retryMutation.isPending || this.retryReconciling || this.retryNeedsRefresh);

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
			this.questionError = errorMessage(error, "The question could not be submitted.");
			return;
		}

		this.questionDraft = "";
		this.submissionAttempt = undefined;
		this.submittedQuestion = submitted;

		const refreshed = await this.refreshVisibleInvestigationData();
		const pagination = refreshed.data?.pagination;
		if (!refreshed.isError && pagination) {
			const lastPage = Math.max(1, Math.ceil(pagination.total / 25));
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
		const staleRows = this.answerRows.filter((row) => row.answerNeedsUpdate);
		let shouldRefetch = false;

		for (const row of staleRows) {
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

	retryLatestFailedTurn = async () => {
		const attributes = this.investigationAttributes;
		if (!this.retryAvailable || this.retryDisabled || !attributes?.latestTurn) {
			return;
		}

		this.retryMessage = "";
		this.retryError = "";
		try {
			await this.retryMutation.mutateAsync({ path: { id: attributes.latestTurn.id } });
		} catch (error) {
			const conflict = errorStatus(error) === 409;
			this.retryNeedsRefresh = true;
			this.retryReconciling = true;
			this.retryError = conflict
				? "Execution changed while retrying. Current investigation state is being refreshed."
				: `${errorMessage(error, "The retry request could not be confirmed.")} Current state is being refreshed before another retry is offered.`;
			try {
				await this.refreshVisibleInvestigationData();
				this.retryNeedsRefresh = this.investigationQuery.isError;
				if (this.retryNeedsRefresh) {
					this.retryError =
						"Current investigation state could not be refreshed. Refresh status before retrying again.";
				} else if (conflict) {
					this.retryError =
						"Execution changed while retrying. Current investigation state has been refreshed.";
				} else {
					this.retryError =
						"Retry response was uncertain. Current investigation state has been refreshed.";
				}
			} catch {
				this.retryNeedsRefresh = true;
				this.retryError =
					"Current investigation state could not be refreshed. Refresh status before retrying again.";
			} finally {
				this.retryReconciling = false;
			}
			return;
		}

		this.retryMessage = "Retry requested for the same turn.";
		this.retryReconciling = true;
		try {
			await this.refreshVisibleInvestigationData();
			this.retryNeedsRefresh = this.investigationQuery.isError;
			if (this.retryNeedsRefresh) {
				this.retryError =
					"Retry was accepted, but the current state could not be confirmed. Refresh status to continue.";
			}
		} catch {
			this.retryNeedsRefresh = true;
			this.retryError =
				"Retry was accepted, but the current state could not be confirmed. Refresh status to continue.";
		} finally {
			this.retryReconciling = false;
		}
	};

	refreshRetryStatus = async () => {
		if (this.retryReconciling) {
			return;
		}

		this.retryReconciling = true;
		try {
			await this.refreshVisibleInvestigationData();
			this.retryNeedsRefresh = this.investigationQuery.isError;
			if (!this.retryNeedsRefresh) {
				this.retryError = "";
			}
		} catch {
			this.retryNeedsRefresh = true;
			this.retryError = "Investigation status could not be refreshed. Try again.";
		} finally {
			this.retryReconciling = false;
		}
	};

	openSource = (target: SourceTarget, trigger: HTMLElement) => {
		this.sourceTarget = target;
		this.sourceTrigger = trigger;
	};

	openEvidence = (id: string, trigger: HTMLElement) => {
		this.openSource({ kind: "knowledgeEvidence", id }, trigger);
	};

	closeSource = async () => {
		const trigger = this.sourceTrigger;
		this.sourceTarget = undefined;
		this.sourceTrigger = undefined;
		await tick();
		if (trigger?.isConnected) {
			trigger.focus();
		}
	};
}

const ctx = new Context<SituationInvestigationController>("SituationInvestigationController");
export const initSituationInvestigationController = (idFn: Getter<string | undefined>) =>
	ctx.set(new SituationInvestigationController(idFn));
export const useSituationInvestigationController = () => ctx.get();
