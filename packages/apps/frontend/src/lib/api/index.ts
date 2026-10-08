import { client } from "@rezible/api-client-ts";
import { toApiError } from "./errors";

client.setConfig({
	baseUrl: "/api/v1",
	credentials: "include",
});
client.interceptors.error.use(async (caught, response) => toApiError(caught, response));

export * from "@rezible/api-client-ts";
export * from "@rezible/api-client-ts/svelte-query";
export { client };
export type { ApiError, ErrorCode } from "./errors";
export { errorDisplay, type ErrorDisplay } from "./errorDisplay";
