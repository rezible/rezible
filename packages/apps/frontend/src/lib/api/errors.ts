import type { ErrorModel } from "@rezible/api-client-ts";

export type ErrorCode = ErrorModel["code"];

/** Every error the API client returns: the response's error model, or one built for a failure without one. */
export type ApiError = ErrorModel & {
	code: ErrorCode;
	traceId?: string;
};

export const unreachableDetail = "Couldn't reach Rezible";

// A proxy or load balancer answers with these when it can't reach the backend.
const gatewayStatuses = [502, 503, 504];

const isErrorModel = (body: unknown): body is ErrorModel => {
	return !!body && typeof body === "object" && "code" in body && typeof body.code === "string";
};

// traceparent is "<version>-<trace id>-<parent id>-<flags>".
const traceIdFromHeader = (traceparent: string | null) => {
	const traceId = traceparent?.split("-")[1];
	return traceId || undefined;
};

/** Converts what the API client caught, with the response if there was one, to an ApiError. */
export const toApiError = (caught: unknown, response?: Response): ApiError => {
	if (!response) {
		return { code: "unavailable", detail: unreachableDetail };
	}

	const traceId = traceIdFromHeader(response.headers.get("traceparent"));
	if (isErrorModel(caught)) {
		return { ...caught, traceId };
	}
	if (gatewayStatuses.includes(response.status)) {
		return { status: response.status, code: "unavailable", detail: unreachableDetail, traceId };
	}
	return { status: response.status, code: "internal", traceId };
};

declare module "@tanstack/svelte-query" {
	interface Register {
		defaultError: ApiError;
	}
}
