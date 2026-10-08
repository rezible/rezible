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
	getSituationOptions,
	getSituationQueryKey,
	getInvestigationOptions,
	getInvestigationReportOptions,
	listSituationAlertEpisodesInfiniteOptions,
	listSituationAlertEpisodesQueryKey,
	listSituationsQueryKey,
	mergeSituationMutation,
	raiseSituationMutation,
	type SetSituationMuteAttributes,
	setSituationHoldMutation,
	setSituationMuteMutation,
	type Situation,
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
import { activeHoldUntil, investigationRunStatus, longRunningStatus } from "$features/situations/lib/status";
import type { WhyOrigin } from "$features/situations/lib/why";
import { incidentResponseStatus } from "$features/incidents/lib/status";
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
	situationId = $state<string>();

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
		watch(
			() => [this.episodesQuery.hasNextPage, this.episodesQuery.isFetching] as const,
			([hasNextPage, isFetching]) => {
				if (hasNextPage && !isFetching && !this.episodesQuery.isError) {
					this.episodesQuery.fetchNextPage();
				}
			}
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
		...getSituationOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
		refetchInterval: SITUATION_POLL_INTERVAL_MS,
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	situation = $derived(
		isDefinitiveUnavailableError(this.situationQuery.error) ? undefined : this.situationQuery.data?.data
	);
	situationUnavailable = $derived(isDefinitiveUnavailableError(this.situationQuery.error));
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
			path: resolve("/incidents/[slug]/[[view=incidentView]]", { slug: incident.attributes.slug }),
			status: incidentResponseStatus(incident.attributes.responseState),
		}))
	);

	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigationId = $derived(this.situationInvestigation?.investigation.id);
	investigationHref = $derived(situationHref(this.situationId ?? "", "investigation", page.url.search));

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
	reportAttributes = $derived(this.report?.attributes);
	reportState = $derived(this.getReportState());

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
			await this.applySituation(response.data);
			await Promise.all(
				[sourceId, targetId].flatMap((id) => [
					this.queryClient.invalidateQueries({ queryKey: getSituationQueryKey({ path: { id } }) }),
					this.queryClient.invalidateQueries({
						queryKey: listSituationAlertEpisodesQueryKey({ path: { id } }),
					}),
				])
			);
			if (this.situationId === sourceId) {
				await goto(situationHref(targetId));
			}
		},
		onError: (error) => this.notifyError("Could not merge the situation", error),
	}));

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

	/** Stores the returned situation under its own ID, then refreshes every list and the Watching count. */
	private async applySituation(situation: Situation) {
		const queryKey = getSituationQueryKey({ path: { id: situation.id } });
		this.queryClient.setQueryData(queryKey, { data: situation });
		await this.queryClient.invalidateQueries({ queryKey: listSituationsQueryKey() });
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
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
