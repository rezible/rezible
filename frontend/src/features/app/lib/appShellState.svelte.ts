import { settings as svelteUxSettings } from "svelte-ux";
import { watch } from "runed";
import type { AvatarProps } from "$components/avatar/Avatar.svelte";

export type PageBreadcrumb = {
	label?: string;
	href?: string;
	avatar?: AvatarProps;
};

const createAppShellState = () => {
	let breadcrumbs = $state<PageBreadcrumb[]>([]);

	const setup = () => {
		svelteUxSettings({
			themes: { light: ["light-old"], dark: ["dark", "bleh"] },
		});
	}

	return {
		setup,
		get breadcrumbs() { return breadcrumbs },
		set	breadcrumbs(value: PageBreadcrumb[]) { breadcrumbs = value },
	};
};
export const appShell = createAppShellState();

export const setPageBreadcrumbs = (source: () => PageBreadcrumb[]) => {
	watch(source, crumbs => {appShell.breadcrumbs = crumbs});
}