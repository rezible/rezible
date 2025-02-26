import { settings as svelteUxSettings } from "svelte-ux";
import { watch } from "runed";
import type { AvatarProps } from "$components/avatar/Avatar.svelte";
import type { Component } from "svelte";
import { page } from "$app/state";
import { onNavigate } from "$app/navigation";

export type PageBreadcrumb = {
	label?: string;
	href?: string;
	avatar?: AvatarProps;
};

export type PageActions = {
	component: Component;
	allowChildren: boolean;
	routeBase: string;
}

const createAppShellState = () => {
	let breadcrumbs = $state<PageBreadcrumb[]>([]);
	let pageActions = $state<PageActions>();

	const setup = () => {
		svelteUxSettings({
			themes: { light: ["light-old"], dark: ["dark", "bleh"] },
		});

		onNavigate(nav => {
			if (!pageActions) return;
			console.log(nav);
			if (!nav.to?.route.id?.startsWith(pageActions.routeBase)) {
				console.log("clear actions");
			}
		})
	}

	const setPageActions = (component: Component, allowChildren: boolean) => {
		pageActions = {component, allowChildren, routeBase: $state.snapshot(page.route.id) ?? ""};
		console.log("set actions", $state.snapshot(pageActions));
	}

	return {
		setup,
		get breadcrumbs() { return breadcrumbs },
		set	breadcrumbs(value: PageBreadcrumb[]) { breadcrumbs = value },
		setPageActions,
		get pageActionsComponent() { return pageActions?.component },
	};
};
export const appShell = createAppShellState();

export const setPageBreadcrumbs = (source: () => PageBreadcrumb[]) => {
	watch(source, crumbs => {appShell.breadcrumbs = crumbs});
}