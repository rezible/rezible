import { Context, watch, type Getter } from "runed";
import type Avatar from "$components/avatar/Avatar.svelte";
import type { Component, ComponentProps } from "svelte";
import { page } from "$app/state";
import { onNavigate } from "$app/navigation";

export type PageBreadcrumb = {
	label?: string;
	href?: string;
	avatar?: ComponentProps<typeof Avatar>;
};

export type PageActions<PComponent extends Component<any>> = {
	component: Component;
	propsFn?: () => ComponentProps<PComponent>;
	allowChildren: boolean;
	routeBase: string;
}

export class AppShellController {
	pageTitle = $state("Rezible")
	breadcrumbs = $state<PageBreadcrumb[]>([]);
	pageActions = $state<PageActions<any>>();

	constructor() {
		onNavigate(nav => {
			this.checkPageActions(nav.to?.route.id ?? "")
		});
	}

	private checkPageActions(newRouteId: string) {
		if (!this.pageActions) return;
		const isChild = newRouteId.startsWith(this.pageActions.routeBase);
		if (!isChild || !this.pageActions.allowChildren) {this.pageActions = undefined}
	}

	setPageActions<PComponent extends Component<any>>(component: PComponent, allowChildren: boolean, propsFn?: () => ComponentProps<PComponent>) {
		this.pageActions = {component, allowChildren, propsFn, routeBase: $state.snapshot(page.route.id) ?? ""};
	}

	setPageBreadcrumbs(crumbsFn: Getter<PageBreadcrumb[]>) {
		watch(crumbsFn, crumbs => {this.breadcrumbs = crumbs});
	}
}

const ctx = new Context<AppShellController>("AppShellController");
export const initAppShell = () => ctx.set(new AppShellController());
export const useAppShell = () => ctx.get();

export const setPageBreadcrumbs = (crumbsFn: Getter<PageBreadcrumb[]>) => {
	useAppShell().setPageBreadcrumbs(crumbsFn);
}
