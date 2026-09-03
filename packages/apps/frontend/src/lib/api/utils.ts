import type { CreateQueryOptions } from "@tanstack/svelte-query";
import type { Options, Pagination } from "@rezible/api-client-ts";

export type PaginatedQueryParameters = {
	page?: number;
	pageSize?: number;
	search?: string;
	archived?: boolean;
};

export type PaginatedFuncQueryOptions = Options<{
	query?: PaginatedQueryParameters;
	url: string;
}>;

export type ResponsePage<T> = {
	readonly $schema?: string;
	data: T[];
	pagination: Pagination;
};

export type PaginatedQueryOptionsFunc<T> = (
	opts: PaginatedFuncQueryOptions
) => CreateQueryOptions<ResponsePage<T>, Error, ResponsePage<T>, any>;

export const getNextPageParam = (lastPage: unknown) => {
	const { page, pageSize, total } = (lastPage as ResponsePage<unknown>).pagination;
	return page * pageSize < total ? page + 1 : undefined;
};
