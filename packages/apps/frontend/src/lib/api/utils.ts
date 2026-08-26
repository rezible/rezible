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

export type ResponsePage<T> = {
	readonly $schema?: string;
	data: T[];
	pagination: ResponsePagination;
};

export type ListQueryOptionsFunc<T> = (opts: ListFuncQueryOptions) => 
	CreateQueryOptions<ResponsePage<T>, Error, ResponsePage<T>, any>;
