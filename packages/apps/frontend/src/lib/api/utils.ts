import type { CreateQueryOptions } from "@tanstack/svelte-query";
import type { Options, ResponsePagination } from "@rezible/api-client-ts";

export type ListQueryParameters = {
	limit?: number;
	offset?: number;
	search?: string;
	archived?: boolean;
};
export type ListFuncQueryOptions = Options<{
	query?: ListQueryParameters;
	url: string;
}>;

export type PaginatedResponse<T> = {
	readonly $schema?: string;
	data: Array<T>;
	pagination: ResponsePagination;
};
export type ListQueryOptionsFunc<T> = (o: ListFuncQueryOptions) 
	=> CreateQueryOptions<PaginatedResponse<T>, Error, PaginatedResponse<T>, any>;
