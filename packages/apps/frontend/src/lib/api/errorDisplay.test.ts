import { describe, expect, test } from "bun:test";

import { errorDisplay } from "./errorDisplay";
import type { ApiError, ErrorCode } from "./errors";

const apiDetail = 'service_labels contains "job" more than once';

const error = (code: ErrorCode, extra: Partial<ApiError> = {}): ApiError => ({
	code,
	detail: apiDetail,
	...extra,
});

describe("errorDisplay", () => {
	test("4xx codes meant for people show the API's message", () => {
		expect(errorDisplay(error("invalid_input"))).toEqual({
			title: "Check the details",
			detail: apiDetail,
		});
		expect(errorDisplay(error("unprocessable"))).toEqual({
			title: "Check the details",
			detail: apiDetail,
		});
		expect(errorDisplay(error("conflict"))).toEqual({ title: "Already exists", detail: apiDetail });
		expect(errorDisplay(error("not_found"))).toEqual({ title: "Not found", detail: apiDetail });
	});

	test("other codes show their own copy", () => {
		expect(errorDisplay(error("forbidden"))).toEqual({
			title: "Not allowed",
			detail: "Your role can't do this. Ask an admin.",
		});
		expect(errorDisplay(error("unauthenticated"))).toEqual({
			title: "Signed out",
			detail: "Sign in again to continue.",
		});
		expect(errorDisplay(error("rate_limited"))).toEqual({
			title: "Too many requests",
			detail: "Wait a moment and try again.",
		});
		expect(errorDisplay(error("not_implemented"))).toEqual({ title: "Not available yet" });
	});

	test("a connection problem shows the message it came with", () => {
		const display = errorDisplay(error("unavailable", { detail: "Couldn't reach Rezible" }));

		expect(display).toEqual({ title: "Connection problem", detail: "Couldn't reach Rezible" });
	});

	test("an internal error shows only a short reference, never the API's message", () => {
		const traceId = "4bf92f3577b34da6a3ce929d0e0e4736";

		expect(errorDisplay(error("internal", { traceId }))).toEqual({
			title: "Something went wrong",
			detail: "Reference: 4bf92f35",
		});
		expect(errorDisplay(error("internal"))).toEqual({ title: "Something went wrong" });
	});
});
