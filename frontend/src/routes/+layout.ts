import { browser, dev } from "$app/environment";
import type { LayoutLoad } from './$types';
import { QueryClient } from "@tanstack/svelte-query";

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

export const load: LayoutLoad = async () => {
	return { queryClient };
};