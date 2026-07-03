import { getAiAgentRunOptions, type AiAgentRun, type AiAgentRunSnapshot, type ErrorModel } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { mockAiAgentRun } from "./mock";

const sortedSnapshots = (snapshots?: AiAgentRunSnapshot[]) => {
    return (!snapshots) 
        ? []
        : snapshots.toSorted((a, b) => (Date.parse(b.attributes.created_at) - Date.parse(a.attributes.created_at)));
}

export class AiAgentRunComponentController {
	runId = $state<string>(null!);
	isExpanded = $state(false);

	constructor(idFn: Getter<string>) {
		this.runId = idFn();
		watch(idFn, (id) => {this.runId = id});
	}

    private isMock = $derived(this.runId === "mock");
	private agentRunQuery = createQuery(() => ({
		...getAiAgentRunOptions({ path: { id: this.runId } }),
		enabled: (!!this.runId && !this.isMock),
	}));

	agentRun = $derived<AiAgentRun | undefined>(this.isMock ? mockAiAgentRun : this.agentRunQuery.data?.data);
	isLoading = $derived(this.agentRunQuery.isLoading || this.agentRunQuery.isPending);
	error = $derived(this.agentRunQuery.error as ErrorModel | undefined);

	snapshots = $derived<AiAgentRunSnapshot[]>(sortedSnapshots(this.agentRun?.attributes.snapshots));

	latestSnapshot = $derived(this.snapshots[0]);
	isLatestPending = $derived(this.latestSnapshot?.attributes.status === "pending");

	setExpanded(open: boolean) {
		this.isExpanded = open;
	}
}

const ctx = new Context<AiAgentRunComponentController>("AiAgentRunComponentController");
export const initAiAgentRunComponentController = (idFn: Getter<string>) =>
	ctx.set(new AiAgentRunComponentController(idFn));
export const useAiAgentRunComponentController = () => ctx.get();
