import { getTeamOptions } from "$src/lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class TeamViewState {
	teamId = $state<string>(null!);
	constructor(idFn: () => string) {
		this.teamId = idFn();
		watch(idFn, id => {this.teamId = id});
	}

	private teamQuery = createQuery(() => ({
		...getTeamOptions({ path: { id: this.teamId } }),
	}));

	team = $derived(this.teamQuery.data?.data);
	teamName = $derived(this.team?.attributes.name ?? "");
}

const teamViewCtx = new Context<TeamViewState>("teamView");
export const setTeamViewState = (s: TeamViewState) => teamViewCtx.set(s);
export const useTeamViewState = () => teamViewCtx.get();