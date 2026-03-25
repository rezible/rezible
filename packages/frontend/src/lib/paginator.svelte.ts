import type { CreateQueryResult } from "@tanstack/svelte-query";
import { watch } from "runed";
import { fromStore } from "svelte/store";
import type { ErrorModel, ResponsePagination } from "./api";

type PaginatedData = {
	data: any;
	pagination: ResponsePagination;
}
type PaginatedQuery<PData extends PaginatedData> = CreateQueryResult<PData, ErrorModel>;

export class QueryPaginatorState {
	// pagination = createPaginationStore({ page: 1, perPage: 10, total: 0 });
	// paginationState = fromStore(this.pagination);

	// page = $derived(this.paginationState.current.page as number);
	// limit = $derived(this.paginationState.current.perPage);
	page = $derived(1)
	limit = $derived(10);
	offset = $derived(Math.max(0, (this.page - 1) * this.limit));

	queryParams = $derived({limit: this.limit, offset: this.offset});

	watchQuery(query: PaginatedQuery<PaginatedData>) {
		watch(() => query.data, d => {
			if (!d) return;
			//this.pagination.setTotal(d.pagination.total);
		})
	}

	paginationProps = $derived({
		perPageOptions: [10, 25, 50],
		show: ["perPage", "pagination", "prevPage", "nextPage"],
	});
}
