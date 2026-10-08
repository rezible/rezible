import { errorDisplay, type ApiError, type ErrorDisplay } from "$lib/api";

const sessionEndedNotice: ErrorDisplay = { title: "Your session ended. Sign in again." };

/** What the login page says about the session query's error. Someone never signed in here sees nothing. */
export const sessionNotice = (error: ApiError | undefined, hadSession: boolean): ErrorDisplay | undefined => {
	if (!error) {
		return undefined;
	}
	if (error.code !== "unauthenticated") {
		return errorDisplay(error);
	}
	if (hadSession) {
		return sessionEndedNotice;
	}
	return undefined;
};
