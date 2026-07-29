import { listUsersOptions, type ErrorModel } from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export class MembersSettingsController {
	session = useUserSessionState();
	search = $state("");

	private usersQueryOptions = $derived(
		listUsersOptions({
			query: {
				search: this.search.trim() || undefined,
				limit: 100,
			},
		})
	);
	private usersQuery = createQuery(() => ({
		...this.usersQueryOptions,
		enabled: this.session.isAdmin,
	}));

	users = $derived(this.usersQuery.data?.data ?? []);
	loading = $derived(this.usersQuery.isPending);
	error = $derived(this.usersQuery.error as ErrorModel | null);
}

const ctx = new Context<MembersSettingsController>("MembersSettingsController");
export const initMembersSettingsController = () => ctx.set(new MembersSettingsController());
