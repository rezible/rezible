import { getOncallRosterOptions } from "$src/lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class RosterViewState {
	rosterId = $state<string>("");
	constructor(idFn: () => string) {
		watch(idFn, id => {this.rosterId = id});
	}

	private rosterQuery = createQuery(() => ({
		...getOncallRosterOptions({ path: { id: this.rosterId } }),
		enabled: !!this.rosterId,
	}));

	roster = $derived(this.rosterQuery.data?.data);
	rosterName = $derived(this.roster?.attributes.name ?? "");
}

export const rosterViewCtx = new Context<RosterViewState>("rosterView");