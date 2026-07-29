import { listAiAgentsOptions, type ErrorModel } from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export class AgentSettingsController {
	session = useUserSessionState();

	private agentsQuery = createQuery(() => ({
		...listAiAgentsOptions(),
		enabled: this.session.isAdmin,
	}));

	agents = $derived(this.agentsQuery.data?.data);
	loading = $derived(this.agentsQuery.isPending);
	error = $derived(this.agentsQuery.error as ErrorModel | null);
}

const ctx = new Context<AgentSettingsController>("AgentSettingsController");
export const initAgentSettingsController = () => ctx.set(new AgentSettingsController());
