import { getOncallRosterOptions, type OncallShift } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

export class OncallRosterViewController {
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

const ctx = new Context<OncallRosterViewController>("OncallRosterViewController");
export const initOncallRosterViewController = (slugFn: Getter<string>) => ctx.set(new OncallRosterViewController(slugFn));
export const useOncallRosterViewController = () => ctx.get();