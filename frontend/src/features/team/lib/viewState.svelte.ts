import { getTeamOptions } from "$lib/api";
import type { Getter } from "$src/lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

class TeamViewState {
	teamSlug = $state<string>(null!);

	constructor(slugFn: Getter<string>) {
		this.teamSlug = slugFn();
		watch(slugFn, id => {this.teamSlug = id});
	}

	private teamQuery = createQuery(() => ({
		...getTeamOptions({ path: { id: this.teamSlug } }),
	}));

	team = $derived(this.teamQuery.data?.data);
	teamId = $derived(this.team?.id);
	teamName = $derived(this.team?.attributes.name ?? "");
}

const teamViewCtx = new Context<TeamViewState>("teamView");
export const setTeamViewState = (slugFn: Getter<string>) => teamViewCtx.set(new TeamViewState(slugFn));
export const useTeamViewState = () => teamViewCtx.get();