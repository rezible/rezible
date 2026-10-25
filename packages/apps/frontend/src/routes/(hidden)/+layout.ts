import { redirect } from "@sveltejs/kit";
import { resolve } from "$app/paths";

export const load = (): void => {
	redirect(307, resolve("/"));
};
