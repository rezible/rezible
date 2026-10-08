import { describe, expect, test } from "bun:test";

import { toApiError, unreachableDetail } from "./errors";

const traceId = "4bf92f3577b34da6a3ce929d0e0e4736";

const response = (status: number, headers: Record<string, string> = {}) =>
	new Response(null, { status, headers });

describe("toApiError", () => {
	test("an error body keeps its fields and gains the response's trace ID", () => {
		const body = {
			status: 409,
			code: "conflict",
			title: "Conflict",
			detail: "A team with that name exists.",
		};

		const error = toApiError(body, response(409, { traceparent: `00-${traceId}-00f067aa0ba902b7-01` }));

		expect(error).toEqual({ ...body, code: "conflict", traceId });
	});

	test("no response is a connection problem with no status", () => {
		const error = toApiError(new TypeError("Failed to fetch"));

		expect(error).toEqual({ code: "unavailable", detail: unreachableDetail });
		expect(error.status).toBeUndefined();
	});

	test("a body that isn't an error model is internal, with the response's status", () => {
		const error = toApiError("<html>oops</html>", response(500));

		expect(error.code).toBe("internal");
		expect(error.status).toBe(500);
		expect(error.detail).toBeUndefined();
	});

	test("a gateway answering for an unreachable backend is a connection problem", () => {
		const error = toApiError("", response(502));

		expect(error).toEqual({ status: 502, code: "unavailable", detail: unreachableDetail });
	});
});
