import { goto } from "$app/navigation";
import { page } from "$app/state";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const clearQueryParams = async () => {
	const empty = new URL(page.url);
	empty.search = "";
	console.log("clear?", empty);
	await goto(empty, { replaceState: true, noScroll: true });
}

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
