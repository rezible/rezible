import { describe, expect, spyOn, test } from "bun:test";
import { hashKey } from "@tanstack/svelte-query";

import { getUserSessionOptions, type ApiError } from "$lib/api";

import { createQueryClient } from "./query-client";

const unauthenticated: ApiError = { status: 401, code: "unauthenticated", detail: "Sign in to continue." };

const setup = () => {
	const queryClient = createQueryClient({ defaultOptions: { queries: { retry: false } } });
	const invalidate = spyOn(queryClient, "invalidateQueries");
	const sessionQueryKey = getUserSessionOptions().queryKey;
	const sessionQueryHash = hashKey(sessionQueryKey);
	const sessionInvalidations = () =>
		invalidate.mock.calls.filter(
			([filters]) => !!filters?.queryKey && hashKey(filters.queryKey) === sessionQueryHash
		).length;
	return { queryClient, sessionQueryKey, sessionInvalidations };
};

const failWith = (error: ApiError) => async () => {
	throw error;
};

describe("createQueryClient", () => {
	test("401s from other queries and mutations failing together invalidate the session query once", async () => {
		const { queryClient, sessionInvalidations } = setup();

		const failures = [
			queryClient.fetchQuery({ queryKey: ["situations"], queryFn: failWith(unauthenticated) }),
			queryClient.fetchQuery({ queryKey: ["incidents"], queryFn: failWith(unauthenticated) }),
			queryClient
				.getMutationCache()
				.build(queryClient, { mutationFn: failWith(unauthenticated) })
				.execute(undefined),
		];
		await Promise.allSettled(failures);

		expect(sessionInvalidations()).toBe(1);
	});

	test("the session query's own 401 does not invalidate it", async () => {
		const { queryClient, sessionQueryKey, sessionInvalidations } = setup();

		await queryClient
			.fetchQuery({ queryKey: sessionQueryKey, queryFn: failWith(unauthenticated) })
			.catch(() => {});

		expect(sessionInvalidations()).toBe(0);
	});

	test("other errors do not invalidate the session query", async () => {
		const { queryClient, sessionInvalidations } = setup();

		await queryClient
			.fetchQuery({ queryKey: ["situations"], queryFn: failWith({ status: 403, code: "forbidden" }) })
			.catch(() => {});

		expect(sessionInvalidations()).toBe(0);
	});
});
