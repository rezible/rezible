import {
	QueryObserver,
	useQueryClient,
	type FetchStatus,
	type FunctionedParams,
	type QueryKey,
	type QueryObserverResult,
	type UndefinedInitialDataOptions,
} from "@tanstack/svelte-query";
import { tryUnwrapApiError, type ErrorModel } from "./api";
// import { parseAbsolute, parseAbsoluteToLocal } from "@internationalized/date";
import { z } from "zod";

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
	const onChange = (res: QueryObserverResult<D, Error>) => {
		if (res.fetchStatus === lastStatus) return;
		lastStatus = res.fetchStatus;
		if (res.isError && onError) {
			const queryErr = tryUnwrapApiError(new Error("foo"));
			onError(queryErr);
		}
		if (res.isSuccess) onData(res.data);
	};

	const queryKey = $derived(optsFn().queryKey);
	const client = useQueryClient();
	$effect(() => {
		if (!queryKey) return;
		const observer = new QueryObserver<any, Error, D, K>(client, {
			queryKey,
		});
		observer.subscribe(onChange);
		return () => {
			observer.destroy();
			lastStatus = undefined;
		};
	});
};

const refineZonedDateTimeString = (dateStr: string) => {
	try {
		const [datePart, timezonePart] = dateStr.split("[");
		if (!datePart || !timezonePart?.endsWith("]")) return false;

		const isoDateTimeRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$/;
		return isoDateTimeRegex.test(datePart) &&
			/^[A-Za-z_/]+$/.test(timezonePart.slice(0, -1));
	} catch {
		return false;
	}
}

export const ZodZonedDateTime = z.string().
	refine(refineZonedDateTimeString, "Invalid ZonedDateTime format");

export const DayHours = [
	'12am', '1am', '2am', '3am', '4am', '5am', '6am',
	'7am', '8am', '9am', '10am', '11am',
	'12pm', '1pm', '2pm', '3pm', '4pm', '5pm',
	'6pm', '7pm', '8pm', '9pm', '10pm', '11pm'
];