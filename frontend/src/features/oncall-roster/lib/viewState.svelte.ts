import { getOncallRosterOptions, type OncallShift } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

class RosterViewState {
	rosterSlug = $state<string>(null!);
	constructor(slugFn: () => string) {
		this.rosterSlug = slugFn();
		watch(slugFn, id => {this.rosterSlug = id});
	}

	private rosterQuery = createQuery(() => ({
		...getOncallRosterOptions({ path: { id: this.rosterSlug } }),
		enabled: !!this.rosterSlug,
	}));

	roster = $derived(this.rosterQuery.data?.data);
	rosterId = $derived(this.roster?.id);
	rosterName = $derived(this.roster?.attributes.name ?? "");

	activeShift = $derived<OncallShift | undefined>(undefined);
}

const ctx = new Context<RosterViewState>("oncallRosterView");
export const setOncallRosterViewState = (slugFn: Getter<string>) => ctx.set(new RosterViewState(slugFn));
export const useOncallRosterViewState = () => ctx.get();