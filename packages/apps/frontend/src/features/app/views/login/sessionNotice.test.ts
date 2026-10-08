import { describe, expect, test } from "bun:test";

import type { ApiError } from "$lib/api";

import { sessionNotice } from "./sessionNotice";

const unauthenticated: ApiError = { status: 401, code: "unauthenticated", detail: "Sign in to continue." };

describe("sessionNotice", () => {
	test("someone never signed in in this tab sees nothing", () => {
		expect(sessionNotice(unauthenticated, false)).toBeUndefined();
	});

	test("someone whose session ended is told to sign in again", () => {
		expect(sessionNotice(unauthenticated, true)).toEqual({ title: "Your session ended. Sign in again." });
	});

	test("any other error shows its mapped copy", () => {
		const unreachable: ApiError = { code: "unavailable", detail: "Couldn't reach Rezible" };

		expect(sessionNotice(unreachable, true)).toEqual({
			title: "Connection problem",
			detail: "Couldn't reach Rezible",
		});
	});

	test("no error, no notice", () => {
		expect(sessionNotice(undefined, true)).toBeUndefined();
	});
});
