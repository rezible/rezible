import { settings as svelteUxSettings } from "svelte-ux";
import { getToastState, setupToastState, ToastState, type Toast } from "$components/toaster/toasts.svelte";
import type { AvatarProps } from "$components/avatar/Avatar.svelte";
import { onMount } from "svelte";
import { watch } from "runed";

export type PageBreadcrumb = {
	label?: string;
	href?: string;
	avatar?: AvatarProps;
};

const createAppState = () => {
	let breadcrumbs = $state<PageBreadcrumb[]>([]);
	let toasts = $state<ToastState>();

	const setup = () => {
		svelteUxSettings({
			themes: { light: ["light-old"], dark: ["dark", "bleh"] },
		});

		toasts = setupToastState();
	}

	return {
		setup,
		get toasts() { return toasts },
		get breadcrumbs() { return breadcrumbs },
		set	breadcrumbs(value: PageBreadcrumb[]) { breadcrumbs = value },
	};
};
export const appState = createAppState();

export const setPageBreadcrumbs = (source: () => PageBreadcrumb[]) => {
	watch(source, crumbs => {appState.breadcrumbs = crumbs});
}