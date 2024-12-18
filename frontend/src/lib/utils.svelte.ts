import { createQuery, Query, QueryClient, QueryObserver, useQueryClient, type DefaultError, type DefinedInitialDataOptions, type FetchStatus, type FunctionedParams, type QueryKey, type QueryObserverResult, type QueryOptions, type UndefinedInitialDataOptions } from "@tanstack/svelte-query";
import { tryUnwrapApiError, type ErrorDetail, type ErrorModel } from "./api";

export function debounce<T extends Function>(cb: T, wait = 100) {
	let timeout: ReturnType<typeof setTimeout>;
	let callable = (...args: any) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => cb(...args), wait);
	};
	return <T>(<any>callable);
}

export const onQueryUpdate = <D, E extends Error, K extends QueryKey>(
	optsFn: FunctionedParams<UndefinedInitialDataOptions<any, E, D, K>>,
	onData: (data: D) => void,
	onError?: (error: ErrorModel) => void
) => {
	let lastStatus = $state<FetchStatus>();
	const onChange = ((res: QueryObserverResult<D, Error>) => {
		if (res.fetchStatus === lastStatus) return;
		lastStatus = res.fetchStatus;
		if (res.isError && onError) {
			const queryErr = tryUnwrapApiError(new Error("foo"));
			onError(queryErr);
		}
		if (res.isSuccess) onData(res.data);
	});

	const queryKey = $derived(optsFn().queryKey);
	const client = useQueryClient();
	$effect(() => {
		if (!queryKey) return;
		const observer = new QueryObserver<any, Error, D, K>(client, {queryKey});
		observer.subscribe(onChange);
		return () => {
			observer.destroy();
			lastStatus = undefined;
		};
	});
}