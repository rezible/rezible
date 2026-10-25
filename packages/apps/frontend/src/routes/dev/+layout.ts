import { dev } from "$app/environment";
import { error } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";

export const load = (() => {
	if (!dev) {
		error(404);
	}
}) satisfies LayoutLoad;
