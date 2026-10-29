import { listUsersOptions, type ApiError } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useUserSessionState } from "$lib/user-session.svelte";
import { Context, Debounced } from "runed";

export class MembersSettingsController {
	session = useUserSessionState();
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, 300);
	private searchQuery = $derived(this.debouncedSearch.current.trim() || undefined);

	paginatedUsersQuery = createPaginatedQuery({
		queryOptions: (pagination) => ({
			...listUsersOptions({
				query: {
					search: this.searchQuery,
					...pagination,
				},
			}),
			enabled: this.session.isAdmin,
		}),
		resetWhen: () => this.searchQuery,
	});

	private usersQuery = $derived(this.paginatedUsersQuery.query);
	users = $derived(this.usersQuery.data?.data ?? []);
	loading = $derived(this.usersQuery.isPending);
	error = $derived(this.usersQuery.error as ApiError | null);

	setSearch = (value: string) => {
		this.search = value;
	};
}

const ctx = new Context<MembersSettingsController>("MembersSettingsController");
export const initMembersSettingsController = () => ctx.set(new MembersSettingsController());
