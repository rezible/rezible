import { createInfiniteQuery, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import {
	getAgentSessionOptions,
	listAgentArtifactsInfiniteOptions,
	listAgentMessagesInfiniteOptions,
	listAgentTurnsInfiniteOptions,
	streamAgentSessionEvents,
	type AgentTurnUpdatedEvent,
	type StreamAgentSessionEventsResponse,
} from "$lib/api";
import { getNextPageParam } from "$lib/api/utils";
import { flattenPages } from "$src/components/system-analysis/lib";
import {
	applyArtifactChunk,
	applyModelChunk,
	emptyOverlay,
	eventMatchesSession,
	groupTranscript,
	mergedArtifacts,
	reconcileTurn,
} from "./model";
import { onMount } from "svelte";

const terminalStatuses = new Set(["completed", "failed", "aborted"]);

type hasSequenceAttribute = { attributes: { sequence: number } };
const sortBySequence = (a: hasSequenceAttribute, b: hasSequenceAttribute) =>
	a.attributes.sequence - b.attributes.sequence;

const hasNextPage = (q: { hasNextPage: boolean }) => q.hasNextPage;

export class SessionDetailController {
	sessionId = $state.raw("");
	connection = $state<"connecting" | "connected" | "reconnecting" | "disconnected">("connecting");
	overlay = $state.raw(emptyOverlay());
	private abort?: AbortController;
	private queryClient = useQueryClient();

	private hasSessionId = $derived(!!this.sessionId);

	sessionQuery = createQuery(() => ({
		...getAgentSessionOptions({
			path: { id: this.sessionId },
		}),
		enabled: this.hasSessionId,
	}));

	session = $derived(this.sessionQuery.data?.data);

	turnsQuery = createInfiniteQuery(() => ({
		...listAgentTurnsInfiniteOptions({
			path: { id: this.sessionId },
			query: { pageSize: 50 },
		}),
		enabled: this.hasSessionId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	private turnsQueryPages = $derived(this.turnsQuery.data?.pages);
	turns = $derived(flattenPages(this.turnsQueryPages).sort(sortBySequence));
	latestTurn = $derived(this.turns.at(-1));

	messagesQuery = createInfiniteQuery(() => ({
		...listAgentMessagesInfiniteOptions({
			path: { id: this.sessionId },
			query: { pageSize: 50 },
		}),
		enabled: this.hasSessionId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	private messagesQueryPages = $derived(this.messagesQuery.data?.pages);
	messages = $derived(flattenPages(this.messagesQueryPages).sort(sortBySequence));

	artifactsQuery = createInfiniteQuery(() => ({
		...listAgentArtifactsInfiniteOptions({
			path: { id: this.sessionId },
			query: { pageSize: 50 },
		}),
		enabled: this.hasSessionId,
		initialPageParam: 1,
		getNextPageParam,
	}));
	private artifactsQueryPages = $derived(this.artifactsQuery.data?.pages);

	persistedArtifacts = $derived(flattenPages(this.artifactsQueryPages));
	artifacts = $derived(mergedArtifacts(this.persistedArtifacts, this.overlay));

	transcript = $derived(groupTranscript(this.turns, this.messages));

	historyIncomplete = $derived(
		this.turnsQuery.hasNextPage || this.messagesQuery.hasNextPage || this.artifactsQuery.hasNextPage
	);
	isPending = $derived(
		this.sessionQuery.isPending ||
			this.turnsQuery.isPending ||
			this.messagesQuery.isPending ||
			this.artifactsQuery.isPending
	);
	error = $derived(
		this.sessionQuery.error ??
			this.turnsQuery.error ??
			this.messagesQuery.error ??
			this.artifactsQuery.error
	);

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.openSession(id);
		});
		onMount(() => {
			return () => this.cleanup();
		});
	}

	private cleanup() {
		this.abort?.abort();
	}

	retry = () => {
		this.sessionQuery.refetch();
		this.refetchAllQueryPages();
		void this.connect(this.sessionId);
	};

	private openSession(id: string) {
		this.overlay = emptyOverlay();
		void this.connect(id);
	}

	private async refetchAllQueryPages() {
		await Promise.all([
			this.turnsQuery.refetch(),
			this.messagesQuery.refetch(),
			this.artifactsQuery.refetch(),
		]);

		while (
			this.turnsQuery.hasNextPage ||
			this.messagesQuery.hasNextPage ||
			this.artifactsQuery.hasNextPage
		) {
			await Promise.all([
				this.turnsQuery.hasNextPage ? this.turnsQuery.fetchNextPage() : Promise.resolve(),
				this.messagesQuery.hasNextPage ? this.messagesQuery.fetchNextPage() : Promise.resolve(),
				this.artifactsQuery.hasNextPage ? this.artifactsQuery.fetchNextPage() : Promise.resolve(),
			]);
		}
	}

	private async onTurnUpdated(event: AgentTurnUpdatedEvent) {
		if (!eventMatchesSession(this.sessionId, event)) {
			return;
		}

		const isTerminalEvent = terminalStatuses.has(event.status);

		if (isTerminalEvent) {
			const newTurns = new Set(this.overlay.reconcilingTurns).add(event.turnId);
			this.overlay = { ...this.overlay, reconcilingTurns: newTurns };
		}

		await Promise.all([this.sessionQuery.refetch(), this.refetchAllQueryPages()]);

		if (isTerminalEvent) {
			this.overlay = reconcileTurn(this.overlay, event.turnId);
			const analysisId = this.session?.attributes.systemAnalysisId;
			if (analysisId)
				await this.queryClient.invalidateQueries({
					predicate: (query) => JSON.stringify(query.queryKey).includes(analysisId),
				});
		}
	}

	private async onStreamResponse(events: StreamAgentSessionEventsResponse) {
		for await (const { event, data } of events) {
			if (!eventMatchesSession(this.sessionId, data)) {
				continue;
			}

			this.connection = "connected";

			if (event === "turn-updated") {
				await this.onTurnUpdated(data);
			} else if (event === "turn-chunk") {
				if (data.model) {
					this.overlay = applyModelChunk(this.overlay, data.turnId, data.model);
				} else if (data.artifact) {
					this.overlay = applyArtifactChunk(this.overlay, data.artifact);
				} else {
					console.log("unhandled turn chunk event");
				}
			}
		}
	}

	private async startEventStream(id: string, abort: AbortController) {
		this.connection = "connecting";

		while (!abort.signal.aborted) {
			try {
				await streamAgentSessionEvents({
					path: { id },
					signal: abort.signal,
					onSseError: () => {
						this.connection = "reconnecting";
						void this.refetchAllQueryPages();
					},
					onSseEvent: (event) => {
						this.onStreamResponse(event.data);
					},
				});
				this.connection = "connected";
			} catch {
				if (abort.signal.aborted) break;
			}

			if (abort.signal.aborted) break;

			this.connection = "reconnecting";
			await new Promise((resolve) => setTimeout(resolve, 3000));
		}

		if (!!this && id === this.sessionId) {
			this.connection = "disconnected";
		}
	}

	private async connect(id: string) {
		this.abort?.abort();
		if (!id) {
			this.connection = "disconnected";
			return;
		}
		this.sessionId = id;
		this.abort = new AbortController();
		this.startEventStream(id, this.abort);
	}
}

const context = new Context<SessionDetailController>("SessionDetailController");
export const initSessionDetailController = (getId: () => string) =>
	context.set(new SessionDetailController(getId));
export const useSessionDetailController = () => context.get();
