import { watch } from "runed";
import type { AvatarProps } from "$components/avatar/Avatar.svelte";
import type { Component, ComponentProps } from "svelte";
import { page } from "$app/state";
import { onNavigate } from "$app/navigation";

export type PageBreadcrumb = {
	label?: string;
	href?: string;
	avatar?: AvatarProps;
};

export type PageActions<PComponent extends Component<any>> = {
	component: Component;
	propsFn?: () => ComponentProps<PComponent>;
	allowChildren: boolean;
	routeBase: string;
}

const createAppShellState = () => {
	let breadcrumbs = $state<PageBreadcrumb[]>([]);
	let pageActions = $state<PageActions<any>>();

	const checkPageActions = (newRouteId: string) => {
		if (!pageActions) return;
		const isChild = newRouteId.startsWith(pageActions.routeBase);
		if (!isChild || !pageActions.allowChildren) {pageActions = undefined}
	}

	const setup = () => {
		onNavigate(nav => checkPageActions(nav.to?.route.id ?? ""));
	}

	const setPageActions = <PComponent extends Component<any>>(component: PComponent, allowChildren: boolean, propsFn?: () => ComponentProps<PComponent>) => {
		pageActions = {component, allowChildren, propsFn, routeBase: $state.snapshot(page.route.id) ?? ""};
	}

	const setPageBreadcrumbs = (crumbsFn: () => PageBreadcrumb[]) => {
		watch(crumbsFn, crumbs => {breadcrumbs = crumbs});
	}

	return {
		setup,
		get breadcrumbs() { return breadcrumbs },
		set	breadcrumbs(value: PageBreadcrumb[]) { breadcrumbs = value },
		setPageBreadcrumbs,
		setPageActions,
		get pageActions() { return pageActions },
	};
};
export const appShell = createAppShellState();
