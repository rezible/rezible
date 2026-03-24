import { API_URL } from "$lib/config";
import { client } from "./oapi.gen/client.gen";
import type { ErrorModel } from "./oapi.gen";

client.setConfig({
	baseUrl: API_URL,
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

export { client };

export * from "./oapi.gen/@tanstack/svelte-query.gen";
export * from "./oapi.gen/types.gen";
