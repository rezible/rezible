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

export class AppShellState {
	breadcrumbs = $state<PageBreadcrumb[]>([]);
	pageActions = $state<PageActions<any>>();

	setup() {
		onNavigate(nav => this.checkPageActions(nav.to?.route.id ?? ""));
	}

	private checkPageActions(newRouteId: string) {
		if (!this.pageActions) return;
		const isChild = newRouteId.startsWith(this.pageActions.routeBase);
		if (!isChild || !this.pageActions.allowChildren) {this.pageActions = undefined}
	}

	setPageActions<PComponent extends Component<any>>(component: PComponent, allowChildren: boolean, propsFn?: () => ComponentProps<PComponent>) {
		this.pageActions = {component, allowChildren, propsFn, routeBase: $state.snapshot(page.route.id) ?? ""};
	}

	setPageBreadcrumbs(crumbsFn: () => PageBreadcrumb[]) {
		watch(crumbsFn, crumbs => {this.breadcrumbs = crumbs});
	}
}

export const appShell = new AppShellState();
