import { z } from "zod";
import {
	createQuery,
	keepPreviousData,
	type CreateQueryOptions,
	type CreateQueryResult,
	type QueryKey,
} from "@tanstack/svelte-query";
import { watch, type Getter } from "runed";
import { useSearchParams, type ReturnUseSearchParams } from "runed/kit";
import type { ErrorModel, Pagination as ResponsePagination } from "$lib/api";

export const pageSizeOptions = [10, 25, 50] as const;
export type PageSize = (typeof pageSizeOptions)[number];

const defaultPage = 1;
const defaultPageSize: PageSize = 25;

const paginationParamsSchema = z.object({
	page: z.number().default(defaultPage),
	pageSize: z.number().default(defaultPageSize),
});

export class QueryPaginator {
	private params: ReturnUseSearchParams<typeof paginationParamsSchema>;

	constructor(source: "local" | "url" = "local", resetParams?: Getter<any>) {
		this.params = useSearchParams(paginationParamsSchema, {
			updateURL: source === "url",
		});
		if (resetParams) {
			watch(resetParams, () => {this.resetPage()}, { lazy: true });
		}
	}

	get page() {
		return this.params.page;
	}

	get pageSize() {
		return this.params.pageSize;
	}

	get queryParams() {
		return { page: this.page, pageSize: this.pageSize };
	}

	setPage(page: number) {
		this.params.page = page;
	};

	setPageSize(size: number) {
		this.params.update({
			pageSize: size,
			page: defaultPage,
		});
	};

	resetPage() {
		this.setPage(defaultPage);
	};

	reconcile(pagination?: ResponsePagination, placeholder = false) {
		if (!pagination || placeholder) return;
        const lastPage = Math.max(defaultPage, Math.ceil(pagination.total / pagination.pageSize));
        if (this.page > lastPage) this.params.page = lastPage;
	};
}

export type PaginatedQueryResult = {
    pagination: ResponsePagination;
};
export type PaginatedQuery<R extends PaginatedQueryResult = PaginatedQueryResult> = CreateQueryResult<R, ErrorModel>;

export const createPaginatedQuery = <
	TQueryData extends PaginatedQueryResult,
	TData extends PaginatedQueryResult = TQueryData,
	TQueryKey extends QueryKey = QueryKey,
>({
	source = "url", 
	resetWhen, 
	queryOptions,
	keepPreviousQueryData = true,
}: {
	source?: "local" | "url";
	resetWhen?: Getter<unknown>;
	queryOptions: (pagination: z.infer<typeof paginationParamsSchema>) => CreateQueryOptions<TQueryData, ErrorModel, TData, TQueryKey>;
	keepPreviousQueryData?: boolean;
}) => {
	const paginator = new QueryPaginator(source, resetWhen);
	const placeholderData = keepPreviousQueryData ? keepPreviousData : undefined;
	const query = createQuery(() => ({
		...queryOptions(paginator.queryParams),
		placeholderData,
	}));

	watch(
		() => [query.data?.pagination, query.isPlaceholderData] as const,
		([pagination, placeholder]) => paginator.reconcile(pagination, placeholder),
	);

	return { paginator, query };
}

export const createPaginatedQuerySimple = <
	TQueryData extends PaginatedQueryResult,
	TData extends PaginatedQueryResult = TQueryData,
	TQueryKey extends QueryKey = QueryKey,
	TQueryParams extends object = Record<string, never>,
>({
	source = "url",
	queryParams,
	optsFn,
	keepPreviousQueryData = true,
}: {
	source?: "local" | "url";
	queryParams: Getter<TQueryParams>;
	optsFn: (
		query: TQueryParams & z.infer<typeof paginationParamsSchema>
	) => CreateQueryOptions<TQueryData, ErrorModel, TData, TQueryKey>;
	keepPreviousQueryData?: boolean;
}) => {
	const paginator = new QueryPaginator(source);
	const placeholderData = keepPreviousQueryData ? keepPreviousData : undefined;

	const query = createQuery(() => ({
		...optsFn({...paginator.queryParams, ...queryParams()}),
		placeholderData,
	}));

	watch(
		() => JSON.stringify(queryParams()),
		() => paginator.resetPage(), { lazy: true },
	);
	watch(
		() => [query.data?.pagination, query.isPlaceholderData] as const,
		([pagination, placeholder]) => paginator.reconcile(pagination, placeholder),
	);

	return { paginator, query };
};
