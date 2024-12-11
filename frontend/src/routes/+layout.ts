import type { LayoutLoad } from './$types';
import { AUTH_REDIRECT_URL, session, notifications } from '$lib/auth.svelte';
import { QueryClient } from '@tanstack/svelte-query';
import { browser, dev } from '$app/environment';
import { redirect } from '@sveltejs/kit';

export const ssr = false;
export const prerender = false;
export const csr = true;

export const load: LayoutLoad = async ({ fetch }) => {
	const errStatus = await session.fetchInitial(fetch);
	if (errStatus) {
		if (errStatus === 401) {
			return redirect(301, AUTH_REDIRECT_URL);
		}
		if (errStatus >= 500) {
			// TODO
			console.log("failed to set auth session", errStatus);
		}
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
