import type { LayoutLoad } from './$types';
import { session, notifications } from '$lib/auth.svelte';
import { QueryClient } from '@tanstack/svelte-query';
import { browser, dev } from '$app/environment';
import { redirect } from '@sveltejs/kit';

export const ssr = false;
export const prerender = false;
export const csr = true;

export const load: LayoutLoad = async ({ fetch }) => {
	const authRedirect = await session.load(fetch);
	if (authRedirect) {
		return redirect(301, authRedirect);
	}
	
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser,
				retry: false,
				refetchOnWindowFocus: dev,
				staleTime: 5000
			}
		}
	});
	notifications.setQueryClient(queryClient);
	
	return {queryClient};
};
