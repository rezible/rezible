import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { listAgentSessionsOptions } from "$lib/api";

const limit = 25;
export class SessionsListController {
	page = $state(0);
	query = createQuery(() => listAgentSessionsOptions({ query: { limit, offset: this.page * limit } }));
	sessions = $derived(this.query.data?.data ?? []);
	total = $derived(this.query.data?.pagination.total ?? 0);
	start = $derived(this.total ? this.page * limit + 1 : 0);
	end = $derived(Math.min((this.page + 1) * limit, this.total));
	previous = () => {
		if (this.page > 0) this.page--;
	};
	next = () => {
		if (this.end < this.total) this.page++;
	};
}
const context = new Context<SessionsListController>("SessionsListController");
export const initSessionsListController = () => context.set(new SessionsListController());
