import {
	hashKey,
	MutationCache,
	QueryCache,
	QueryClient,
	type QueryClientConfig,
	type QueryKey,
} from "@tanstack/svelte-query";
import { getUserSessionOptions, type ApiError } from "$lib/api";

const SessionInvalidationWindowMs = 1000;

/**
 * Creates the app's query client. A 401 from any query or mutation other than the session query
 * invalidates the session query, so the session state sees the session has ended and sends people to
 * sign in. Requests failing together cause one invalidation.
 */
export const createQueryClient = (config: Omit<QueryClientConfig, "queryCache" | "mutationCache">) => {
	const sessionQueryKey = getUserSessionOptions().queryKey;
	const sessionQueryHash = hashKey(sessionQueryKey);
	let lastInvalidatedAt: number | undefined;

	const invalidateSessionOn401 = (error: ApiError, queryKey?: QueryKey) => {
		if (error.status !== 401) {
			return;
		}
		if (!!queryKey && hashKey(queryKey) === sessionQueryHash) {
			return;
		}

		const now = Date.now();
		if (lastInvalidatedAt !== undefined && now - lastInvalidatedAt < SessionInvalidationWindowMs) {
			return;
		}
		lastInvalidatedAt = now;

		void queryClient.invalidateQueries({ queryKey: sessionQueryKey, exact: true });
	};

	const queryClient = new QueryClient({
		...config,
		queryCache: new QueryCache({
			onError: (error, query) => invalidateSessionOn401(error, query.queryKey),
		}),
		mutationCache: new MutationCache({
			onError: (error) => invalidateSessionOn401(error),
		}),
	});
	return queryClient;
};
