import { createQuery } from "@tanstack/svelte-query";
import { Debounced } from "runed";

import { listUsersOptions, type User } from "$lib/api";

const PICKER_PAGE_SIZE = 20;
const SEARCH_DEBOUNCE_MS = 250;

/** Members of the organization matching a search. */
export class MemberPickerController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, SEARCH_DEBOUNCE_MS);

	private query = createQuery(() =>
		listUsersOptions({
			query: {
				page: 1,
				pageSize: PICKER_PAGE_SIZE,
				search: this.debouncedSearch.current.trim() || undefined,
			},
		})
	);

	isLoading = $derived(this.query.isPending && !this.query.data);
	isError = $derived(this.query.isError && !this.query.data);
	users = $derived<User[]>(this.query.data?.data ?? []);

	setSearch = (value: string) => {
		this.search = value;
	};
}
