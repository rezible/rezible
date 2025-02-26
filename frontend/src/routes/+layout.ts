import type { LayoutLoad } from "./$types";
import { session } from "$lib/auth.svelte";
import { redirect } from "@sveltejs/kit";

export const ssr = false;
export const prerender = false;
export const csr = true;

export const load: LayoutLoad = async ({ fetch, route }) => {
	let isAuthRoute = !!route.id?.startsWith("/auth");
	const sessionOk = await session.load(fetch);	

	if (!sessionOk && !isAuthRoute) return redirect(301, "/auth");

	return {};
};
