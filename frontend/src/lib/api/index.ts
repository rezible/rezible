import { dev } from "$app/environment";
import { createConfig, type ClientOptions } from "./oapi.gen/client";
import { client } from "./oapi.gen/client.gen";
import type { ErrorModel, ResponsePagination } from "./oapi.gen/types.gen";

import { type Options } from "@hey-api/client-fetch";
import type { CreateQueryOptions } from "@tanstack/svelte-query";

import { PUBLIC_REZ_API_BASE_URL } from '$env/static/public';

export const API_BASE_URL = PUBLIC_REZ_API_BASE_URL;
export const BACKEND_URL = dev ? "https://app.dev.rezible.com" : "";

const clientConfig = createConfig<ClientOptions>({ 
	baseUrl: API_BASE_URL,
	// credentials: "include",
});
client.setConfig(clientConfig);

client.interceptors.error.use(async (err, resp, req, opts) => {
	const status = resp.status;
	if (!!err) return unwrapApiError(err as Error, status);
	return { title: "Unknown Error", status, detail: "" } as ErrorModel;
});

const unwrapApiError = (err: Error, status = 503): ErrorModel => {
	try {
		if ("detail" in err) return err as ErrorModel;
		return JSON.parse(err.message) as ErrorModel;
	} catch {
		return {
			title: "Error",
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

export const simulateApiDelay = async (ms: number) => (await new Promise(res => setTimeout(res, ms)));

export * from "./oapi.gen/@tanstack/svelte-query.gen";
export * from "./oapi.gen/types.gen";
