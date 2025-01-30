import { ApiError } from "./lib/api";

declare global {
  namespace App {
    interface Error {
      status: number;
      errorId: string;
      apiError?: ApiError;
    }
  }
}

export {};
