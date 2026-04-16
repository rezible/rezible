import { ApiError } from "./lib/api";

declare global {
	interface IdProp {id: string};

	namespace App {
		interface Error {
			status: number;
			errorId: string;
			apiError?: ApiError;
		}
	}
}

export {};
