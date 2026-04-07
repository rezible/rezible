import { API_PATH_BASE } from "$lib/config";
import { client, type ErrorModel } from "@rezible/api-client-ts";

client.setConfig({
	baseUrl: API_PATH_BASE +"/v1",
	credentials: "include",
});
client.interceptors.error.use(async (rawErr, resp, req, opts) => {
	const status = resp.status;
	if (!rawErr) return { title: "Unknown Error", status, detail: "" } as ErrorModel;
	const err = rawErr as Error;
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
});

export * from "@rezible/api-client-ts";
export * from "@rezible/api-client-ts/svelte-query";
export { client };