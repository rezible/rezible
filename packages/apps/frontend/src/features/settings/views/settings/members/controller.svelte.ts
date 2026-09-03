import { listUsersOptions, type ErrorModel } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useUserSessionState } from "$lib/user-session.svelte";
import { Context } from "runed";

export class MembersSettingsController {
	session = useUserSessionState();
	search = $state("");

	paginatedUsersQuery = createPaginatedQuery({
		queryOptions: (pagination) => ({
			...listUsersOptions({
				query: {
					search: this.search.trim() || undefined,
					...pagination,
				},
			}),
			enabled: this.session.isAdmin,
		}),
		resetWhen: () => this.search,
	});

	private usersQuery = $derived(this.paginatedUsersQuery.query);
	users = $derived(this.usersQuery.data?.data ?? []);
	loading = $derived(this.usersQuery.isPending);
	error = $derived(this.usersQuery.error as ErrorModel | null);
}

const ctx = new Context<MembersSettingsController>("MembersSettingsController");
export const initMembersSettingsController = () => ctx.set(new MembersSettingsController());
