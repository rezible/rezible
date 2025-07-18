import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { watch } from "runed";
import { fromStore } from "svelte/store";
import type { ResponsePagination } from "./api";
import type { ComponentProps } from "svelte";
import { Pagination } from "svelte-ux";

type PaginatedData = {
	data: any;
	pagination: ResponsePagination;
}
type PaginatedQuery<PData extends PaginatedData> = CreateQueryResult<PData, Error>;

export class QueryPaginatorState {
	pagination = createPaginationStore({ page: 1, perPage: 10, total: 0 });
	private pagState = $derived(fromStore(this.pagination));

	page = $derived(this.pagState.current.page as number);
	limit = $derived(this.pagState.current.perPage);
	offset = $derived(Math.max(0, (this.page - 1) * this.limit));

	watchQuery(query: PaginatedQuery<PaginatedData>) {
		watch(() => query.data, d => {
			if (!d) return;
			this.pagination.setTotal(d.pagination.total);
		})
	}

	paginationProps = $derived<ComponentProps<Pagination>>({
		pagination: this.pagination,
		perPageOptions: [10, 25, 50],
		show: ["perPage", "pagination", "prevPage", "nextPage"],
	})
}
