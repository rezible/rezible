import type { ApiError } from "./errors";

export type ErrorDisplay = {
	title: string;
	detail?: string;
};

const ReferenceLength = 8;

/** The copy people see for an API error, chosen by its code. Only a 4xx shows the API's own message. */
export const errorDisplay = (error: ApiError): ErrorDisplay => {
	switch (error.code) {
		case "invalid_input":
		case "unprocessable":
			return { title: "Check the details", detail: error.detail };
		case "conflict":
			return { title: "Already exists", detail: error.detail };
		case "not_found":
			return { title: "Not found", detail: error.detail };
		case "forbidden":
			return { title: "Not allowed", detail: "Your role can't do this. Ask an admin." };
		case "unauthenticated":
			return { title: "Signed out", detail: "Sign in again to continue." };
		case "rate_limited":
			return { title: "Too many requests", detail: "Wait a moment and try again." };
		case "unavailable":
			return { title: "Connection problem", detail: error.detail };
		case "not_implemented":
			return { title: "Not available yet" };
		default:
			return { title: "Something went wrong", detail: internalReference(error.traceId) };
	}
};

const internalReference = (traceId?: string) => {
	if (!traceId) {
		return undefined;
	}
	return `Reference: ${traceId.slice(0, ReferenceLength)}`;
};
