import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const normalized = (value: string): string => value.trim().toLowerCase();

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };

/** The URL when it is an absolute http(s) URL, otherwise undefined. Stored references are not always URLs. */
export function safeExternalUrl(value: string | undefined): string | undefined {
	if (!value) {
		return undefined;
	}
	try {
		const url = new URL(value);
		if (url.protocol === "https:" || url.protocol === "http:") {
			return url.href;
		}
	} catch {
		return undefined;
	}
	return undefined;
}
