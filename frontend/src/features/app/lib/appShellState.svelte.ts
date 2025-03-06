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
	const pageActionsComponent = $derived(pageActions?.component);

	const checkPageActions = (newRouteId: string) => {
		if (!pageActions) return;
		const isChild = newRouteId.startsWith(pageActions.routeBase);
		if (!isChild || !pageActions.allowChildren) {pageActions = undefined}
	}

	const setup = () => {
		onNavigate(nav => checkPageActions(nav.to?.route.id ?? ""));
	}

	const setPageActions = (component: Component, allowChildren: boolean) => {
		pageActions = {component, allowChildren, routeBase: $state.snapshot(page.route.id) ?? ""};
	}

	return {
		setup,
		get breadcrumbs() { return breadcrumbs },
		set	breadcrumbs(value: PageBreadcrumb[]) { breadcrumbs = value },
		setPageActions,
		get pageActionsComponent() { return pageActionsComponent },
	};
};
export const appShell = createAppShellState();

export const setPageBreadcrumbs = (source: () => PageBreadcrumb[]) => {
	watch(source, crumbs => {appShell.breadcrumbs = crumbs});
}