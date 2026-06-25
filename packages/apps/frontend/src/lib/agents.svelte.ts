import type { AgentRun } from "@rezible/api-client-ts";

export const isAgentRunActive = (run: AgentRun) => {
    return !!run.attributes.startedAt && !run.attributes.result
}

export const getAgentRunStatus = (run?: AgentRun) => {
    if (!run) return;
    if (!run.attributes.startedAt) return "queued";
    if (!run.attributes.result) return "running";
    if (!!run.attributes.result.attributes.errorMessage) return "failed";
    return "success";
}