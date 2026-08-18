import {
	getAgentSessionOptions,
	listAgentTurnsOptions,
	type AgentSession,
	type AgentTurn,
	type ErrorModel,
} from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { mockAgentSession, mockAgentTurns } from "./mock";

export class AgentSessionComponentController {
	sessionId = $state<string>(null!);
	isExpanded = $state(false);

	constructor(idFn: Getter<string>) {
		this.sessionId = idFn();
		watch(idFn, (id) => {
			this.sessionId = id;
		});
	}

	private isMock = $derived(this.sessionId === "mock");
	private sessionQuery = createQuery(() => ({
		...getAgentSessionOptions({ path: { id: this.sessionId } }),
		enabled: Boolean(this.sessionId) && !this.isMock,
	}));
	private turnsQuery = createQuery(() => ({
		...listAgentTurnsOptions({ path: { id: this.sessionId } }),
		enabled: Boolean(this.sessionId) && !this.isMock,
	}));

	session = $derived<AgentSession | undefined>(
		this.isMock ? mockAgentSession : this.sessionQuery.data?.data
	);
	turns = $derived<AgentTurn[]>(this.isMock ? mockAgentTurns : (this.turnsQuery.data?.data ?? []));
	latestTurn = $derived.by(() => {
		return this.turns.reduce<AgentTurn | undefined>((latest, turn) => {
			if (!latest) return turn;
			if (turn.attributes.sequence > latest.attributes.sequence) return turn;
			return new Date(turn.attributes.updatedAt) > new Date(latest.attributes.updatedAt)
				? turn
				: latest;
		}, undefined);
	});
	isLoading = $derived(
		this.sessionQuery.isLoading ||
			this.sessionQuery.isPending ||
			this.turnsQuery.isLoading ||
			this.turnsQuery.isPending
	);
	error = $derived((this.sessionQuery.error ?? this.turnsQuery.error) as ErrorModel | undefined);

	setExpanded(open: boolean) {
		this.isExpanded = open;
	}
}

const ctx = new Context<AgentSessionComponentController>("AgentSessionComponentController");
export const initAgentSessionComponentController = (idFn: Getter<string>) =>
	ctx.set(new AgentSessionComponentController(idFn));
export const useAgentSessionComponentController = () => ctx.get();
