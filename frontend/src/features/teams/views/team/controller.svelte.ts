import { getTeamOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

class TeamViewController {
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

const ctx = new Context<TeamViewController>("TeamViewController");
export const initTeamViewController = (slugFn: Getter<string>) => ctx.set(new TeamViewController(slugFn));
export const useTeamViewController = () => ctx.get();