import { SvelteSet } from "svelte/reactivity";
import { Context, watch, type Getter } from "runed";
import {
	createInfiniteQuery,
	createMutation,
	createQueries,
	createQuery,
	useQueryClient,
} from "@tanstack/svelte-query";
import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { page } from "$app/state";
import { toast } from "svelte-sonner";
import {
	type AlertEpisode,
	clearSituationHoldMutation,
	clearSituationMuteMutation,
	closeSituationMutation,
	type ApiError,
	getIncidentOptions,
	type Incident,
	getSituationBriefOptions,
	getSituationOptions,
	getSituationQueryKey,
	getInvestigationOptions,
	getInvestigationReportOptions,
	listInvestigationUserInputsOptions,
	listSituationAlertEpisodesInfiniteOptions,
	listSituationAlertEpisodesQueryKey,
	listSituationsQueryKey,
	mergeSituationMutation,
	type GetSituationResponse,
	raiseSituationMutation,
	retryAgentTurnMutation,
	type SetSituationMuteAttributes,
	setSituationHoldMutation,
	setSituationMuteMutation,
	type Situation,
	updateInvestigationMutation,
} from "$lib/api";
import { getNextPageParam } from "$lib/api/utils";
import {
	isDefinitiveUnavailableError,
	isProvisionalReport,
	reportSummary,
	SITUATION_POLL_INTERVAL_MS,
	type ReportState,
} from "$features/situations/lib/model";
import { situationHref } from "$features/situations/lib/routes";
import { referenceUrl } from "$lib/reference-url";
import {
	activeHoldUntil,
	investigationRunStatus,
	investigationUpdateNotice,
	longRunningStatus,
} from "$features/situations/lib/status";
import type { BriefLoad } from "$features/situations/lib/brief";
import type { WhyOrigin } from "$features/situations/lib/why";
import { incidentResponseStatus } from "$features/incidents/lib/status";
import {
	declareSituationFrom,
	useIncidentCreateDialog,
} from "$features/incidents/components/create-incident-dialog";
import type { PageRelatedLink } from "$lib/app-shell.svelte";
import RiFireLine from "remixicon-svelte/icons/fire-line";

const idPath = (id?: string) => ({ id: id ?? "" });

/** The API's largest page; further pages load until every current episode is present. */
const EPISODES_PAGE_SIZE = 50;

export type MuteReason = SetSituationMuteAttributes["reason"];

/** Lifecycle actions available in the loaded situation's state. */
export type SituationActions = {
	raise: boolean;
	startInvestigation: boolean;
	mute: boolean;
	merge: boolean;
	keepOpen: boolean;
	release: boolean;
	unmute: boolean;
	close: boolean;
};

/**
 * What the page offers while there is no investigation: raise a candidate, start one for a raised
 * situation, or nothing while muted or closed.
 */
export type InvestigationOffer = "raise" | "start" | "muted" | "closed";

export class SituationController {
	private queryClient = useQueryClient();
	private incidentDialog = useIncidentCreateDialog();

	/** The URL value: the situation's reference, or another value the API resolves, such as its UUID. */
	ref = $state("");

	constructor(refFn: Getter<string>) {
		this.ref = refFn();
		watch(refFn, (ref) => {
			this.ref = ref;
		});
		watch(
			() => this.situationQuery.data,
			(response) => {
				if (response) {
					this.useReferenceUrl(response);
				}
			}
		);
		// ?declare=incident, the target of the notice's button, opens the declare dialog for this situation once
		// it has loaded at its reference URL, so switching the address doesn't bring the parameter back. The
		// parameter is cleared straight away, so going back or a refetch doesn't reopen it.
		watch(
			() => [this.declareRequested, this.reference === this.ref] as const,
			([requested, atReferenceUrl]) => {
				if (!requested || !atReferenceUrl) {
					return;
				}
				if (this.canDeclareIncident) {
					this.declareIncident();
				}
				this.incidentDialog.clearDeclareParam();
			}
		);
		watch(
			() => [this.episodesQuery.hasNextPage, this.episodesQuery.isFetching] as const,
			([hasNextPage, isFetching]) => {
				if (hasNextPage && !isFetching && !this.episodesQuery.isError) {
					this.episodesQuery.fetchNextPage();
				}
			}
		);
		// The brief is not polled: each load looks up changes per service. It reloads when what it reflects may
		// have changed: the situation, the latest report or the investigation's latest turn.
		watch(
			() =>
				[
					this.situation?.attributes.updatedAt,
					this.report?.id,
					this.investigationAttributes?.latestTurn?.id,
				] as const,
			() => {
				if (this.briefQuery.isFetched) {
					this.briefQuery.refetch();
				}
			},
			{ lazy: true }
		);
		watch(
			() => this.investigationId,
			() => this.resetRunActions(),
			{ lazy: true }
		);
		// Refresh with each fetch, and again when a hold deadline passes so Release turns back into Keep open.
		watch(
			() => [this.situation?.attributes.holdUntil, this.situationQuery.dataUpdatedAt] as const,
			([holdUntil]) => {
				this.now = Date.now();
				const remaining = Date.parse(holdUntil ?? "") - this.now;
				if (!(remaining > 0)) {
					return;
				}
				const timer = setTimeout(() => {
					this.now = Date.now();
				}, remaining + 1);
				return () => clearTimeout(timer);
			}
		);
	}

	situationQuery = createQuery(() => ({
		...getSituationOptions({ path: { id: this.ref } }),
		enabled: !!this.ref,
		refetchInterval: SITUATION_POLL_INTERVAL_MS,
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	situation = $derived(
		isDefinitiveUnavailableError(this.situationQuery.error) ? undefined : this.situationQuery.data?.data
	);
	situationUnavailable = $derived(isDefinitiveUnavailableError(this.situationQuery.error));
	/** The loaded situation's UUID, which every other request uses. */
	situationId = $derived(this.situation?.id);
	reference = $derived(this.situation?.attributes.reference);
	/** The value links to this situation use: its reference once loaded, until then the URL value. */
	linkRef = $derived(this.reference ?? this.ref);
	/** Presentation time for holds and the long-running marker; never used for detection. */
	now = $state(Date.now());

	/** Every current episode: all pages load, and a poll refetches each loaded page with the situation. */
	episodesQuery = createInfiniteQuery(() => ({
		...listSituationAlertEpisodesInfiniteOptions({
			path: idPath(this.situationId),
			query: { pageSize: EPISODES_PAGE_SIZE },
		}),
		enabled: !!this.situationId,
		initialPageParam: 1,
		getNextPageParam,
		refetchInterval: SITUATION_POLL_INTERVAL_MS,
		refetchIntervalInBackground: false,
	}));
	episodesByEntity = $derived(this.getEpisodesByEntity());
	/** True until every page has loaded, unless loading a page failed. */
	originLoading = $derived(
		this.episodesQuery.isPending || (this.episodesQuery.hasNextPage && !this.episodesQuery.isError)
	);
	origin = $derived(this.getOrigin());

	incidentIds = $derived([...new SvelteSet(this.situation?.attributes.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	/** Loaded linked incidents in `linkedIncidentIds` order; loading or failed ones are omitted. */
	linkedIncidents = $derived(this.getLinkedIncidents());
	linkedIncidentsLoading = $derived(this.incidentsQuery.some((query) => query.isPending && !query.data));
	relatedLinks = $derived<PageRelatedLink[]>(
		this.linkedIncidents.map((incident) => ({
			key: incident.id,
			kind: "Incident",
			icon: RiFireLine,
			label: incident.attributes.title,
			path: resolve("/incidents/[ref]/[[view=incidentView]]", { ref: incident.attributes.reference }),
			status: incidentResponseStatus(incident.attributes.responseState),
		}))
	);

	private declareRequested = $derived(page.url.searchParams.get("declare") === "incident");
	/** Open and closed situations can both start an incident. */
	canDeclareIncident = $derived(this.incidentDialog.enabled);

	private declareIncident() {
		if (!this.situation) {
			return;
		}
		this.incidentDialog.openCreate({ situation: declareSituationFrom(this.situation) });
	}

	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigationId = $derived(this.situationInvestigation?.investigation.id);
	investigationHref = $derived(situationHref(this.linkRef, "investigation", page.url.search));

	investigationQuery = createQuery(() => ({
		...getInvestigationOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			return isDefinitiveUnavailableError(query.state.error) ? false : SITUATION_POLL_INTERVAL_MS;
		},
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	investigation = $derived(
		isDefinitiveUnavailableError(this.investigationQuery.error)
			? undefined
			: this.investigationQuery.data?.data
	);
	investigationUnavailable = $derived(isDefinitiveUnavailableError(this.investigationQuery.error));
	investigationAttributes = $derived(this.investigation?.attributes);

	reportQuery = createQuery(() => ({
		...getInvestigationReportOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			if ([401, 403].includes(query.state.error?.status ?? 0) || this.investigationUnavailable) {
				return false;
			}
			return SITUATION_POLL_INTERVAL_MS;
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

	runStatus = $derived(investigationRunStatus(this.investigationAttributes));
	updateNotice = $derived(investigationUpdateNotice(this.investigationAttributes));
	reportAttributes = $derived(this.report?.attributes);
	reportState = $derived(this.getReportState());

	briefQuery = createQuery(() => ({
		...getSituationBriefOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
		refetchOnWindowFocus: true,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));
	briefLoad = $derived(this.getBriefLoad());

	stage = $derived(this.situation?.attributes.stage);
	holdUntil = $derived(this.situation ? activeHoldUntil(this.situation.attributes, this.now) : undefined);
	longRunning = $derived(
		this.situation ? longRunningStatus(this.situation.attributes, this.now) : undefined
	);
	actions = $derived(this.getActions());
	investigationOffer = $derived(this.getInvestigationOffer());

	raiseMutation = createMutation(() => ({
		...raiseSituationMutation(),
		onSuccess: async (response, variables) => {
			await this.applySituation(response.data);
			if (this.situationId === variables.path.id) {
				await goto(this.investigationHref);
			}
		},
		onError: (error) => this.notifyError("Could not raise the situation", error),
	}));

	muteMutation = createMutation(() => ({
		...setSituationMuteMutation(),
		onSuccess: async (response) => {
			this.muteDialogOpen = false;
			await this.applySituation(response.data);
		},
		onError: (error) => this.notifyError("Could not mute the situation", error),
	}));

	unmuteMutation = createMutation(() => ({
		...clearSituationMuteMutation(),
		onSuccess: (response) => this.applySituation(response.data),
		onError: (error) => this.notifyError("Could not unmute the situation", error),
	}));

	holdMutation = createMutation(() => ({
		...setSituationHoldMutation(),
		onSuccess: (response) => this.applySituation(response.data),
		onError: (error) => this.notifyError("Could not keep the situation open", error),
	}));

	releaseMutation = createMutation(() => ({
		...clearSituationHoldMutation(),
		onSuccess: (response) => this.applySituation(response.data),
		onError: (error) => this.notifyError("Could not release the hold", error),
	}));

	closeMutation = createMutation(() => ({
		...closeSituationMutation(),
		onSuccess: async (response) => {
			this.closeDialogOpen = false;
			await this.applySituation(response.data);
		},
		onError: (error) => this.notifyError("Could not close the situation", error),
	}));

	mergeMutation = createMutation(() => ({
		...mergeSituationMutation(),
		onSuccess: async (response, variables) => {
			this.mergeDialogOpen = false;
			const sourceId = variables.path.id;
			const targetId = response.data.id;
			await Promise.all([
				// The ID-less key matches both situations: this page caches the source under its URL value, and
				// other pages cache a situation under its UUID.
				this.queryClient.invalidateQueries({
					queryKey: getSituationQueryKey({} as Parameters<typeof getSituationQueryKey>[0]),
				}),
				this.queryClient.invalidateQueries({ queryKey: listSituationsQueryKey() }),
				...[sourceId, targetId].map((id) =>
					this.queryClient.invalidateQueries({
						queryKey: listSituationAlertEpisodesQueryKey({ path: { id } }),
					})
				),
			]);
			if (this.situationId === sourceId) {
				await goto(situationHref(targetId));
			}
		},
		onError: (error) => this.notifyError("Could not merge the situation", error),
	}));

	updateMutation = createMutation(() => updateInvestigationMutation());
	retryMutation = createMutation(() => retryAgentTurnMutation());

	updateError = $state("");
	retryReconciling = $state(false);
	retryNeedsRefresh = $state(false);
	retryMessage = $state("");
	retryError = $state("");

	retryAvailable = $derived(
		this.investigationAttributes?.latestTurn?.status === "failed" &&
			!this.investigationAttributes.activeTurn
	);
	retryDisabled = $derived(this.retryMutation.isPending || this.retryReconciling || this.retryNeedsRefresh);

	private mutations = $derived([
		this.raiseMutation,
		this.muteMutation,
		this.unmuteMutation,
		this.holdMutation,
		this.releaseMutation,
		this.closeMutation,
		this.mergeMutation,
	]);
	actionPending = $derived(this.mutations.some((mutation) => mutation.isPending));

	muteDialogOpen = $state(false);
	muteReason = $state<MuteReason>("not_noteworthy");
	mergeDialogOpen = $state(false);
	mergeTargetId = $state<string>();
	closeDialogOpen = $state(false);

	/** Raises a candidate, or starts the missing investigation of a raised situation. */
	raise = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.raiseMutation.mutate({ path: { id } });
	};

	openMuteDialog = () => {
		this.muteReason = "not_noteworthy";
		this.muteDialogOpen = true;
	};

	setMuteReason = (reason: string) => {
		this.muteReason = reason as MuteReason;
	};

	submitMute = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.muteMutation.mutate({ path: { id }, body: { attributes: { reason: this.muteReason } } });
	};

	unmute = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.unmuteMutation.mutate({ path: { id } });
	};

	/** Holds the situation open for the server's default length. */
	keepOpen = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.holdMutation.mutate({ path: { id }, body: { attributes: {} } });
	};

	release = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.releaseMutation.mutate({ path: { id } });
	};

	openCloseDialog = () => {
		this.closeDialogOpen = true;
	};

	confirmClose = () => {
		const id = this.situationId;
		if (!id) {
			return;
		}
		this.closeMutation.mutate({ path: { id }, body: { attributes: {} } });
	};

	openMergeDialog = () => {
		this.mergeTargetId = undefined;
		this.mergeDialogOpen = true;
	};

	setMergeTarget = (id: string) => {
		this.mergeTargetId = id;
	};

	submitMerge = () => {
		const id = this.situationId;
		const targetId = this.mergeTargetId;
		if (!id || !targetId) {
			return;
		}
		this.mergeMutation.mutate({ path: { id }, body: { attributes: { targetId } } });
	};

	retryInvestigation = () => {
		this.investigationQuery.refetch();
	};

	retryReport = () => {
		this.reportQuery.refetch();
	};

	retryBrief = () => {
		this.briefQuery.refetch();
	};

	/** Starts one turn for pending evidence, even when automatic updates are paused. */
	updateNow = async () => {
		if (!this.investigationId || this.updateMutation.isPending) {
			return;
		}

		this.updateError = "";
		const path = idPath(this.investigationId);
		try {
			const updated = await this.updateMutation.mutateAsync({ path });
			this.queryClient.setQueryData(getInvestigationOptions({ path }).queryKey, updated);
		} catch (error) {
			this.updateError = (error as ApiError).detail ?? "The update could not be started.";
		}
		await this.refreshInvestigation();
	};

	retryLatestFailedTurn = async () => {
		const latestTurn = this.investigationAttributes?.latestTurn;
		if (!this.retryAvailable || this.retryDisabled || !latestTurn) {
			return;
		}

		this.retryMessage = "";
		this.retryError = "";
		let accepted = false;
		let conflict = false;
		try {
			await this.retryMutation.mutateAsync({ path: { id: latestTurn.id } });
			accepted = true;
			this.retryMessage = "Retry requested for the same turn.";
		} catch (error) {
			const apiError = error as ApiError;
			conflict = apiError.code === "conflict";
			this.retryNeedsRefresh = true;
			if (conflict) {
				this.retryError =
					"Execution changed while retrying. Current investigation state is being refreshed.";
			} else {
				const message = apiError.detail ?? "The retry request could not be confirmed.";
				this.retryError = `${message} Current state is being refreshed before another retry is offered.`;
			}
		}

		const refreshed = await this.reconcileRetryStatus();
		if (!refreshed) {
			if (accepted) {
				this.retryError =
					"Retry was accepted, but the current state could not be confirmed. Refresh status to continue.";
			} else {
				this.retryError =
					"Current investigation state could not be refreshed. Refresh status before retrying again.";
			}
			return;
		}

		if (accepted) {
			this.retryError = "";
		} else if (conflict) {
			this.retryError =
				"Execution changed while retrying. Current investigation state has been refreshed.";
		} else {
			this.retryError = "Retry response was uncertain. Current investigation state has been refreshed.";
		}
	};

	refreshRetryStatus = async () => {
		if (this.retryReconciling) {
			return;
		}

		const refreshed = await this.reconcileRetryStatus();
		if (refreshed) {
			this.retryError = "";
		} else {
			this.retryError = "Investigation status could not be refreshed. Try again.";
		}
	};

	/** Refreshes the investigation, its report and its questions after a run action. */
	private async refreshInvestigation() {
		if (!this.investigationId) {
			return;
		}
		const path = idPath(this.investigationId);
		await Promise.all([
			this.queryClient.invalidateQueries({ queryKey: getInvestigationOptions({ path }).queryKey }),
			this.queryClient.invalidateQueries({
				queryKey: getInvestigationReportOptions({ path }).queryKey,
			}),
			this.queryClient.invalidateQueries({
				queryKey: listInvestigationUserInputsOptions({ path, query: {} }).queryKey,
			}),
		]);
	}

	private async reconcileRetryStatus() {
		this.retryReconciling = true;
		try {
			await this.refreshInvestigation();
			this.retryNeedsRefresh = this.investigationQuery.isError;
		} catch {
			this.retryNeedsRefresh = true;
		} finally {
			this.retryReconciling = false;
		}
		return !this.retryNeedsRefresh;
	}

	private resetRunActions() {
		this.updateError = "";
		this.retryMessage = "";
		this.retryError = "";
		this.retryNeedsRefresh = false;
	}

	private getBriefLoad(): BriefLoad {
		const brief = this.briefQuery.data?.data;
		if (brief) {
			return { kind: "loaded", brief };
		}
		if (this.briefQuery.isError) {
			return { kind: "error" };
		}
		return { kind: "loading" };
	}

	/** Stores the returned situation under this page's key, then refreshes every list and the Watching count. */
	async applySituation(situation: Situation) {
		const queryKey = getSituationQueryKey({ path: { id: this.ref } });
		this.queryClient.setQueryData(queryKey, { data: situation });
		await this.queryClient.invalidateQueries({ queryKey: listSituationsQueryKey() });
	}

	/** Replaces a URL that names the situation another way with its reference URL, caching the response first. */
	private useReferenceUrl(response: GetSituationResponse) {
		const reference = response.data.attributes.reference;
		if (reference === this.ref) {
			return;
		}
		this.queryClient.setQueryData(getSituationQueryKey({ path: { id: reference } }), response);
		void goto(referenceUrl(page.url, this.ref, reference), {
			replaceState: true,
			noScroll: true,
			keepFocus: true,
		});
	}

	private notifyError(title: string, error: ApiError) {
		toast.error(title, { description: error.detail });
	}

	private getActions(): SituationActions {
		const attributes = this.situation?.attributes;
		const open = !!attributes && this.stage !== "closed";
		const muted = open && !!attributes?.mutedAt;
		const active = open && !muted;
		const raised = this.stage === "raised";

		return {
			raise: active && !raised,
			startInvestigation: active && raised && !this.investigationId,
			mute: active,
			merge: active,
			keepOpen: active && raised && !this.holdUntil,
			release: active && raised && !!this.holdUntil,
			unmute: muted,
			close: open && raised,
		};
	}

	private getInvestigationOffer(): InvestigationOffer | undefined {
		if (!this.situation || this.investigationId) {
			return undefined;
		}
		if (this.stage === "closed") {
			return "closed";
		}
		if (this.situation.attributes.mutedAt) {
			return "muted";
		}
		if (this.stage === "raised") {
			return "start";
		}
		return "raise";
	}

	private getEpisodesByEntity() {
		const episodes = new Map<string, AlertEpisode>();
		const loaded = this.episodesQuery.data?.pages.flatMap((page) => page.data) ?? [];
		for (const episode of loaded) {
			const entityId = episode.attributes.knowledgeEntityId;
			if (entityId) {
				episodes.set(entityId, episode);
			}
		}
		return episodes;
	}

	/** The seed's details, when one of the situation's current episodes is the seed. */
	private getOrigin(): WhyOrigin | undefined {
		const seedEntityId = this.situation?.attributes.seedEntityId;
		const seed = seedEntityId ? this.episodesByEntity.get(seedEntityId) : undefined;
		const title = seed?.attributes.definition?.attributes.title;
		if (!seed || !title) {
			return undefined;
		}
		return { title, startedAt: seed.attributes.startedAt };
	}

	private getReportState(): ReportState {
		const report = this.reportAttributes;
		if (report) {
			return {
				kind: "published",
				attributes: report,
				summary: reportSummary(report),
				provisional: isProvisionalReport(report),
				refreshFailed: this.reportQuery.isError,
			};
		}
		if (this.reportAccessLost) {
			return { kind: "unavailable" };
		}
		if (this.reportQuery.isPending) {
			return { kind: "loading" };
		}
		if (this.reportQuery.isError && this.reportQuery.error?.status !== 404) {
			return { kind: "error" };
		}
		return { kind: "none" };
	}

	private getLinkedIncidents() {
		const incidents: Incident[] = [];
		for (const query of this.incidentsQuery) {
			const incident = query.data?.data;
			if (incident) {
				incidents.push(incident);
			}
		}
		return incidents;
	}
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (refFn: Getter<string>) => ctx.set(new SituationController(refFn));
export const useSituationController = () => ctx.get();
