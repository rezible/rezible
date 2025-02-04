import { dev } from "$app/environment";
import { client } from "./oapi.gen";
import type { ErrorModel, ResponsePagination } from "./oapi.gen/types.gen";

import { createConfig, type Options } from "@hey-api/client-fetch";
import type { CreateQueryOptions } from "@tanstack/svelte-query";

const clientConfig = createConfig({ baseUrl: dev ? "/api" : undefined });
client.setConfig(clientConfig);

client.interceptors.error.use(async (err, resp, req, opts) => {
	const status = resp.status;
	if (!!err) return tryUnwrapApiError(err as Error, status);
	return { title: "Unknown Error", status, detail: "" } as ErrorModel;
});

export const tryUnwrapApiError = (err: Error, status = 503): ErrorModel => {
	try {
		if ("detail" in err) return err as ErrorModel;
		return JSON.parse(err.message) as ErrorModel;
	} catch {
		return {
			title: "Server Error",
			detail: err.message ?? "Unknown Error",
			status,
		};
	}
};

export const defaultListQueryLimit = 10;

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
export type ListQueryOptionsFunc<T> = (
	opts: ListFuncQueryOptions
) => CreateQueryOptions<PaginatedResponse<T>, Error, PaginatedResponse<T>, any>;

export { client };

export * from "./oapi.gen/@tanstack/svelte-query.gen";
export * from "./oapi.gen/types.gen";
