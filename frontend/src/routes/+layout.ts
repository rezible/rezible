import type { LayoutLoad } from "./$types";
import { session } from "$lib/auth.svelte";
import { redirect } from "@sveltejs/kit";
import { QueryClient } from "@tanstack/svelte-query";
import { browser, dev } from "$app/environment";

export const ssr = false;
export const prerender = false;
export const csr = true;

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			enabled: browser,
			retry: false,
			refetchOnWindowFocus: dev,
			staleTime: 5000,
		},
	},
});

export const load: LayoutLoad = async ({ fetch, route }) => {
	let isAuthRoute = !!route.id?.startsWith("/auth");
	const sessionOk = await session.load(fetch);

	if (!sessionOk && !isAuthRoute) return redirect(301, "/auth");

	return {queryClient};
};
