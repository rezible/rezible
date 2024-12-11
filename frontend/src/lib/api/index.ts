import { dev } from '$app/environment';
import { client } from "./oapi.gen/services.gen";
import type { ErrorModel, ResponsePagination } from "./oapi.gen/types.gen";

import { createConfig, type Options } from '@hey-api/client-fetch';
import type { CreateQueryOptions } from '@tanstack/svelte-query';

client.setConfig(createConfig({
	baseUrl: dev ? '/api' : undefined,
}));

client.interceptors.error.use((err, resp, req, opts) => {
	if (!err) {
		// TODO
		return {title: "", status: resp.status, detail: ""} as ErrorModel;
	}
	return tryUnwrapApiError(err as Error);
});

export const tryUnwrapApiError = (err: Error): ErrorModel => {
	console.log("unwrap", err);
	try {
		if ("detail" in err) return err as ErrorModel;
		const parsed = JSON.parse(err.message) as ErrorModel;
		return parsed;
	} catch {
		return {title: "Server Error", detail: err.message ?? "Unknown Error", status: 503}
	}
}

export const defaultListQueryLimit = 10;

export type ListQueryParameters = {
	limit?: number;
	offset?: number;
	search?: string;
	archived?: boolean;
};
export type ListFuncQueryOptions = Options<{query?: ListQueryParameters}>;

export type PaginatedResponse<T> = {
	readonly $schema?: string;
	data: Array<T>;
	pagination: ResponsePagination;
}
export type ListQueryOptionsFunc<T> = (opts: ListFuncQueryOptions) => 
	CreateQueryOptions<PaginatedResponse<T>, Error, PaginatedResponse<T>, any>;

export { client };

export * from "./oapi.gen/@tanstack/svelte-query.gen";
export * from "./oapi.gen/types.gen";
