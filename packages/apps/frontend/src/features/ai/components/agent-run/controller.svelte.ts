import { getAiAgentRunOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

export class AiAgentRunComponentController {
    runId = $state<string>(null!);

    constructor(idFn: Getter<string>) {
        this.runId = idFn();
        watch(idFn, id => {this.runId = id});
    }

    private agentRunQuery = createQuery(() => getAiAgentRunOptions({ path: { id: this.runId } }));
    agentRun = $derived(this.agentRunQuery.data?.data);
}

const ctx = new Context<AiAgentRunComponentController>("AiAgentRunComponentController");
export const initAiAgentRunComponentController = (idFn: Getter<string>) => ctx.set(new AiAgentRunComponentController(idFn));
export const useAiAgentRunComponentController = () => ctx.get();